package doraemon

import (
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMap(t *testing.T) {
	t.Run("default calc", func(t *testing.T) {
		m := NewMap[string, int](4, nil)
		assert.NotNil(t, m)
		assert.Equal(t, 4, len(m.locks))
		assert.Equal(t, 4, len(m.mp))
		assert.NotNil(t, m.shardFunc)
	})

	t.Run("custom calc", func(t *testing.T) {
		calcFunc := func(key string) int {
			return len(key) % 2
		}
		m := NewMap[string, int](2, calcFunc)
		assert.NotNil(t, m)
		assert.Equal(t, 2, len(m.locks))
		assert.Equal(t, 2, len(m.mp))
	})
}

func TestGet(t *testing.T) {
	m := NewMap[string, int](2, nil)
	m.Set("key1", 10)
	m.Set("key2", 20)

	t.Run("key exists", func(t *testing.T) {
		val, ok := m.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, 10, val)
	})

	t.Run("key does not exist", func(t *testing.T) {
		val, ok := m.Get("key3")
		assert.False(t, ok)
		assert.Equal(t, 0, val)
	})
}

func TestLoadOrStore(t *testing.T) {
	m := NewMap[string, int](2, nil)

	t.Run("key does not exist", func(t *testing.T) {
		val := m.LoadOrStore("key1", 10)
		assert.Equal(t, 10, val)
		val, ok := m.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, 10, val)
	})

	t.Run("key exists", func(t *testing.T) {
		val := m.LoadOrStore("key1", 20)
		assert.Equal(t, 10, val) // original value should be returned
		val, ok := m.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, 10, val)
	})
}

func TestSet(t *testing.T) {
	m := NewMap[string, int](2, nil)
	m.Set("key1", 10)
	val, ok := m.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, 10, val)
	m.Set("key1", 20) //overwrite
	val, ok = m.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, 20, val)
}

func TestStoreIfAbsent(t *testing.T) {
	m := NewMap[string, int](2, nil)

	t.Run("key not present", func(t *testing.T) {
		m.StoreIfAbsent("key1", 10)
		val, ok := m.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, 10, val)
	})

	t.Run("key already present", func(t *testing.T) {
		m.StoreIfAbsent("key1", 20)
		val, ok := m.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, 10, val) // should not change if present
	})
}

func TestStoreIfPresent(t *testing.T) {
	m := NewMap[string, int](2, nil)

	t.Run("key present", func(t *testing.T) {
		m.Set("key1", 10)
		m.StoreIfPresent("key1", 20)
		val, ok := m.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, 20, val)
	})
	t.Run("key not present", func(t *testing.T) {
		m.StoreIfPresent("key2", 20)
		val, ok := m.Get("key2")
		assert.False(t, ok)
		assert.Equal(t, 0, val) //should not change
	})
}

func TestDelete(t *testing.T) {
	m := NewMap[string, int](2, nil)
	m.Set("key1", 10)

	t.Run("key exists", func(t *testing.T) {
		m.Delete("key1")
		_, ok := m.Get("key1")
		assert.False(t, ok)
	})
	t.Run("key does not exist", func(t *testing.T) {
		m.Delete("key2") // should not panic if key not present
		_, ok := m.Get("key2")
		assert.False(t, ok)
	})
}

func TestContains(t *testing.T) {
	m := NewMap[string, int](2, nil)
	m.Set("key1", 10)

	t.Run("key exists", func(t *testing.T) {
		assert.True(t, m.Contains("key1"))
	})

	t.Run("key does not exist", func(t *testing.T) {
		assert.False(t, m.Contains("key2"))
	})
}

func TestRemove(t *testing.T) {
	m := NewMap[string, int](2, nil)
	m.Set("key1", 10)

	t.Run("key exists", func(t *testing.T) {
		val, ok := m.Remove("key1")
		assert.True(t, ok)
		assert.Equal(t, 10, val)
		_, ok = m.Get("key1")
		assert.False(t, ok)
	})
	t.Run("key does not exist", func(t *testing.T) {
		val, ok := m.Remove("key2")
		assert.False(t, ok)
		assert.Equal(t, 0, val)
	})
}

func TestRange(t *testing.T) {
	m := NewMap[string, int](2, nil)
	m.Set("key1", 10)
	m.Set("key2", 20)

	var count int
	var sum int
	m.Range(func(key string, value int) bool {
		count++
		sum += value
		return true
	})

	assert.Equal(t, 2, count)
	assert.Equal(t, 30, sum)
}

func TestLen(t *testing.T) {
	m := NewMap[string, int](2, nil)
	assert.Equal(t, 0, m.Len())

	m.Set("key1", 10)
	assert.Equal(t, 1, m.Len())

	m.Set("key2", 20)
	assert.Equal(t, 2, m.Len())

	m.Delete("key1")
	assert.Equal(t, 1, m.Len())
}

func TestShards(t *testing.T) {
	m := NewMap[string, int](4, nil)
	assert.Equal(t, 4, m.Shards())

	m2 := NewMap[string, int](2, nil)
	assert.Equal(t, 2, m2.Shards())
}

func TestShard(t *testing.T) {
	m := NewMap[string, int](4, nil)
	// The default shard func depends on the hash of the key.
	//  It can be any of the shard, so we cannot check the exact number of shard.
	// Just check shard index is within the valid range

	assert.True(t, m.ShardIndex("key1") >= 0)
	assert.True(t, m.ShardIndex("key1") < m.Shards())

	assert.True(t, m.ShardIndex("key2") >= 0)
	assert.True(t, m.ShardIndex("key2") < m.Shards())
}

func TestConcurrency(t *testing.T) {
	m := NewMap[int, int](10, nil)
	var wg sync.WaitGroup
	const numGoRoutines = 100
	const numOperations = 100

	for i := 0; i < numGoRoutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := (routineID * numOperations) + j
				m.Set(key, key*2)
				val, ok := m.Get(key)
				assert.True(t, ok)
				assert.Equal(t, key*2, val)
			}
		}(i)
	}

	wg.Wait()
	assert.Equal(t, numGoRoutines*numOperations, m.Len())
}

func TestConcurrencyWithRange(t *testing.T) {
	m := NewMap[int, int](10, nil)
	var wg sync.WaitGroup

	// Add elements concurrently
	const numGoRoutines = 100
	const numOperations = 100

	for i := 0; i < numGoRoutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := (routineID * numOperations) + j
				m.Set(key, key*2)
			}
		}(i)
	}
	wg.Wait()

	// Check the length
	assert.Equal(t, numGoRoutines*numOperations, m.Len())

	// Range and verify
	var rangeWg sync.WaitGroup
	rangeWg.Add(1)
	go func() {
		defer rangeWg.Done()
		m.Range(func(key int, value int) bool {
			expectedValue := key * 2
			assert.Equal(t, expectedValue, value, fmt.Sprintf("Unexpected value for key %d", key))
			return true
		})
	}()
	rangeWg.Wait()
}

func TestNewMap2(t *testing.T) {
	sm := NewMap[string, int](8, nil)
	if sm.Shards() != 8 {
		t.Errorf("Expected 8 shards, got %d", sm.Shards())
	}
}

func TestGet2(t *testing.T) {
	sm := NewMap[string, int](4, nil)
	sm.Set("key1", 100)

	value, ok := sm.Get("key1")
	if !ok || value != 100 {
		t.Errorf("Expected (100, true), got (%d, %v)", value, ok)
	}

	_, ok = sm.Get("non-existent")
	if ok {
		t.Error("Expected false for non-existent key")
	}
}

func TestLoadOrStore2(t *testing.T) {
	sm := NewMap[string, int](4, nil)

	value := sm.LoadOrStore("key1", 100)
	if value != 100 {
		t.Errorf("Expected 100, got %d", value)
	}

	value = sm.LoadOrStore("key1", 200)
	if value != 100 {
		t.Errorf("Expected 100, got %d", value)
	}
}

func TestSet2(t *testing.T) {
	sm := NewMap[string, int](4, nil)

	sm.Set("key1", 100)
	value, ok := sm.Get("key1")
	if !ok || value != 100 {
		t.Errorf("Expected (100, true), got (%d, %v)", value, ok)
	}

	sm.Set("key1", 200)
	value, ok = sm.Get("key1")
	if !ok || value != 200 {
		t.Errorf("Expected (200, true), got (%d, %v)", value, ok)
	}
}

func TestStoreIfAbsent2(t *testing.T) {
	sm := NewMap[string, int](4, nil)

	sm.StoreIfAbsent("key1", 100)
	value, ok := sm.Get("key1")
	if !ok || value != 100 {
		t.Errorf("Expected (100, true), got (%d, %v)", value, ok)
	}

	sm.StoreIfAbsent("key1", 200)
	value, ok = sm.Get("key1")
	if !ok || value != 100 {
		t.Errorf("Expected (100, true), got (%d, %v)", value, ok)
	}
}

func TestStoreIfPresent2(t *testing.T) {
	sm := NewMap[string, int](4, nil)

	sm.StoreIfPresent("key1", 100)
	_, ok := sm.Get("key1")
	if ok {
		t.Error("Expected key not to be present")
	}

	sm.Set("key1", 100)
	sm.StoreIfPresent("key1", 200)
	value, ok := sm.Get("key1")
	if !ok || value != 200 {
		t.Errorf("Expected (200, true), got (%d, %v)", value, ok)
	}
}

func TestDelete2(t *testing.T) {
	sm := NewMap[string, int](4, nil)

	sm.Set("key1", 100)
	sm.Delete("key1")

	_, ok := sm.Get("key1")
	if ok {
		t.Error("Expected key to be deleted")
	}
}

func TestContains2(t *testing.T) {
	sm := NewMap[string, int](4, nil)

	if sm.Contains("key1") {
		t.Error("Expected false for non-existent key")
	}

	sm.Set("key1", 100)
	if !sm.Contains("key1") {
		t.Error("Expected true for existing key")
	}
}

func TestRemove2(t *testing.T) {
	sm := NewMap[string, int](4, nil)

	sm.Set("key1", 100)
	value, ok := sm.Remove("key1")
	if !ok || value != 100 {
		t.Errorf("Expected (100, true), got (%d, %v)", value, ok)
	}

	_, ok = sm.Get("key1")
	if ok {
		t.Error("Expected key to be removed")
	}

	_, ok = sm.Remove("non-existent")
	if ok {
		t.Error("Expected false for non-existent key")
	}
}

func TestRange2(t *testing.T) {
	sm := NewMap[string, int](4, nil)
	expected := map[string]int{
		"key1": 100,
		"key2": 200,
		"key3": 300,
	}

	for k, v := range expected {
		sm.Set(k, v)
	}

	count := 0
	sm.Range(func(key string, value int) bool {
		count++
		if expected[key] != value {
			t.Errorf("Expected %d for key %s, got %d", expected[key], key, value)
		}
		return true
	})

	if count != len(expected) {
		t.Errorf("Expected %d iterations, got %d", len(expected), count)
	}
}

func TestIterate(t *testing.T) {
	sm := NewMap[string, int](4, nil)
	expected := map[string]int{
		"key1": 100,
		"key2": 200,
		"key3": 300,
	}

	for k, v := range expected {
		sm.Set(k, v)
	}

	count := 0
	for k, v := range sm.Range {
		count++
		if expected[k] != v {
			t.Errorf("Expected %d for key %s, got %d", expected[k], k, v)
		}
	}

	if count != len(expected) {
		t.Errorf("Expected %d iterations, got %d", len(expected), count)
	}

}

func TestLen2(t *testing.T) {
	sm := NewMap[string, int](4, nil)

	if sm.Len() != 0 {
		t.Errorf("Expected length 0, got %d", sm.Len())
	}

	sm.Set("key1", 100)
	sm.Set("key2", 200)

	if sm.Len() != 2 {
		t.Errorf("Expected length 2, got %d", sm.Len())
	}
}

func TestShard2(t *testing.T) {
	sm := NewMap[string, int](4, nil)

	shard1 := sm.ShardIndex("key1")
	shard2 := sm.ShardIndex("key2")

	if shard1 == shard2 {
		t.Error("Expected different shards for different keys")
	}

	if shard1 < 0 || shard1 >= 4 || shard2 < 0 || shard2 >= 4 {
		t.Error("Shard index out of range")
	}
}

func TestConcurrency2(t *testing.T) {
	sm := NewMap[int, int](8, nil)
	var wg sync.WaitGroup
	numOps := 1000

	wg.Add(4)

	// Writer goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < numOps; i++ {
			sm.Set(i, i*10)
		}
	}()

	// Reader goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < numOps; i++ {
			sm.Get(i)
		}
	}()

	// LoadOrStore goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < numOps; i++ {
			sm.LoadOrStore(i, i*20)
		}
	}()

	// Delete goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < numOps; i++ {
			if i%2 == 0 {
				sm.Delete(i)
			}
		}
	}()

	wg.Wait()

}

func TestConcurrencyWrite(t *testing.T) {
	sm := NewMap[int, int](8, nil)
	var wg sync.WaitGroup
	numOps := 1000

	wg.Add(numOps)
	for i := 0; i < numOps; i++ {
		go func(key, value int) {
			defer wg.Done()
			sm.Set(key, value)
		}(i, i*10)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numOps; i++ {
			_, _ = sm.Get(i)
		}
	}()
	wg.Wait()

	for i := 0; i < numOps; i++ {
		val, ok := sm.Get(i)
		if !ok {
			t.Errorf("Expected key %d to exist", i)
		}
		if val != i*10 {
			t.Errorf("Expected value for key %d to be %d, got %d", i, i*10, val)
		}
	}
}

func TestCustomShardingFunction(t *testing.T) {
	customShardFunc := func(key int) int {
		return key % 4
	}

	sm := NewMap[int, string](4, customShardFunc)

	for i := 0; i < 100; i++ {
		sm.Set(i, fmt.Sprintf("value%d", i))
		if sm.ShardIndex(i) != i%4 {
			t.Errorf("Expected shard %d for key %d, got %d", i%4, i, sm.ShardIndex(i))
		}
	}
}

func BenchmarkSet(b *testing.B) {
	sm := NewMap[int, int](8, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.Set(i, i)
	}
}

func BenchmarkGet(b *testing.B) {
	sm := NewMap[int, int](8, nil)
	for i := 0; i < 1000000; i++ {
		sm.Set(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.Get(i % 1000000)
	}
}

func BenchmarkLoadOrStore(b *testing.B) {
	sm := NewMap[int, int](8, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.LoadOrStore(i%100000, i)
	}
}

const (
	shardCount = 32
	keyCount   = 100000
)

// BenchmarkShardedMapGet benchmarks the Get operation of ShardedMap.
func BenchmarkShardedMapGet(b *testing.B) {
	m := NewMap[string, int](shardCount, nil)
	for i := 0; i < keyCount; i++ {
		m.Set(strconv.Itoa(i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < keyCount; i++ {
				_, _ = m.Get(strconv.Itoa(i))
			}
		}
	})
}

// BenchmarkBuiltinMapGet benchmarks the Get operation of the built-in map with RWMutex.
func BenchmarkBuiltinMapGet(b *testing.B) {
	var mu sync.RWMutex
	m := make(map[string]int)
	for i := 0; i < keyCount; i++ {
		m[strconv.Itoa(i)] = i
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < keyCount; i++ {
				mu.RLock()
				_ = m[strconv.Itoa(i)]
				mu.RUnlock()
			}
		}
	})
}

// BenchmarkShardedMapSet benchmarks the Set operation of ShardedMap.
func BenchmarkShardedMapSet(b *testing.B) {
	m := NewMap[string, int](shardCount, nil)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Set(strconv.Itoa(i), i)
			i++
		}
	})
}

// BenchmarkBuiltinMapSet benchmarks the Set operation of the built-in map with Mutex.
func BenchmarkBuiltinMapSet(b *testing.B) {
	var mu sync.Mutex
	m := make(map[string]int)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			mu.Lock()
			m[strconv.Itoa(i)] = i
			mu.Unlock()
			i++
		}
	})
}

// BenchmarkShardedMapLoadOrStore benchmarks the LoadOrStore operation of ShardedMap.
func BenchmarkShardedMapLoadOrStore(b *testing.B) {
	m := NewMap[string, int](shardCount, nil)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = m.LoadOrStore(strconv.Itoa(i), i)
			i++
		}
	})
}

// BenchmarkBuiltinMapLoadOrStore benchmarks the LoadOrStore operation of the built-in map with Mutex.
func BenchmarkBuiltinMapLoadOrStore(b *testing.B) {
	var mu sync.Mutex
	m := make(map[string]int)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := strconv.Itoa(i)
			mu.Lock()
			if _, ok := m[key]; !ok {
				m[key] = i
			}
			mu.Unlock()
			i++
		}
	})
}

// BenchmarkShardedMapDelete benchmarks the Delete operation of ShardedMap.
func BenchmarkShardedMapDelete(b *testing.B) {
	m := NewMap[string, int](shardCount, nil)
	for i := 0; i < keyCount; i++ {
		m.Set(strconv.Itoa(i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < keyCount; i++ {
				m.Delete(strconv.Itoa(i))
			}
		}
	})
}

// BenchmarkBuiltinMapDelete benchmarks the Delete operation of the built-in map with Mutex.
func BenchmarkBuiltinMapDelete(b *testing.B) {
	var mu sync.Mutex
	m := make(map[string]int)
	for i := 0; i < keyCount; i++ {
		m[strconv.Itoa(i)] = i
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < keyCount; i++ {
				mu.Lock()
				delete(m, strconv.Itoa(i))
				mu.Unlock()
			}
		}
	})
}

// BenchmarkShardedMapRange benchmarks the Range operation of ShardedMap.
func BenchmarkShardedMapRange(b *testing.B) {
	m := NewMap[string, int](shardCount, nil)
	for i := 0; i < keyCount; i++ {
		m.Set(strconv.Itoa(i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Range(func(key string, value int) bool {
				_ = fmt.Sprint(key, value)
				return true
			})
		}
	})
}

// BenchmarkBuiltinMapRange benchmarks the Range operation of the built-in map with RWMutex.
func BenchmarkBuiltinMapRange(b *testing.B) {
	var mu sync.RWMutex
	m := make(map[string]int)
	for i := 0; i < keyCount; i++ {
		m[strconv.Itoa(i)] = i
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.RLock()
			for k, v := range m {
				_ = fmt.Sprint(k, v)
			}
			mu.RUnlock()
		}
	})
}

// BenchmarkShardedMapMixed benchmarks mixed read and write operations of ShardedMap.
func BenchmarkShardedMapMixed(b *testing.B) {
	m := NewMap[string, int](shardCount, nil)
	for i := 0; i < keyCount; i++ {
		m.Set(strconv.Itoa(i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := keyCount
		for pb.Next() {
			if i%2 == 0 {
				m.Set(strconv.Itoa(i), i)
			} else {
				_, _ = m.Get(strconv.Itoa(i % keyCount))
			}
			i++
		}
	})
}

// BenchmarkBuiltinMapMixed benchmarks mixed read and write operations of the built-in map with RWMutex and Mutex.
func BenchmarkBuiltinMapMixed(b *testing.B) {
	var mu sync.Mutex
	m := make(map[string]int)
	for i := 0; i < keyCount; i++ {
		m[strconv.Itoa(i)] = i
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := keyCount
		for pb.Next() {
			if i%2 == 0 {
				mu.Lock()
				m[strconv.Itoa(i)] = i
				mu.Unlock()
			} else {
				mu.Lock()
				_ = m[strconv.Itoa(i%keyCount)]
				mu.Unlock()
			}
			i++
		}
	})
}

// BenchmarkBuiltinMapMixed benchmarks mixed read and write operations of the built-in map with RWMutex and Mutex.
func BenchmarkBuiltinMapMixed2(b *testing.B) {
	var mu sync.RWMutex
	m := make(map[string]int)
	for i := 0; i < keyCount; i++ {
		m[strconv.Itoa(i)] = i
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := keyCount
		for pb.Next() {
			if i%2 == 0 {
				mu.Lock()
				m[strconv.Itoa(i)] = i
				mu.Unlock()
			} else {
				mu.RLock()
				_ = m[strconv.Itoa(i%keyCount)]
				mu.RUnlock()
			}
			i++
		}
	})
}
