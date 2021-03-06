/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txn

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/common/discovery"
	grpcCodes "google.golang.org/grpc/codes"

	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
	"github.com/trustbloc/fabric-peer-ext/pkg/txn/api"
	"github.com/trustbloc/fabric-peer-ext/pkg/txn/client"
	"github.com/trustbloc/fabric-peer-ext/pkg/txn/handler"
)

const (
	configApp     = "txn"
	configVersion = "1"

	generalConfigComponent = "general"
	generalConfigVersion   = "1"

	sdkConfigComponent = "sdk"
	sdkConfigVersion   = "1"
)

type providers struct {
	peerConfig                api.PeerConfig
	configService             config.Service
	clientProvider            clientProvider
	gossip                    gossipapi.GossipService
	proposalResponseValidator api.ProposalResponseValidator
}

// Service implements a Transaction service that gathers multiple endorsements (according to chaincode policy) and
// (optionally) sends the transaction to the Orderer.
type Service struct {
	*providers
	*discovery.Discovery
	channelID       string
	txnCfgKey       *config.Key
	sdkCfgKey       *config.Key
	cfgTxID         string
	c               channelClient
	mutex           sync.RWMutex
	retryOpts       retry.Opts
	commitRetryOpts retry.Opts
}

// New returns a new transaction service
func newService(channelID string, p *providers) (*Service, error) {
	logger.Debugf("[%s] Creating TXN service", channelID)

	s := &Service{
		providers: p,
		channelID: channelID,
		txnCfgKey: config.NewPeerComponentKey(p.peerConfig.MSPID(), p.peerConfig.PeerID(), configApp, configVersion, generalConfigComponent, generalConfigVersion),
		sdkCfgKey: config.NewPeerComponentKey(p.peerConfig.MSPID(), p.peerConfig.PeerID(), configApp, configVersion, sdkConfigComponent, sdkConfigVersion),
		Discovery: discovery.New(channelID, p.gossip),
	}

	if err := s.load(); err != nil {
		return nil, err
	}

	p.configService.AddUpdateHandler(s.handleConfigUpdate)

	return s, nil
}

func (s *Service) handleConfigUpdate(kv *config.KeyValue) {
	logger.Debugf("[%s] Got config update: %s", s.channelID, kv.Key)

	if kv.AppName != configApp || kv.MspID != s.peerConfig.MSPID() || kv.PeerID != s.peerConfig.PeerID() {
		// This config update is not relevant to us
		return
	}

	// If multiple components are updated in the same transaction then we'll get multiple notifications,
	// so avoid reloading the config multiple times by checking the ID of the last transaction that was handled.
	if !s.compareAndSetTxID(kv.TxID) {
		logger.Debugf("[%s] Transaction service config was updated for [%s] but the update for TxID [%s] was already handled", s.channelID, kv.Key, kv.TxID)
		return
	}

	logger.Infof("[%s] Transaction service config was updated for [%s]", s.channelID, kv.Key)

	go func() {
		logger.Debugf("[%s] Reloading transaction service with new config: %s", s.channelID, kv)

		if err := s.load(); err != nil {
			logger.Warnf("Error loading transaction service config: %s", err)
		}
	}()
}

func (s *Service) load() error {
	txnCfg, err := s.getTxnConfig()
	if err != nil {
		return err
	}

	sdkCfg, err := s.getSDKConfig()
	if err != nil {
		return err
	}

	c, err := s.clientProvider.CreateClient(s.channelID, txnCfg.User, s.peerConfig, []byte(sdkCfg.Config), sdkCfg.Format)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.c != nil {
		closableClient, ok := s.c.(closable)
		if ok {
			logger.Debugf("[%s] Closing old client", s.channelID)
			closableClient.Close()
		}
	}

	logger.Debugf("[%s] Loaded client", s.channelID)

	s.c = c
	s.retryOpts = newRetryOpts(txnCfg)
	s.commitRetryOpts = newCommitRetryOpts(s.retryOpts)

	return nil
}

// Endorse collects endorsements according to chaincode policy
func (s *Service) Endorse(req *api.Request) (*channel.Response, error) {
	if err := s.validateTxnIDFromRequest(req); err != nil {
		return nil, err
	}

	h := req.Handler
	if h == nil {
		h = invoke.NewProposalProcessorHandler(
			invoke.NewEndorsementHandlerWithOpts(
				invoke.NewEndorsementValidationHandler(
					invoke.NewSignatureValidationHandler(),
				),
				getTxnOptsProvider(req),
			),
		)
	}

	numRetries := 0
	var lastErr error

	resp, err := s.client().InvokeHandler(
		h, asChannelRequest(req),
		channel.WithTargets(req.Targets...),
		channel.WithTargetFilter(newTargetFilter(newEndorserFilter(s.Discovery, req.PeerFilter))),
		channel.WithRetry(s.retryOpts),
		channel.WithBeforeRetry(s.beforeRetryHandler(&numRetries, &lastErr)))
	if err != nil {
		if numRetries > 0 {
			logger.Infof("[%s] Failed after %d retries. Last error: %s", s.channelID, numRetries, err)
		}

		return nil, err
	}

	if numRetries > 0 {
		logger.Infof("[%s] Succeeded after %d retries. Last error: %s", s.channelID, numRetries, lastErr)
	}

	return &resp, nil
}

// EndorseAndCommit collects endorsements (according to chaincode policy) and sends the endorsements to the Orderer
func (s *Service) EndorseAndCommit(req *api.Request) (*channel.Response, bool, error) {
	if err := s.validateTxnIDFromRequest(req); err != nil {
		return nil, false, err
	}

	checkForCommit := handler.NewCheckForCommitHandler(req.IgnoreNameSpaces, req.CommitType,
		handler.NewCommitHandler(req.AsyncCommit),
	)

	h := req.Handler
	if h == nil {
		h = invoke.NewProposalProcessorHandler(
			invoke.NewEndorsementHandlerWithOpts(
				invoke.NewEndorsementValidationHandler(
					invoke.NewSignatureValidationHandler(
						checkForCommit,
					),
				),
				getTxnOptsProvider(req),
			),
		)
	}

	numRetries := 0
	var lastErr error

	resp, err := s.client().InvokeHandler(
		h, asChannelRequest(req),
		channel.WithTargets(req.Targets...),
		channel.WithTargetFilter(newTargetFilter(newEndorserFilter(s.Discovery, req.PeerFilter))),
		channel.WithRetry(s.retryOpts),
		channel.WithBeforeRetry(s.beforeRetryHandler(&numRetries, &lastErr)))
	if err != nil {
		if numRetries > 0 {
			logger.Infof("[%s] Failed after %d retries. Last error: %s", s.channelID, numRetries, err)
		}

		return nil, false, err
	}

	if numRetries > 0 {
		logger.Infof("[%s] Succeeded after %d retries. Last error: %s", s.channelID, numRetries, lastErr)
	}

	return &resp, checkForCommit.ShouldCommit, nil
}

// CommitEndorsements commits the provided endorsements.
func (s *Service) CommitEndorsements(req *api.CommitRequest) (*channel.Response, bool, error) {
	checkForCommit := handler.NewCheckForCommitHandler(req.IgnoreNameSpaces, req.CommitType,
		handler.NewCommitHandler(req.AsyncCommit),
	)

	h := req.Handler
	if h == nil {
		h = handler.NewPreEndorsedHandler(req.EndorsementResponse, checkForCommit)
	}

	numRetries := 0
	var lastErr error

	resp, err := s.client().InvokeHandler(
		h,
		// put dummy values for ChaincodeID and fcn because sdk requires them even if not used by the handler chain
		channel.Request{ChaincodeID: "cc", Fcn: "fcn"},
		channel.WithRetry(s.commitRetryOpts),
		channel.WithBeforeRetry(s.beforeRetryHandler(&numRetries, &lastErr)))
	if err != nil {
		if numRetries > 0 {
			logger.Infof("[%s] Failed after %d retries. Last error: %s", s.channelID, numRetries, err)
		}

		return nil, false, err
	}

	if numRetries > 0 {
		logger.Infof("[%s] Succeeded after %d retries. Last error: %s", s.channelID, numRetries, lastErr)
	}

	return &resp, checkForCommit.ShouldCommit, nil
}

// SigningIdentity returns the serialized identity of the proposal signer
func (s *Service) SigningIdentity() ([]byte, error) {
	return s.client().SigningIdentity()
}

// GetPeer returns the peer for the given endpoint
func (s *Service) GetPeer(endpoint string) (fab.Peer, error) {
	return s.client().GetPeer(endpoint)
}

// VerifyProposalSignature verifies that the signed proposal is valid
func (s *Service) VerifyProposalSignature(signedProposal *pb.SignedProposal) error {
	return s.client().VerifyProposalSignature(signedProposal)
}

// ValidateProposalResponses validates the given proposal responses
func (s *Service) ValidateProposalResponses(signedProposal *pb.SignedProposal, proposalResponses []*pb.ProposalResponse) (pb.TxValidationCode, error) {
	return s.proposalResponseValidator.Validate(signedProposal, proposalResponses)
}

type closable interface {
	Close()
}

// Close releases the resources for this service
func (s *Service) Close() {
	closableClient, ok := s.client().(closable)
	if ok {
		logger.Debugf("[%s] Closing client", s.channelID)
		closableClient.Close()
	}
}

type txnConfig struct {
	User           string
	RetryAttempts  int
	InitialBackoff string
	MaxBackoff     string
	BackoffFactor  float64
	RetryableCodes []int
}

func (s *Service) client() channelClient {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.c
}

func (s *Service) getTxnConfig() (*txnConfig, error) {
	txnCfg, err := s.configService.Get(s.txnCfgKey)
	if err != nil {
		return nil, errors.WithMessagef(err, "cannot load config for sdkCfgKey %s", s.txnCfgKey)
	}

	txnConfig := &txnConfig{}
	err = json.Unmarshal([]byte(txnCfg.Config), txnConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "error unmarshalling TXN config")
	}

	return txnConfig, nil
}

func (s *Service) getSDKConfig() (*config.Value, error) {
	sdkCfg, err := s.configService.Get(s.sdkCfgKey)
	if err != nil {
		return nil, errors.WithMessagef(err, "cannot load config for sdkCfgKey %s", s.sdkCfgKey)
	}

	return sdkCfg, nil
}

// compareAndSetTxID sets the value of the transaction ID if it's not already set and returns true.
// If the transaction ID is already set then false is returned.
func (s *Service) compareAndSetTxID(txID string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.cfgTxID != txID {
		s.cfgTxID = txID
		return true
	}

	return false
}

func (s *Service) beforeRetryHandler(numRetries *int, lastErr *error) retry.BeforeRetryHandler {
	return func(err error) {
		*numRetries++
		*lastErr = err

		logger.Infof("[%s] Retry #%d on error: %s", s.channelID, *numRetries, err.Error())
	}
}

func (s *Service) validateTxnIDFromRequest(req *api.Request) error {
	if len(req.Nonce) == 0 && req.TransactionID == "" {
		return nil
	}

	if req.TransactionID == "" && len(req.Nonce) > 0 {
		return errors.New("TransactionID must be provided if nonce is present")
	}

	if req.TransactionID != "" && len(req.Nonce) == 0 {
		return errors.New("nonce must be provided if TransactionID is present")
	}

	logger.Debugf("[%s] TransactionID [%s] and nonce were provided in the request. Validating...", s.channelID, req.TransactionID)

	txnID, err := s.client().ComputeTxnID(req.Nonce)
	if err != nil {
		return err
	}

	if txnID != req.TransactionID {
		logger.Debugf("[%s] Invalid TransactionID [%s] - expecting [%s]", s.channelID, req.TransactionID, txnID)

		return api.ErrInvalidTxnID
	}

	return nil
}

func getTxnOptsProvider(req *api.Request) invoke.TxnHeaderOptsProvider {
	if len(req.Nonce) == 0 {
		return nil
	}

	return func() []fabApi.TxnHeaderOpt {
		return []fabApi.TxnHeaderOpt{fabApi.WithNonce(req.Nonce)}
	}
}

func newRetryOpts(cfg *txnConfig) retry.Opts {
	attempts := cfg.RetryAttempts
	initialBackoff, err := time.ParseDuration(cfg.InitialBackoff)
	if err != nil {
		logger.Warnf("Invalid value for InitialBackoff [%s]. Will use default InitialBackoff", cfg.InitialBackoff)
	}

	maxBackoff, err := time.ParseDuration(cfg.MaxBackoff)
	if err != nil {
		logger.Warnf("Invalid value for MaxBackoff [%s]. Will use default MaxBackoff", cfg.MaxBackoff)
	}

	factor := cfg.BackoffFactor

	if attempts == 0 {
		attempts = retry.DefaultAttempts
	}

	if initialBackoff == 0 {
		initialBackoff = retry.DefaultInitialBackoff
	}

	if maxBackoff == 0 {
		maxBackoff = retry.DefaultMaxBackoff
	}

	if factor == 0 {
		factor = retry.DefaultBackoffFactor
	}

	retryableCodes := make(map[status.Group][]status.Code)
	for key, value := range retry.ChannelClientRetryableCodes {
		retryableCodes[key] = value
	}

	for _, code := range cfg.RetryableCodes {
		retryableCodes[status.ChaincodeStatus] = append(retryableCodes[status.ChaincodeStatus], status.Code(code))
	}

	return retry.Opts{
		Attempts:       attempts,
		InitialBackoff: initialBackoff,
		MaxBackoff:     maxBackoff,
		BackoffFactor:  factor,
		RetryableCodes: retryableCodes,
	}
}

func newCommitRetryOpts(opts retry.Opts) retry.Opts {
	opts.RetryableCodes = make(map[status.Group][]status.Code)

	for key, value := range commitOnlyRetryableCodes {
		opts.RetryableCodes[key] = value
	}

	return opts
}

func asInvocationChain(chain []*api.ChaincodeCall) []*fab.ChaincodeCall {
	invocationChain := make([]*fab.ChaincodeCall, len(chain))
	for i, call := range chain {
		invocationChain[i] = &fab.ChaincodeCall{
			ID:          call.ChaincodeName,
			Collections: call.Collections,
		}
	}
	return invocationChain
}

func asChannelRequest(req *api.Request) channel.Request {
	var fcn string
	if len(req.Args) > 0 {
		fcn = string(req.Args[0])
	}

	return channel.Request{
		ChaincodeID:     req.ChaincodeID,
		Fcn:             fcn,
		Args:            req.Args[1:],
		TransientMap:    req.TransientData,
		InvocationChain: asInvocationChain(req.InvocationChain),
	}
}

type channelClient interface {
	client.ChannelClient
	ComputeTxnID(nonce []byte) (string, error)
	SigningIdentity() ([]byte, error)
	GetPeer(endpoint string) (fab.Peer, error)
	VerifyProposalSignature(signedProposal *pb.SignedProposal) error
}

type clientProvider interface {
	CreateClient(channelID, userName string, peerConfig api.PeerConfig, sdkCfgBytes []byte, format config.Format) (channelClient, error)
}

type defaultClientProvider struct {
}

func (p *defaultClientProvider) CreateClient(channelID, userName string, peerConfig api.PeerConfig, sdkCfgBytes []byte, format config.Format) (channelClient, error) {
	return client.New(channelID, userName, peerConfig, sdkCfgBytes, format)
}

type peer struct {
	endpoint string
	mspID    string
}

func (p *peer) Endpoint() string {
	return p.endpoint
}

func (p *peer) MSPID() string {
	return p.mspID
}

func newPeer(mspID, endpoint string) *peer {
	return &peer{
		endpoint: endpoint,
		mspID:    mspID,
	}
}

type targetFilter struct {
	filter api.PeerFilter
}

func (f *targetFilter) Accept(peer fab.Peer) bool {
	return f.filter.Accept(newPeer(peer.MSPID(), peer.URL()))
}

func newTargetFilter(filter api.PeerFilter) fab.TargetFilter {
	if filter == nil {
		return nil
	}

	return &targetFilter{filter: filter}
}

// commitOnlyRetryableCodes are the suggested codes for commit only
var commitOnlyRetryableCodes = map[status.Group][]status.Code{
	status.OrdererClientStatus: {
		status.ConnectionFailed,
	},
	status.OrdererServerStatus: {
		status.Code(common.Status_SERVICE_UNAVAILABLE),
		status.Code(common.Status_INTERNAL_SERVER_ERROR),
	},
	// TODO: gRPC introduced retries in v1.8.0. This can be replaced with the
	// gRPC fail fast option, once available
	status.GRPCTransportStatus: {
		status.Code(grpcCodes.Unavailable),
	},
}
