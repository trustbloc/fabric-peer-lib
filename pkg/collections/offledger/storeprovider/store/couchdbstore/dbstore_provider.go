/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package couchdbstore

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/hyperledger/fabric/common/metrics/disabled"
	coreconfig "github.com/hyperledger/fabric/core/config"
	"github.com/hyperledger/fabric/core/ledger"
	couchdb "github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb/statecouchdb"
	viper "github.com/spf13/viper2015"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/storeprovider/store/api"
	"github.com/trustbloc/fabric-peer-ext/pkg/config"
)

const (
	idField            = "_id"
	revField           = "_rev"
	deletedField       = "_deleted"
	txnIDField         = "~txnID"
	expiryField        = "~expiry"
	binaryWrapperField = "valueBytes"
	expiryIndexName    = "by_expiry"
	expiryIndexDoc     = "indexExpiry"

	expiryIndexDef = `
	{
		"index": {
			"partial_filter_selector": {
				"` + expiryField + `": {
					"$ne": 0
				}
			},
			"fields": ["` + expiryField + `"]
		},
		"name": "` + expiryIndexName + `",
		"ddoc": "` + expiryIndexDoc + `",
		"type": "json"
	}`
)

// CouchDBProvider provides an handle to a db
type CouchDBProvider struct {
	couchInstance *couchdb.CouchInstance
	cimutex       sync.RWMutex
	stores        map[string]*dbstore
	mutex         sync.RWMutex
	done          chan struct{}
	closed        bool
}

// NewDBProvider creates a CouchDB Provider
func NewDBProvider() *CouchDBProvider {
	return &CouchDBProvider{
		done:   make(chan struct{}, 1),
		stores: make(map[string]*dbstore),
	}
}

//GetDB based on ns%coll
func (p *CouchDBProvider) GetDB(channelID string, coll string, ns string) (api.DB, error) {
	dbName := dbName(channelID, ns, coll)

	p.mutex.RLock()
	s, ok := p.stores[dbName]
	p.mutex.RUnlock()

	if ok {
		return s, nil
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !ok {
		ci, err := p.getCouchInstance()
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		db, err := couchdb.CreateCouchDatabase(ci, dbName)
		if nil != err {
			logger.Error(err)
			return nil, nil
		}
		s = newDBStore(db, dbName)

		err = db.CreateNewIndexWithRetry(expiryIndexDef, expiryIndexDoc)
		if err != nil {
			return nil, err
		}
		p.stores[dbName] = s
	}

	return s, nil
}

// Close cleans up the Provider
func (p *CouchDBProvider) Close() {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.closed {
		p.done <- struct{}{}
		p.closed = true
	}
}

func (p *CouchDBProvider) getCouchInstance() (*couchdb.CouchInstance, error) {
	p.cimutex.RLock()
	ci := p.couchInstance
	p.cimutex.RUnlock()

	if ci != nil {
		return ci, nil
	}

	return p.createCouchInstance()
}

func (p *CouchDBProvider) createCouchInstance() (*couchdb.CouchInstance, error) {
	p.cimutex.Lock()
	defer p.cimutex.Unlock()

	if p.couchInstance != nil {
		return p.couchInstance, nil
	}

	var err error
	p.couchInstance, err = couchdb.CreateCouchInstance(getCouchDBConfig(), &disabled.Provider{})
	if err != nil {
		return nil, err
	}

	p.periodicPurge()

	return p.couchInstance, nil
}

// periodicPurge goroutine to purge dataModel based on config interval time
func (p *CouchDBProvider) periodicPurge() {
	ticker := time.NewTicker(config.GetOLCollExpirationCheckInterval())
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for _, s := range p.getStores() {
					err := s.DeleteExpiredKeys()
					if err != nil {
						logger.Errorf("Error deleting expired keys for [%s]", s.dbName)
					}
				}
			case <-p.done:
				logger.Infof("Periodic purge is exiting")
				return
			}
		}
	}()
}

// getStores retrieves dbstores contained in the provider
func (p *CouchDBProvider) getStores() []*dbstore {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var stores []*dbstore
	for _, s := range p.stores {
		stores = append(stores, s)
	}
	return stores
}

func dbName(channelID, ns, coll string) string {
	return fmt.Sprintf("%s_%s$$p%s", channelID, ns, coll)
}

// getCouchDBConfig return the couchdb config
// TODO The ledgerconfig can't be passed to offledger provider as cscc calls the createChain which inturn initiates
// CollectionDataStoreFactory(https://github.com/trustbloc/fabric-mod/blob/f195099d41db44623724131f2f487474707e84f2/core/peer/peer.go#L471).
// More over this is using state couchdb configurations. Need to have configs specific to feature/functionality(blockstorage/offledger).
// Created an issue https://github.com/trustbloc/fabric-peer-ext/issues/149. Also, added this as private function to avoid access from external packages.
func getCouchDBConfig() *ledger.CouchDBConfig {
	// set defaults
	warmAfterNBlocks := 1
	if viper.IsSet("ledger.state.couchDBConfig.warmIndexesAfterNBlocks") {
		warmAfterNBlocks = viper.GetInt("ledger.state.couchDBConfig.warmIndexesAfterNBlocks")
	}
	internalQueryLimit := 1000
	if viper.IsSet("ledger.state.couchDBConfig.internalQueryLimit") {
		internalQueryLimit = viper.GetInt("ledger.state.couchDBConfig.internalQueryLimit")
	}
	maxBatchUpdateSize := 500
	if viper.IsSet("ledger.state.couchDBConfig.maxBatchUpdateSize") {
		maxBatchUpdateSize = viper.GetInt("ledger.state.couchDBConfig.maxBatchUpdateSize")
	}
	rootFSPath := filepath.Join(coreconfig.GetPath("peer.fileSystemPath"), "ledgersData")

	return &ledger.CouchDBConfig{
		Address:                 viper.GetString("ledger.state.couchDBConfig.couchDBAddress"),
		Username:                viper.GetString("ledger.state.couchDBConfig.username"),
		Password:                viper.GetString("ledger.state.couchDBConfig.password"),
		MaxRetries:              viper.GetInt("ledger.state.couchDBConfig.maxRetries"),
		MaxRetriesOnStartup:     viper.GetInt("ledger.state.couchDBConfig.maxRetriesOnStartup"),
		RequestTimeout:          viper.GetDuration("ledger.state.couchDBConfig.requestTimeout"),
		InternalQueryLimit:      internalQueryLimit,
		MaxBatchUpdateSize:      maxBatchUpdateSize,
		WarmIndexesAfterNBlocks: warmAfterNBlocks,
		CreateGlobalChangesDB:   viper.GetBool("ledger.state.couchDBConfig.createGlobalChangesDB"),
		RedoLogPath:             filepath.Join(rootFSPath, "couchdbRedoLogs"),
	}
}
