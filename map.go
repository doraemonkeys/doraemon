package doraemon

import (
	"fmt"
	"hash"
	"hash/fnv"
	"runtime"
	"sync"
	"unsafe"
)

func numHash[
	K ~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64](key K, h hash.Hash32) uint32 {
	ptr := &key
	_, _ = h.Write(unsafe.Slice((*byte)(unsafe.Pointer(ptr)), unsafe.Sizeof(key)))
	runtime.KeepAlive(ptr)
	return h.Sum32()
}

func stringHash(key string, h hash.Hash32) uint32 {
	_, _ = h.Write([]byte(key))
	return h.Sum32()
}

func boolHash(key bool, h hash.Hash32) uint32 {
	if key {
		_, _ = h.Write([]byte{1})
	} else {
		_, _ = h.Write([]byte{0})
	}
	return h.Sum32()
}

func defaultHash[K comparable](key K, h hash.Hash32) uint32 {
	_, _ = fmt.Fprint(h, key)
	return h.Sum32()
}

// ShardedMap is a concurrent map sharded across multiple locks.
type ShardedMap[K comparable, V any] struct {
	locks     []sync.RWMutex  // Locks for each shard
	mp        []map[K]V       // Array of maps, each representing a shard
	shardFunc func(key K) int // Function to determine shard index for a key
}

// DefaultHashCalc returns a default sharding function using FNV-1a hash.
func DefaultHashCalc[K comparable](shardCount int) func(key K) int {
	return func(key K) int {
		var h uint32
		switch k := any(key).(type) {
		case int:
			h = numHash(k, fnv.New32a())
		case int8:
			h = numHash(k, fnv.New32a())
		case int16:
			h = numHash(k, fnv.New32a())
		case int32:
			h = numHash(k, fnv.New32a())
		case int64:
			h = numHash(k, fnv.New32a())
		case uint:
			h = numHash(k, fnv.New32a())
		case uint8:
			h = numHash(k, fnv.New32a())
		case uint16:
			h = numHash(k, fnv.New32a())
		case uint32:
			h = numHash(k, fnv.New32a())
		case uint64:
			h = numHash(k, fnv.New32a())
		case float32:
			h = numHash(k, fnv.New32a())
		case float64:
			h = numHash(k, fnv.New32a())
		case string:
			h = stringHash(k, fnv.New32a())
		case bool:
			h = boolHash(k, fnv.New32a())
		default:
			h = defaultHash(k, fnv.New32a())
		}
		return int(h % uint32(shardCount))
	}
}

// NewMap creates a new ShardedMap with the specified shard count and sharding function.
// If calcFunc is nil, it uses DefaultCalc.
func NewMap[K comparable, V any](shardCount int, calcFunc func(key K) int) *ShardedMap[K, V] {
	if calcFunc == nil {
		calcFunc = DefaultHashCalc[K](shardCount)
	}
	locks := make([]sync.RWMutex, shardCount)
	mp := make([]map[K]V, shardCount)
	for i := range shardCount {
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

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m *ShardedMap[K, V]) Range(f func(key K, value V) bool) {
	for i, shard := range m.mp {
		m.locks[i].RLock()
		for key, value := range shard {
			if !f(key, value) {
				break
			}
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

// ShardIndex returns the shard index for a given key.
func (m *ShardedMap[K, V]) ShardIndex(key K) int {
	return m.shardFunc(key)
}
