/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"encoding/hex"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	fabapi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	sdkconfig "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"

	"github.com/trustbloc/fabric-peer-ext/pkg/common/reference"
	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
	"github.com/trustbloc/fabric-peer-ext/pkg/txn/api"
)

var logger = flogging.MustGetLogger("ext_txn")

// ChannelClient defines functions to collect endorsements and send them to the orderer
type ChannelClient interface {
	InvokeHandler(handler invoke.Handler, request channel.Request, options ...channel.RequestOption) (channel.Response, error)
}

type identitySerializer interface {
	Serialize() ([]byte, error)
}

type cryptoSuiteProvider interface {
	CryptoSuite() core.CryptoSuite
}

// Client holds an SDK client instance
type Client struct {
	ChannelClient
	identitySerializer
	cryptoSuiteProvider
	api.PeerConfig
	fabapi.DiscoveryService
	refCount  *reference.Counter
	channelID string
	sdk       *fabsdk.FabricSDK
}

// New returns a new instance of an SDK client for the given channel
func New(channelID, userName string, peerConfig api.PeerConfig, sdkCfgBytes []byte, format config.Format) (*Client, error) {
	configProvider, endpointConfig, err := GetEndpointConfig(sdkCfgBytes, format)
	if err != nil {
		return nil, err
	}

	org, err := orgFromMSPID(endpointConfig, peerConfig)
	if err != nil {
		return nil, err
	}

	customEndpointConfig, err := newEndpointConfig(endpointConfig, peerConfig)
	if err != nil {
		return nil, err
	}

	sdk, err := newSDK(channelID, configProvider, customEndpointConfig, peerConfig)
	if err != nil {
		return nil, err
	}

	chClient, ids, csp, discovery, err := newChannelClient(channelID, userName, org, sdk)
	if err != nil {
		return nil, err
	}

	c := &Client{
		ChannelClient:       chClient,
		identitySerializer:  ids,
		cryptoSuiteProvider: csp,
		PeerConfig:          peerConfig,
		channelID:           channelID,
		sdk:                 sdk,
		DiscoveryService:    discovery,
	}

	c.refCount = reference.NewCounter(c.close)

	return c, nil
}

// InvokeHandler invokes the given handler chain.
func (c *Client) InvokeHandler(handler invoke.Handler, request channel.Request, options ...channel.RequestOption) (channel.Response, error) {
	_, err := c.refCount.Increment()
	if err != nil {
		return channel.Response{}, err
	}
	defer c.decrementCounter()

	return c.ChannelClient.InvokeHandler(handler, request, options...)
}

// ComputeTxnID returns a transaction ID computed using the given nonce and the identity in the channel context
func (c *Client) ComputeTxnID(nonce []byte) (string, error) {
	_, err := c.refCount.Increment()
	if err != nil {
		return "", err
	}
	defer c.decrementCounter()

	creator, err := c.Serialize()
	if err != nil {
		return "", errors.WithMessagef(err, "error serializing identity")
	}

	hash, err := c.CryptoSuite().GetHash(cryptosuite.GetSHA256Opts())
	if err != nil {
		return "", errors.WithMessagef(err, "hash function creation failed")
	}

	b := append(nonce, creator...)

	_, err = hash.Write(b)
	if err != nil {
		return "", errors.WithMessagef(err, "hashing of nonce and creator failed")
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// SigningIdentity returns the serialized identity of the proposal signer
func (c *Client) SigningIdentity() ([]byte, error) {
	identity, err := c.Serialize()
	if err != nil {
		return nil, errors.WithMessagef(err, "error serializing identity")
	}

	return identity, nil
}

// GetPeer returns the peer matching the given endpoint
func (c *Client) GetPeer(endpoint string) (fabapi.Peer, error) {
	logger.Debugf("[%s] Finding peer through discovery for URL [%s]", c.channelID, endpoint)

	peers, err := c.GetPeers()
	if err != nil {
		return nil, errors.WithMessagef(err, "Failed to get peers from discovery service")
	}

	logger.Debugf("[%d] Found %d peers through discovery", c.channelID, len(peers))

	for _, peer := range peers {
		peerURL := peer.URL()

		if strings.EqualFold(endpoint, peerURL) {
			logger.Debugf("[%s] Selecting discovered peer [%s]", c.channelID, peer.URL())
			return peer, nil
		} else if strings.Contains(peerURL, "://") {
			if strings.EqualFold(endpoint, strings.Split(peerURL, "://")[1]) {
				logger.Debugf("[%s] Selecting discovered peer [%s]", c.channelID, peer.URL())
				return peer, nil
			}
		}
		logger.Debugf("[%s] Discovered peer[%s] did not match selected peer [%s]", c.channelID, peer.URL(), endpoint)
	}

	logger.Debugf("[%s] Failed to get matching discovered peer for given URL [%s]", c.channelID, endpoint)
	return nil, errors.Errorf("peer [%s] not found", endpoint)
}

// Close will close the SDK after all references have been released.
func (c *Client) Close() {
	c.refCount.Close()
}

func (c *Client) close() {
	if c.sdk != nil {
		logger.Debugf("[%s] Closing the SDK", c.channelID)
		c.sdk.Close()
	}
}

func (c *Client) decrementCounter() {
	_, err := c.refCount.Decrement()
	if err != nil {
		logger.Warning(err.Error())
	}
}

func orgFromMSPID(endpointConfig fabapi.EndpointConfig, peerCfg api.PeerConfig) (string, error) {
	for orgName, org := range endpointConfig.NetworkConfig().Organizations {
		if org.MSPID == peerCfg.MSPID() {
			return orgName, nil
		}
	}

	return "", errors.Errorf("org not configured for MSP [%s]", peerCfg.MSPID())
}

// GetEndpointConfig unmarshals the given bytes and returns the SDK endpoint config and config provider.
func GetEndpointConfig(configBytes []byte, format config.Format) (core.ConfigProvider, fabapi.EndpointConfig, error) {
	configProvider := func() ([]core.ConfigBackend, error) {
		// Make sure the buffer is created each time it is called, otherwise
		// there will be no data left in the buffer the second time it's called
		return sdkconfig.FromRaw(configBytes, string(format))()
	}

	configBackends, err := configProvider()
	if err != nil {
		return nil, nil, err
	}

	endpointConfig, err := fab.ConfigFromBackend(configBackends...)
	if err != nil {
		return nil, nil, err
	}

	return configProvider, endpointConfig, nil
}

var newSDK = func(channelID string, configProvider core.ConfigProvider, config fabapi.EndpointConfig, peerCfg api.PeerConfig) (*fabsdk.FabricSDK, error) {
	sdk, err := fabsdk.New(
		configProvider,
		fabsdk.WithEndpointConfig(config),
		fabsdk.WithCorePkg(newCorePkg()),
		fabsdk.WithMSPPkg(newMSPPkg(peerCfg.MSPConfigPath())),
	)
	if err != nil {
		return nil, errors.WithMessagef(err, "Error creating SDK on channel [%s]", channelID)
	}

	return sdk, nil
}

var newChannelClient = func(channelID, userName, org string, sdk *fabsdk.FabricSDK) (ChannelClient, identitySerializer, cryptoSuiteProvider, fabapi.DiscoveryService, error) {
	ctx, err := sdk.ChannelContext(channelID, fabsdk.WithUser(userName), fabsdk.WithOrg(org))()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	client, err := channel.New(func() (context.Channel, error) { return ctx, nil })
	if err != nil {
		return nil, nil, nil, nil, err
	}

	discovery, err := ctx.ChannelService().Discovery()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return client, ctx, ctx, discovery, nil
}
