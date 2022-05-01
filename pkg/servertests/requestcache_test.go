package servertests

import (
	"github.com/parkanaur/rpc-relay/pkg/egress"
	"github.com/parkanaur/rpc-relay/pkg/ingress"
	"github.com/parkanaur/rpc-relay/pkg/relayutil"
	"github.com/stretchr/testify/assert"
	"sort"
	"sync"
	"testing"
	"time"
)

func NewTestCache() *ingress.RequestCache {
	return ingress.NewRequestCache(NewTestConfig())
}

func NewRPCCalcRequest(params []any) *egress.RPCRequest {
	return &egress.RPCRequest{
		ModuleName: "calculateSum",
		MethodName: "calculateSum",
		Params:     params,
		ID:         1,
		Method:     "calculateSum_calculateSum",
		JSONRPC:    "2.0",
	}
}

func FillTestCache(cache *ingress.RequestCache, numRequests int) map[string]*egress.RPCRequest {
	wg := sync.WaitGroup{}
	wg.Add(numRequests)

	requests := make(map[string]*egress.RPCRequest)
	for i := 0; i < numRequests; i++ {
		request := NewRPCCalcRequest(append(make([]any, 0), i, i+1))
		requests[request.GetRequestKey()] = request
	}
	for key, request := range requests {
		go func(key string, request *egress.RPCRequest) {
			cache.Add(requests[key], []byte(key))
			wg.Done()
		}(key, request)
	}

	wg.Wait()
	return requests
}

func TestNewRequestCache(t *testing.T) {
	cache := NewTestCache()

	assert.NotNil(t, cache.Cache)
	assert.Equal(t, len(cache.Cache), 0)
}

func TestRequestCache_Add(t *testing.T) {
	egrFixture := NewRelayFixture(t, NewTestConfig())
	defer egrFixture.Shutdown()
	cache := NewTestCache()

	requests := FillTestCache(cache, 5)

	assert.Equal(t, len(cache.Cache), 5)
	cacheValues := make(map[string]*egress.RPCRequest)
	for k, v := range cache.Cache {
		cacheValues[k] = v.Request
	}
	assert.Equal(t, cacheValues, requests)
}

func TestRequestCache_GetRequestByKey(t *testing.T) {
	egrFixture := NewRelayFixture(t, NewTestConfig())
	defer egrFixture.Shutdown()
	cache := NewTestCache()

	requests := FillTestCache(cache, 5)
	for k, _ := range requests {
		cReq, exists := cache.GetRequestByKey(k)
		assert.True(t, exists)
		assert.Equal(t, cReq.Request.GetRequestKey(), k)
	}
}

func TestRequestCache_GetRequestByValue(t *testing.T) {
	egrFixture := NewRelayFixture(t, NewTestConfig())
	defer egrFixture.Shutdown()
	cache := NewTestCache()

	requests := FillTestCache(cache, 5)
	for k, v := range requests {
		cReq, exists := cache.GetRequestByValue(v)
		assert.True(t, exists)
		assert.Equal(t, cReq.Request.GetRequestKey(), k)
	}
}

func TestRequestCache_RemoveByKey(t *testing.T) {
	egrFixture := NewRelayFixture(t, NewTestConfig())
	defer egrFixture.Shutdown()
	cache := NewTestCache()

	requests := FillTestCache(cache, 5)
	wg := sync.WaitGroup{}
	wg.Add(5)

	errors := make([]error, 0, 5)
	errLock := sync.Mutex{}
	for k, _ := range requests {
		go func(key string) {
			err := cache.RemoveByKey(key)
			if err != nil {
				errLock.Lock()
				errors = append(errors, err)
				errLock.Unlock()
			}
			wg.Done()
		}(k)
	}

	wg.Wait()
	assert.Equal(t, len(errors), 0)
	assert.Equal(t, len(cache.Cache), 0)
}

func TestRequestCache_RemoveByValue(t *testing.T) {
	egrFixture := NewRelayFixture(t, NewTestConfig())
	defer egrFixture.Shutdown()
	cache := NewTestCache()

	requests := FillTestCache(cache, 5)
	wg := sync.WaitGroup{}
	wg.Add(5)

	errors := make([]error, 0, 5)
	errLock := sync.Mutex{}
	for _, req := range requests {
		go func(req *egress.RPCRequest) {
			err := cache.RemoveByValue(req)
			if err != nil {
				errLock.Lock()
				errors = append(errors, err)
				errLock.Unlock()
			}
			wg.Done()
		}(req)
	}

	wg.Wait()
	assert.Equal(t, len(errors), 0)
	assert.Equal(t, len(cache.Cache), 0)
}

func TestRequestCache_DeleteStaleValues(t *testing.T) {
	cf := NewTestConfig()
	egrFixture := NewRelayFixture(t, cf)
	defer egrFixture.Shutdown()
	cache := NewTestCache()

	requests := FillTestCache(cache, 5)
	keys := make([]string, 0)
	for k, _ := range requests {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i := 0; i < 3; i++ {
		// Add() and Sub() have different parameter types - ???????????
		cache.Cache[keys[i]].CTime = cache.Cache[keys[i]].CTime.Add(
			-(relayutil.GetDurationInSeconds(cf.Ingress.ExpireCachedRequestThreshold) + time.Second*5))
	}
	cache.DeleteStaleValues(relayutil.GetDurationInSeconds(cf.Ingress.ExpireCachedRequestThreshold))

	assert.Equal(t, len(cache.Cache), 2)
	for i := 3; i < 5; i++ {
		_, contains := cache.GetRequestByKey(keys[i])
		assert.True(t, contains)
	}
}
