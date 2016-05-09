// This file was generated by counterfeiter
package clientfakes

import "sync"

type FakeHttpClient struct {
	PutStub        func(route string, requestData interface{}, responseData interface{}) error
	putMutex       sync.RWMutex
	putArgsForCall []struct {
		route        string
		requestData  interface{}
		responseData interface{}
	}
	putReturns struct {
		result1 error
	}
}

func (fake *FakeHttpClient) Put(route string, requestData interface{}, responseData interface{}) error {
	fake.putMutex.Lock()
	fake.putArgsForCall = append(fake.putArgsForCall, struct {
		route        string
		requestData  interface{}
		responseData interface{}
	}{route, requestData, responseData})
	fake.putMutex.Unlock()
	if fake.PutStub != nil {
		return fake.PutStub(route, requestData, responseData)
	} else {
		return fake.putReturns.result1
	}
}

func (fake *FakeHttpClient) PutCallCount() int {
	fake.putMutex.RLock()
	defer fake.putMutex.RUnlock()
	return len(fake.putArgsForCall)
}

func (fake *FakeHttpClient) PutArgsForCall(i int) (string, interface{}, interface{}) {
	fake.putMutex.RLock()
	defer fake.putMutex.RUnlock()
	return fake.putArgsForCall[i].route, fake.putArgsForCall[i].requestData, fake.putArgsForCall[i].responseData
}

func (fake *FakeHttpClient) PutReturns(result1 error) {
	fake.PutStub = nil
	fake.putReturns = struct {
		result1 error
	}{result1}
}