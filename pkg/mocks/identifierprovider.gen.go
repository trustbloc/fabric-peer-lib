// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"sync"

	"github.com/trustbloc/fabric-peer-ext/pkg/collections/common"
)

type IdentifierProvider struct {
	GetIdentifierStub        func() (string, error)
	getIdentifierMutex       sync.RWMutex
	getIdentifierArgsForCall []struct{}
	getIdentifierReturns     struct {
		result1 string
		result2 error
	}
	getIdentifierReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *IdentifierProvider) GetIdentifier() (string, error) {
	fake.getIdentifierMutex.Lock()
	ret, specificReturn := fake.getIdentifierReturnsOnCall[len(fake.getIdentifierArgsForCall)]
	fake.getIdentifierArgsForCall = append(fake.getIdentifierArgsForCall, struct{}{})
	fake.recordInvocation("GetIdentifier", []interface{}{})
	fake.getIdentifierMutex.Unlock()
	if fake.GetIdentifierStub != nil {
		return fake.GetIdentifierStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.getIdentifierReturns.result1, fake.getIdentifierReturns.result2
}

func (fake *IdentifierProvider) GetIdentifierCallCount() int {
	fake.getIdentifierMutex.RLock()
	defer fake.getIdentifierMutex.RUnlock()
	return len(fake.getIdentifierArgsForCall)
}

func (fake *IdentifierProvider) GetIdentifierReturns(result1 string, result2 error) {
	fake.GetIdentifierStub = nil
	fake.getIdentifierReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *IdentifierProvider) GetIdentifierReturnsOnCall(i int, result1 string, result2 error) {
	fake.GetIdentifierStub = nil
	if fake.getIdentifierReturnsOnCall == nil {
		fake.getIdentifierReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.getIdentifierReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *IdentifierProvider) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getIdentifierMutex.RLock()
	defer fake.getIdentifierMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *IdentifierProvider) recordInvocation(key string, args []interface{}) {
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

var _ common.IdentifierProvider = new(IdentifierProvider)
