// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"sync"

	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
)

type ConfigValidatorRegistry struct {
	RegisterStub        func(v config.Validator)
	registerMutex       sync.RWMutex
	registerArgsForCall []struct {
		v config.Validator
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *ConfigValidatorRegistry) Register(v config.Validator) {
	fake.registerMutex.Lock()
	fake.registerArgsForCall = append(fake.registerArgsForCall, struct {
		v config.Validator
	}{v})
	fake.recordInvocation("Register", []interface{}{v})
	fake.registerMutex.Unlock()
	if fake.RegisterStub != nil {
		fake.RegisterStub(v)
	}
}

func (fake *ConfigValidatorRegistry) RegisterCallCount() int {
	fake.registerMutex.RLock()
	defer fake.registerMutex.RUnlock()
	return len(fake.registerArgsForCall)
}

func (fake *ConfigValidatorRegistry) RegisterArgsForCall(i int) config.Validator {
	fake.registerMutex.RLock()
	defer fake.registerMutex.RUnlock()
	return fake.registerArgsForCall[i].v
}

func (fake *ConfigValidatorRegistry) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.registerMutex.RLock()
	defer fake.registerMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *ConfigValidatorRegistry) recordInvocation(key string, args []interface{}) {
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
