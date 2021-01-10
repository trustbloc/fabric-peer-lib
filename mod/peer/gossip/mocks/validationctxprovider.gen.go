// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"sync"
)

type ValidationContextProvider struct {
	CancelBlockValidationStub        func(channelID string, blockNum uint64)
	cancelBlockValidationMutex       sync.RWMutex
	cancelBlockValidationArgsForCall []struct {
		channelID string
		blockNum  uint64
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *ValidationContextProvider) CancelBlockValidation(channelID string, blockNum uint64) {
	fake.cancelBlockValidationMutex.Lock()
	fake.cancelBlockValidationArgsForCall = append(fake.cancelBlockValidationArgsForCall, struct {
		channelID string
		blockNum  uint64
	}{channelID, blockNum})
	fake.recordInvocation("CancelBlockValidation", []interface{}{channelID, blockNum})
	fake.cancelBlockValidationMutex.Unlock()
	if fake.CancelBlockValidationStub != nil {
		fake.CancelBlockValidationStub(channelID, blockNum)
	}
}

func (fake *ValidationContextProvider) CancelBlockValidationCallCount() int {
	fake.cancelBlockValidationMutex.RLock()
	defer fake.cancelBlockValidationMutex.RUnlock()
	return len(fake.cancelBlockValidationArgsForCall)
}

func (fake *ValidationContextProvider) CancelBlockValidationArgsForCall(i int) (string, uint64) {
	fake.cancelBlockValidationMutex.RLock()
	defer fake.cancelBlockValidationMutex.RUnlock()
	return fake.cancelBlockValidationArgsForCall[i].channelID, fake.cancelBlockValidationArgsForCall[i].blockNum
}

func (fake *ValidationContextProvider) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.cancelBlockValidationMutex.RLock()
	defer fake.cancelBlockValidationMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *ValidationContextProvider) recordInvocation(key string, args []interface{}) {
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
