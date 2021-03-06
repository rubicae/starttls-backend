package checker

import (
	"fmt"
	"sync"
	"time"
)

// ScanStore is an interface for using and retrieving scan results.
type ScanStore interface {
	GetHostnameScan(string) (HostnameResult, error)
	PutHostnameScan(string, HostnameResult) error
}

// ScanCache wraps a scan storage object. When calling GetScan, only returns a scan
// if there was made in the last ExpireTime window
type ScanCache struct {
	ScanStore
	ExpireTime time.Duration
}

// GetHostnameScan retrieves the scan from underlying storage if there is one
// present within the cached time window.
func (c *ScanCache) GetHostnameScan(hostname string) (HostnameResult, error) {
	result, err := c.ScanStore.GetHostnameScan(hostname)
	if err != nil {
		return result, err
	}
	if time.Now().Sub(result.Timestamp) > c.ExpireTime {
		return result, fmt.Errorf("most recent scan for %s expired", hostname)
	}
	return result, nil
}

// PutHostnameScan puts in a scan.
func (c *ScanCache) PutHostnameScan(hostname string, result HostnameResult) error {
	return c.ScanStore.PutHostnameScan(hostname, result)
}

// SimpleStore is simple HostnameResult storage backed by map.
type SimpleStore struct {
	m  map[string]HostnameResult
	mu sync.RWMutex
}

// GetHostnameScan wraps a map get. Returns error if not present in map.
func (s *SimpleStore) GetHostnameScan(hostname string) (HostnameResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result, ok := s.m[hostname]
	if !ok {
		return result, fmt.Errorf("Couldn't find scan for hostname %s", hostname)
	}
	return result, nil
}

// PutHostnameScan wraps a map set. Can never return error.
func (s *SimpleStore) PutHostnameScan(hostname string, result HostnameResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[hostname] = result
	return nil
}

// MakeSimpleCache creates a cache with a SimpleStore backing it.
func MakeSimpleCache(expiryTime time.Duration) *ScanCache {
	store := SimpleStore{m: make(map[string]HostnameResult)}
	return &ScanCache{ScanStore: &store, ExpireTime: expiryTime}
}
