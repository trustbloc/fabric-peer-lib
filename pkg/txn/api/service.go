/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// Service provides functions to collect endorsements and send endorsements to the Orderer
type Service interface {
	// Endorse collects endorsements according to chaincode policy
	Endorse(req *Request) (resp *channel.Response, err error)

	// EndorseAndCommit collects endorsements (according to chaincode policy) and sends the endorsements to the Orderer.
	// Returns the response and true if the transaction was committed.
	EndorseAndCommit(req *Request) (resp *channel.Response, committed bool, err error)

	// CommitEndorsements commits the provided endorsements. First the endorsements are verified for signature and policy,
	// and then the endorsements are sent to the Orderer.
	CommitEndorsements(req *CommitRequest) (*channel.Response, bool, error)

	// SigningIdentity returns the serialized identity of the proposal signer
	SigningIdentity() ([]byte, error)

	// GetPeer returns the peer for the given endpoint
	GetPeer(endpoint string) (fab.Peer, error)

	// VerifyProposalSignature verifies that the signed proposal is valid
	VerifyProposalSignature(signedProposal *pb.SignedProposal) error

	// ValidateProposalResponses validates the given proposal responses
	ValidateProposalResponses(signedProposal *pb.SignedProposal, proposalResponses []*pb.ProposalResponse) (pb.TxValidationCode, error)
}
