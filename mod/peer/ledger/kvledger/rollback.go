/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package kvledger

import (
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/common/ledger/util/leveldbhelper"
	"github.com/hyperledger/fabric/common/metrics/disabled"
	"github.com/hyperledger/fabric/core/ledger"
	"github.com/hyperledger/fabric/core/ledger/kvledger"
	couchdb "github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb/statecouchdb"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/blkstorage/cdbblkstorage"
	"github.com/trustbloc/fabric-peer-ext/pkg/config"
)

var logger = flogging.MustGetLogger("ext_kvledger")

func RollbackKVLedger(ledgerconfig *ledger.Config, ledgerID string, blockNum uint64) error {
	if config.GetBlockStoreDBType() == config.CouchDBType {
		fileLockPath := fileLockPath(ledgerconfig.RootFSPath)
		fileLock := leveldbhelper.NewFileLock(fileLockPath)
		if err := fileLock.Lock(); err != nil {
			return errors.Wrap(err, "as another peer node command is executing,"+
				" wait for that command to complete its execution or terminate it before retrying")
		}
		defer fileLock.Unlock()

		stateDBCouchInstance, err := couchdb.CreateCouchInstance(ledgerconfig.StateDBConfig.CouchDB, &disabled.Provider{})
		if err != nil {
			return errors.WithMessage(err, "obtaining CouchDB instance failed")
		}

		if err := cdbblkstorage.ValidateRollbackParams(stateDBCouchInstance, ledgerID, blockNum); err != nil {
			return err
		}

		logger.Infof("Dropping databases")
		if err := dropDBs(ledgerconfig); err != nil {
			return err
		}

		logger.Info("Rolling back ledger store")
		if err := cdbblkstorage.Rollback(stateDBCouchInstance, ledgerconfig.StateDBConfig.CouchDB.InternalQueryLimit, ledgerID, blockNum); err != nil {
			return err
		}
		logger.Infof("The channel [%s] has been successfully rolled back to the block number [%d]", ledgerID, blockNum)
		return nil
	}

	return kvledger.RollbackKVLedger(ledgerconfig.RootFSPath, ledgerID, blockNum)
}
