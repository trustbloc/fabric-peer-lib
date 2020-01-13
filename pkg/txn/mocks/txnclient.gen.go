// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/trustbloc/fabric-peer-ext/pkg/txn/api"
)

type TxnClient struct {
	EndorseStub        func(req *api.Request) (*channel.Response, error)
	endorseMutex       sync.RWMutex
	endorseArgsForCall []struct {
		req *api.Request
	}
	endorseReturns struct {
		result1 *channel.Response
		result2 error
	}
	endorseReturnsOnCall map[int]struct {
		result1 *channel.Response
		result2 error
	}
	EndorseAndCommitStub        func(req *api.Request) (*channel.Response, error)
	endorseAndCommitMutex       sync.RWMutex
	endorseAndCommitArgsForCall []struct {
		req *api.Request
	}
	endorseAndCommitReturns struct {
		result1 *channel.Response
		result2 error
	}
	endorseAndCommitReturnsOnCall map[int]struct {
		result1 *channel.Response
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *TxnClient) Endorse(req *api.Request) (*channel.Response, error) {
	fake.endorseMutex.Lock()
	ret, specificReturn := fake.endorseReturnsOnCall[len(fake.endorseArgsForCall)]
	fake.endorseArgsForCall = append(fake.endorseArgsForCall, struct {
		req *api.Request
	}{req})
	fake.recordInvocation("Endorse", []interface{}{req})
	fake.endorseMutex.Unlock()
	if fake.EndorseStub != nil {
		return fake.EndorseStub(req)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.endorseReturns.result1, fake.endorseReturns.result2
}

func (fake *TxnClient) EndorseCallCount() int {
	fake.endorseMutex.RLock()
	defer fake.endorseMutex.RUnlock()
	return len(fake.endorseArgsForCall)
}

func (fake *TxnClient) EndorseArgsForCall(i int) *api.Request {
	fake.endorseMutex.RLock()
	defer fake.endorseMutex.RUnlock()
	return fake.endorseArgsForCall[i].req
}

func (fake *TxnClient) EndorseReturns(result1 *channel.Response, result2 error) {
	fake.EndorseStub = nil
	fake.endorseReturns = struct {
		result1 *channel.Response
		result2 error
	}{result1, result2}
}

func (fake *TxnClient) EndorseReturnsOnCall(i int, result1 *channel.Response, result2 error) {
	fake.EndorseStub = nil
	if fake.endorseReturnsOnCall == nil {
		fake.endorseReturnsOnCall = make(map[int]struct {
			result1 *channel.Response
			result2 error
		})
	}
	fake.endorseReturnsOnCall[i] = struct {
		result1 *channel.Response
		result2 error
	}{result1, result2}
}

func (fake *TxnClient) EndorseAndCommit(req *api.Request) (*channel.Response, error) {
	fake.endorseAndCommitMutex.Lock()
	ret, specificReturn := fake.endorseAndCommitReturnsOnCall[len(fake.endorseAndCommitArgsForCall)]
	fake.endorseAndCommitArgsForCall = append(fake.endorseAndCommitArgsForCall, struct {
		req *api.Request
	}{req})
	fake.recordInvocation("EndorseAndCommit", []interface{}{req})
	fake.endorseAndCommitMutex.Unlock()
	if fake.EndorseAndCommitStub != nil {
		return fake.EndorseAndCommitStub(req)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.endorseAndCommitReturns.result1, fake.endorseAndCommitReturns.result2
}

func (fake *TxnClient) EndorseAndCommitCallCount() int {
	fake.endorseAndCommitMutex.RLock()
	defer fake.endorseAndCommitMutex.RUnlock()
	return len(fake.endorseAndCommitArgsForCall)
}

func (fake *TxnClient) EndorseAndCommitArgsForCall(i int) *api.Request {
	fake.endorseAndCommitMutex.RLock()
	defer fake.endorseAndCommitMutex.RUnlock()
	return fake.endorseAndCommitArgsForCall[i].req
}

func (fake *TxnClient) EndorseAndCommitReturns(result1 *channel.Response, result2 error) {
	fake.EndorseAndCommitStub = nil
	fake.endorseAndCommitReturns = struct {
		result1 *channel.Response
		result2 error
	}{result1, result2}
}

func (fake *TxnClient) EndorseAndCommitReturnsOnCall(i int, result1 *channel.Response, result2 error) {
	fake.EndorseAndCommitStub = nil
	if fake.endorseAndCommitReturnsOnCall == nil {
		fake.endorseAndCommitReturnsOnCall = make(map[int]struct {
			result1 *channel.Response
			result2 error
		})
	}
	fake.endorseAndCommitReturnsOnCall[i] = struct {
		result1 *channel.Response
		result2 error
	}{result1, result2}
}

func (fake *TxnClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.endorseMutex.RLock()
	defer fake.endorseMutex.RUnlock()
	fake.endorseAndCommitMutex.RLock()
	defer fake.endorseAndCommitMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *TxnClient) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}
