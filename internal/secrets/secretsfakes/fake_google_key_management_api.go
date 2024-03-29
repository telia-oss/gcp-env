// Code generated by counterfeiter. DO NOT EDIT.
package secretsfakes

import (
	"context"
	"sync"

	gax "github.com/googleapis/gax-go/v2"
	"github.com/telia-oss/gcp-env/internal/secrets"
	kms "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

type FakeGoogleKeyManagementAPI struct {
	DecryptStub        func(context.Context, *kms.DecryptRequest, ...gax.CallOption) (*kms.DecryptResponse, error)
	decryptMutex       sync.RWMutex
	decryptArgsForCall []struct {
		arg1 context.Context
		arg2 *kms.DecryptRequest
		arg3 []gax.CallOption
	}
	decryptReturns struct {
		result1 *kms.DecryptResponse
		result2 error
	}
	decryptReturnsOnCall map[int]struct {
		result1 *kms.DecryptResponse
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeGoogleKeyManagementAPI) Decrypt(arg1 context.Context, arg2 *kms.DecryptRequest, arg3 ...gax.CallOption) (*kms.DecryptResponse, error) {
	fake.decryptMutex.Lock()
	ret, specificReturn := fake.decryptReturnsOnCall[len(fake.decryptArgsForCall)]
	fake.decryptArgsForCall = append(fake.decryptArgsForCall, struct {
		arg1 context.Context
		arg2 *kms.DecryptRequest
		arg3 []gax.CallOption
	}{arg1, arg2, arg3})
	stub := fake.DecryptStub
	fakeReturns := fake.decryptReturns
	fake.recordInvocation("Decrypt", []interface{}{arg1, arg2, arg3})
	fake.decryptMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3...)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeGoogleKeyManagementAPI) DecryptCallCount() int {
	fake.decryptMutex.RLock()
	defer fake.decryptMutex.RUnlock()
	return len(fake.decryptArgsForCall)
}

func (fake *FakeGoogleKeyManagementAPI) DecryptCalls(stub func(context.Context, *kms.DecryptRequest, ...gax.CallOption) (*kms.DecryptResponse, error)) {
	fake.decryptMutex.Lock()
	defer fake.decryptMutex.Unlock()
	fake.DecryptStub = stub
}

func (fake *FakeGoogleKeyManagementAPI) DecryptArgsForCall(i int) (context.Context, *kms.DecryptRequest, []gax.CallOption) {
	fake.decryptMutex.RLock()
	defer fake.decryptMutex.RUnlock()
	argsForCall := fake.decryptArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeGoogleKeyManagementAPI) DecryptReturns(result1 *kms.DecryptResponse, result2 error) {
	fake.decryptMutex.Lock()
	defer fake.decryptMutex.Unlock()
	fake.DecryptStub = nil
	fake.decryptReturns = struct {
		result1 *kms.DecryptResponse
		result2 error
	}{result1, result2}
}

func (fake *FakeGoogleKeyManagementAPI) DecryptReturnsOnCall(i int, result1 *kms.DecryptResponse, result2 error) {
	fake.decryptMutex.Lock()
	defer fake.decryptMutex.Unlock()
	fake.DecryptStub = nil
	if fake.decryptReturnsOnCall == nil {
		fake.decryptReturnsOnCall = make(map[int]struct {
			result1 *kms.DecryptResponse
			result2 error
		})
	}
	fake.decryptReturnsOnCall[i] = struct {
		result1 *kms.DecryptResponse
		result2 error
	}{result1, result2}
}

func (fake *FakeGoogleKeyManagementAPI) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.decryptMutex.RLock()
	defer fake.decryptMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeGoogleKeyManagementAPI) recordInvocation(key string, args []interface{}) {
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

var _ secrets.GoogleKeyManagementAPI = new(FakeGoogleKeyManagementAPI)
