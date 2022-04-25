package ingress

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"rpc-relay/pkg/egress"
	"rpc-relay/pkg/relayutil"
	"sync"
	"time"
)

type CachedRequest struct {
	ctime   time.Time
	request *egress.RPCRequest
	// egress.RPCResponse serialized into bytes. Since we don't need to unmarshal it at any point to return to the
	// user, it's wise to keep it here as raw bytes.
	response []byte
}

func (request *CachedRequest) IsRequestStale(timeToLive time.Duration) bool {
	return request.ctime.Add(timeToLive).Before(time.Now())
}

type RequestCache struct {
	sync.RWMutex
	// TODO: think of using sync.Map?
	Cache map[string]*CachedRequest
}

func NewRequestCache() *RequestCache {
	return &RequestCache{
		Cache: make(map[string]*CachedRequest),
	}
}

func (cache *RequestCache) Add(request *egress.RPCRequest, response []byte) {
	cache.Lock()
	defer cache.Unlock()

	cache.Cache[request.GetRequestKey()] = &CachedRequest{time.Now(), request, response}
}

func (cache *RequestCache) GetRequestByKey(requestKey string) (*CachedRequest, bool) {
	cache.RLock()
	defer cache.RUnlock()

	request, ok := cache.Cache[requestKey]
	return request, ok
}

func (cache *RequestCache) GetRequestByValue(request *egress.RPCRequest) (*CachedRequest, bool) {
	return cache.GetRequestByKey(request.GetRequestKey())
}

func (cache *RequestCache) RemoveByKey(requestKey string) error {
	if _, ok := cache.GetRequestByKey(requestKey); !ok {
		return fmt.Errorf("request not found in cache")
	}

	cache.Lock()
	defer cache.Unlock()
	delete(cache.Cache, requestKey)
	return nil
}

func (cache *RequestCache) RemoveByValue(request *egress.RPCRequest) error {
	return cache.RemoveByKey(request.GetRequestKey())
}

func (cache *RequestCache) RequestKeyStale(requestKey string, timeToLive time.Duration) bool {
	cachedRequest, ok := cache.GetRequestByKey(requestKey)
	if !ok {
		return false
	}
	return cachedRequest.IsRequestStale(timeToLive)
}

func (cache *RequestCache) DeleteStaleValues(timeToLive time.Duration) {
	for requestKey, cachedRequest := range cache.Cache {
		if cachedRequest.IsRequestStale(timeToLive) {
			err := cache.RemoveByKey(requestKey)
			if err != nil {
				log.Errorln("Failed to delete request from cache", err)
			}
		}
	}
}

func (cache *RequestCache) InvalidateStaleValuesLoop(config *relayutil.Config, done <-chan bool) {
	for {
		select {
		case <-done:
			return
		default:
			log.Infoln("Cleaning up cache, size:", len(cache.Cache))
			cache.DeleteStaleValues(relayutil.GetDurationInSeconds(config.Ingress.ExpireCachedRequestThreshold))
			log.Infoln("Cache invalidated, size:", len(cache.Cache))
			time.Sleep(relayutil.GetDurationInSeconds(config.Ingress.InvalidateCacheLoopSleepPeriod))
		}
	}
}
