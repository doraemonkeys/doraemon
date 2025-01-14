package doraemon

import (
	"fmt"
	"hash/fnv"
	"sync"
)

// ShardedMap is a concurrent map sharded across multiple locks.
type ShardedMap[K comparable, V any] struct {
	locks     []sync.RWMutex  // Locks for each shard
	mp        []map[K]V       // Array of maps, each representing a shard
	shardFunc func(key K) int // Function to determine shard index for a key
}

// DefaultCalc returns a default sharding function using FNV-1a hash.
func DefaultCalc[K comparable](shardCount int) func(key K) int {
	return func(key K) int {
		h := fnv.New32a()
		_, _ = fmt.Fprint(h, key)
		return int(h.Sum32() % uint32(shardCount))
	}
}

// NewMap creates a new ShardedMap with the specified shard count and sharding function.
// If calcFunc is nil, it uses DefaultCalc.
func NewMap[K comparable, V any](shardCount int, calcFunc func(key K) int) *ShardedMap[K, V] {
	if calcFunc == nil {
		calcFunc = DefaultCalc[K](shardCount)
	}
	locks := make([]sync.RWMutex, shardCount)
	mp := make([]map[K]V, shardCount)
	for i := 0; i < shardCount; i++ {
		mp[i] = make(map[K]V)
	}
	return &ShardedMap[K, V]{locks: locks, mp: mp, shardFunc: calcFunc}
}

// Get retrieves a value from the map, returning the value and a boolean indicating success.
func (m *ShardedMap[K, V]) Get(key K) (V, bool) {
	shardIndex := m.shardFunc(key)
	m.locks[shardIndex].RLock()
	defer m.locks[shardIndex].RUnlock()
	value, ok := m.mp[shardIndex][key]
	return value, ok
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores the given value and returns it.
func (m *ShardedMap[K, V]) LoadOrStore(key K, value V) V {
	shardIndex := m.shardFunc(key)
	m.locks[shardIndex].Lock()
	defer m.locks[shardIndex].Unlock()
	v, ok := m.mp[shardIndex][key]
	if !ok {
		m.mp[shardIndex][key] = value
		return value
	}
	return v
}

// Set stores a value for the given key.
func (m *ShardedMap[K, V]) Set(key K, value V) {
	shardIndex := m.shardFunc(key)
	m.locks[shardIndex].Lock()
	defer m.locks[shardIndex].Unlock()
	m.mp[shardIndex][key] = value
}

// StoreIfAbsent stores a value only if the key is not present.
func (m *ShardedMap[K, V]) StoreIfAbsent(key K, value V) {
	shardIndex := m.shardFunc(key)
	m.locks[shardIndex].Lock()
	defer m.locks[shardIndex].Unlock()
	if _, ok := m.mp[shardIndex][key]; !ok {
		m.mp[shardIndex][key] = value
	}
}

// StoreIfPresent stores a value only if the key is already present.
func (m *ShardedMap[K, V]) StoreIfPresent(key K, value V) {
	shardIndex := m.shardFunc(key)
	m.locks[shardIndex].Lock()
	defer m.locks[shardIndex].Unlock()
	if _, ok := m.mp[shardIndex][key]; ok {
		m.mp[shardIndex][key] = value
	}
}

// Delete removes the key-value pair from the map.
func (m *ShardedMap[K, V]) Delete(key K) {
	shardIndex := m.shardFunc(key)
	m.locks[shardIndex].Lock()
	defer m.locks[shardIndex].Unlock()
	delete(m.mp[shardIndex], key)
}

// Contains checks if a key exists in the map.
func (m *ShardedMap[K, V]) Contains(key K) bool {
	shardIndex := m.shardFunc(key)
	m.locks[shardIndex].RLock()
	defer m.locks[shardIndex].RUnlock()
	_, ok := m.mp[shardIndex][key]
	return ok
}

// Remove removes the key-value pair from the map, returning the removed value and a boolean indicating success.
func (m *ShardedMap[K, V]) Remove(key K) (V, bool) {
	shardIndex := m.shardFunc(key)
	m.locks[shardIndex].Lock()
	defer m.locks[shardIndex].Unlock()
	value, ok := m.mp[shardIndex][key]
	delete(m.mp[shardIndex], key)
	return value, ok
}

// Range iterates over the map and calls the given function for each key-value pair.
func (m *ShardedMap[K, V]) Range(f func(key K, value V)) {
	for i, shard := range m.mp {
		m.locks[i].RLock()
		for key, value := range shard {
			f(key, value)
		}
		m.locks[i].RUnlock()
	}
}

// Len returns the total number of elements in the map.
func (m *ShardedMap[K, V]) Len() int {
	count := 0
	for _, shard := range m.mp {
		count += len(shard)
	}
	return count
}

// Shards returns the number of shards in the map.
func (m *ShardedMap[K, V]) Shards() int {
	return len(m.locks)
}

// Shard returns the shard index for a given key.
func (m *ShardedMap[K, V]) Shard(key K) int {
	return m.shardFunc(key)
}
