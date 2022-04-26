package ingress

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"rpc-relay/pkg/egress"
	"rpc-relay/pkg/relayutil"
	"sync"
	"time"
)

// CachedRequest holds the RPC request as well as the time it was added to the queue.
// Response is also cached here as bytes
type CachedRequest struct {
	ctime   time.Time
	request *egress.RPCRequest
	// egress.RPCResponse serialized into bytes. Since we don't need to unmarshal it at any point to return to the
	// user, it's wise to keep it here as raw bytes.
	response []byte
}

// IsRequestStale compares the request cache time to current time and checks if it has exceeded the given
// time to live
func (request *CachedRequest) IsRequestStale(timeToLive time.Duration) bool {
	return request.ctime.Add(timeToLive).Before(time.Now())
}

// RequestCache is a wrapper around the map which maps request keys (see RPCRequest) to
// corresponding cached requests
type RequestCache struct {
	sync.RWMutex
	// TODO: think of using sync.Map?
	Cache map[string]*CachedRequest
}

// NewRequestCache returns an empty request cache
func NewRequestCache() *RequestCache {
	return &RequestCache{
		Cache: make(map[string]*CachedRequest),
	}
}

// Add adds a new RPCRequest and its response to the cache
func (cache *RequestCache) Add(request *egress.RPCRequest, response []byte) {
	cache.Lock()
	defer cache.Unlock()

	cache.Cache[request.GetRequestKey()] = &CachedRequest{time.Now(), request, response}
}

// GetRequestByKey searches for and returns the cached request by its key
func (cache *RequestCache) GetRequestByKey(requestKey string) (*CachedRequest, bool) {
	cache.RLock()
	defer cache.RUnlock()

	request, ok := cache.Cache[requestKey]
	return request, ok
}

// GetRequestByValue searches for and returns the cached request by its key
func (cache *RequestCache) GetRequestByValue(request *egress.RPCRequest) (*CachedRequest, bool) {
	return cache.GetRequestByKey(request.GetRequestKey())
}

// RemoveByKey removes a request from the cache by its key
func (cache *RequestCache) RemoveByKey(requestKey string) error {
	if _, ok := cache.GetRequestByKey(requestKey); !ok {
		return fmt.Errorf("request not found in cache")
	}

	cache.Lock()
	defer cache.Unlock()
	delete(cache.Cache, requestKey)
	return nil
}

// RemoveByValue removes a request from the cache by its key
func (cache *RequestCache) RemoveByValue(request *egress.RPCRequest) error {
	return cache.RemoveByKey(request.GetRequestKey())
}

// RequestKeyStale checks if the request pointed to by its key has exceeded the timeToLive duration
func (cache *RequestCache) RequestKeyStale(requestKey string, timeToLive time.Duration) bool {
	cachedRequest, ok := cache.GetRequestByKey(requestKey)
	if !ok {
		return false
	}
	return cachedRequest.IsRequestStale(timeToLive)
}

// DeleteStaleValues runs through the whole cache and removes the old enough entries
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

// InvalidateStaleValuesLoop runs the DeleteStaleValues method every N seconds,
// where N is defined by config's ingress.expireCachedRequestThreshold key
func (cache *RequestCache) InvalidateStaleValuesLoop(config *relayutil.Config, done <-chan bool, wg *sync.WaitGroup) {
	for {
		select {
		case <-done:
			log.Infoln("Stopped cache")
			wg.Done()
			return
		case <-time.After(relayutil.GetDurationInSeconds(config.Ingress.InvalidateCacheLoopSleepPeriod)):
			log.Infoln("Cleaning up cache, size:", len(cache.Cache))
			cache.DeleteStaleValues(relayutil.GetDurationInSeconds(config.Ingress.ExpireCachedRequestThreshold))
			log.Infoln("Cache invalidated, size:", len(cache.Cache))
		}
	}
}
