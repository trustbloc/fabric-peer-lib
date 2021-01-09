// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"sync"

	vcommon "github.com/trustbloc/fabric-peer-ext/pkg/validation/common"
)

type DistributedValidatorProvider struct {
	GetValidatorForChannelStub        func(channelID string) vcommon.DistributedValidator
	getValidatorForChannelMutex       sync.RWMutex
	getValidatorForChannelArgsForCall []struct {
		channelID string
	}
	getValidatorForChannelReturns struct {
		result1 vcommon.DistributedValidator
	}
	getValidatorForChannelReturnsOnCall map[int]struct {
		result1 vcommon.DistributedValidator
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *DistributedValidatorProvider) GetValidatorForChannel(channelID string) vcommon.DistributedValidator {
	fake.getValidatorForChannelMutex.Lock()
	ret, specificReturn := fake.getValidatorForChannelReturnsOnCall[len(fake.getValidatorForChannelArgsForCall)]
	fake.getValidatorForChannelArgsForCall = append(fake.getValidatorForChannelArgsForCall, struct {
		channelID string
	}{channelID})
	fake.recordInvocation("GetValidatorForChannel", []interface{}{channelID})
	fake.getValidatorForChannelMutex.Unlock()
	if fake.GetValidatorForChannelStub != nil {
		return fake.GetValidatorForChannelStub(channelID)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.getValidatorForChannelReturns.result1
}

func (fake *DistributedValidatorProvider) GetValidatorForChannelCallCount() int {
	fake.getValidatorForChannelMutex.RLock()
	defer fake.getValidatorForChannelMutex.RUnlock()
	return len(fake.getValidatorForChannelArgsForCall)
}

func (fake *DistributedValidatorProvider) GetValidatorForChannelArgsForCall(i int) string {
	fake.getValidatorForChannelMutex.RLock()
	defer fake.getValidatorForChannelMutex.RUnlock()
	return fake.getValidatorForChannelArgsForCall[i].channelID
}

func (fake *DistributedValidatorProvider) GetValidatorForChannelReturns(result1 vcommon.DistributedValidator) {
	fake.GetValidatorForChannelStub = nil
	fake.getValidatorForChannelReturns = struct {
		result1 vcommon.DistributedValidator
	}{result1}
}

func (fake *DistributedValidatorProvider) GetValidatorForChannelReturnsOnCall(i int, result1 vcommon.DistributedValidator) {
	fake.GetValidatorForChannelStub = nil
	if fake.getValidatorForChannelReturnsOnCall == nil {
		fake.getValidatorForChannelReturnsOnCall = make(map[int]struct {
			result1 vcommon.DistributedValidator
		})
	}
	fake.getValidatorForChannelReturnsOnCall[i] = struct {
		result1 vcommon.DistributedValidator
	}{result1}
}

func (fake *DistributedValidatorProvider) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getValidatorForChannelMutex.RLock()
	defer fake.getValidatorForChannelMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *DistributedValidatorProvider) recordInvocation(key string, args []interface{}) {
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
