package memoization

import (
	"encoding/binary"
	"fmt"
	"os"
	"testing"
)

// actionType 是一个枚举类型，表示缓存操作类型
type actionType int

const (
	set actionType = iota // 设置操作
	get                   // 获取操作
)

// cacheAction 是一个结构体，表示缓存操作
// 用于定义测试用例中的操作步骤
type cacheAction struct {
	actionType    actionType  // 操作类型（set或get）
	key           HString     // 操作的键
	value         interface{} // 要设置的值（仅set操作使用）
	expectedValue interface{} // 预期的返回值（仅get操作使用）
}

// testCase 是一个结构体，表示测试用例
type testCase struct {
	name     string        // 测试用例名称
	capacity int           // 缓存容量
	actions  []cacheAction // 测试操作步骤
}

// TestCache 是一个测试函数，测试MemoCache的各种功能
func TestCache(t *testing.T) {
	tests := []testCase{
		{
			name:     "TestNewMemoCache", // 测试新创建的缓存
			capacity: 5,                  // 缓存容量为5
			actions: []cacheAction{
				{actionType: get, expectedValue: nil}, // 获取不存在的键，预期返回nil
			},
		},
		{
			name:     "TestSetAndGet", // 测试设置和获取操作
			capacity: 10,              // 缓存容量为10
			actions: []cacheAction{
				{actionType: set, key: "key1", value: "value1"},              // 设置key1为value1
				{actionType: get, key: "key1", expectedValue: "value1"},      // 获取key1，预期返回value1
				{actionType: set, key: "key1", value: "newValue1"},           // 更新key1为newValue1
				{actionType: get, key: "key1", expectedValue: "newValue1"},   // 获取key1，预期返回newValue1
				{actionType: get, key: "nonExistentKey", expectedValue: nil}, // 获取不存在的键，预期返回nil
				{actionType: set, key: "nilKey", value: ""},                  // 设置nilKey为空字符串
				{actionType: get, key: "nilKey", expectedValue: ""},          // 获取nilKey，预期返回空字符串
				{actionType: set, key: "keyA", value: "valueA"},              // 设置keyA为valueA
				{actionType: set, key: "keyB", value: "valueB"},              // 设置keyB为valueB
				{actionType: get, key: "keyA", expectedValue: "valueA"},      // 获取keyA，预期返回valueA
				{actionType: get, key: "keyB", expectedValue: "valueB"},      // 获取keyB，预期返回valueB
			},
		},
		{
			name:     "TestSetNilValue", // 测试设置nil值
			capacity: 10,                // 缓存容量为10
			actions: []cacheAction{
				{actionType: set, key: HString("nilKey"), value: nil},         // 设置nilKey为nil
				{actionType: get, key: HString("nilKey"), expectedValue: nil}, // 获取nilKey，预期返回nil
			},
		},
		{
			name:     "TestGetAfterEviction", // 测试淘汰后的获取操作
			capacity: 2,                      // 缓存容量为2
			actions: []cacheAction{
				{actionType: set, key: HString("1"), value: 1},           // 设置key1为1
				{actionType: set, key: HString("2"), value: 2},           // 设置key2为2
				{actionType: set, key: HString("3"), value: 3},           // 设置key3为3（此时缓存已满，会淘汰key1）
				{actionType: get, key: HString("1"), expectedValue: nil}, // 获取key1，预期返回nil（已被淘汰）
				{actionType: get, key: HString("2"), expectedValue: 2},   // 获取key2，预期返回2
			},
		},
		{
			name:     "TestGetAfterLRU", // 测试LRU（最近最少使用）策略
			capacity: 2,                 // 缓存容量为2
			actions: []cacheAction{
				{actionType: set, key: HString("1"), value: 1},           // 设置key1为1
				{actionType: set, key: HString("2"), value: 2},           // 设置key2为2
				{actionType: get, key: HString("1"), expectedValue: 1},   // 获取key1，预期返回1（更新使用时间）
				{actionType: set, key: HString("3"), value: 3},           // 设置key3为3（此时缓存已满，会淘汰最近最少使用的key2）
				{actionType: get, key: HString("1"), expectedValue: 1},   // 获取key1，预期返回1
				{actionType: get, key: HString("3"), expectedValue: 3},   // 获取key3，预期返回3
				{actionType: get, key: HString("2"), expectedValue: nil}, // 获取key2，预期返回nil（已被淘汰）
			},
		},
		{
			name:     "TestLRU_Capacity3", // 测试容量为3的LRU策略
			capacity: 3,                   // 缓存容量为3
			actions: []cacheAction{
				{actionType: set, key: HString("1"), value: 1},           // 设置key1为1
				{actionType: set, key: HString("2"), value: 2},           // 设置key2为2
				{actionType: set, key: HString("3"), value: 3},           // 设置key3为3
				{actionType: get, key: HString("1"), expectedValue: 1},   // 获取key1，预期返回1（更新使用时间）
				{actionType: set, key: HString("4"), value: 4},           // 设置key4为4（此时缓存已满，会淘汰最近最少使用的key2）
				{actionType: get, key: HString("2"), expectedValue: nil}, // 获取key2，预期返回nil（已被淘汰）
				{actionType: get, key: HString("1"), expectedValue: 1},   // 获取key1，预期返回1
				{actionType: get, key: HString("3"), expectedValue: 3},   // 获取key3，预期返回3
				{actionType: get, key: HString("4"), expectedValue: 4},   // 获取key4，预期返回4
			},
		},
		// 测试不同访问模式下的LRU行为
		{
			name:     "TestLRU_VaryingAccesses", // 测试不同访问模式下的LRU行为
			capacity: 3,                         // 缓存容量为3
			actions: []cacheAction{
				{actionType: set, key: HString("1"), value: 1},           // 设置key1为1
				{actionType: set, key: HString("2"), value: 2},           // 设置key2为2
				{actionType: set, key: HString("3"), value: 3},           // 设置key3为3
				{actionType: get, key: HString("1"), expectedValue: 1},   // 获取key1，预期返回1（更新使用时间）
				{actionType: get, key: HString("2"), expectedValue: 2},   // 获取key2，预期返回2（更新使用时间）
				{actionType: set, key: HString("4"), value: 4},           // 设置key4为4（此时缓存已满，会淘汰最近最少使用的key3）
				{actionType: get, key: HString("3"), expectedValue: nil}, // 获取key3，预期返回nil（已被淘汰）
				{actionType: get, key: HString("1"), expectedValue: 1},   // 获取key1，预期返回1
				{actionType: get, key: HString("2"), expectedValue: 2},   // 获取key2，预期返回2
				{actionType: get, key: HString("4"), expectedValue: 4},   // 获取key4，预期返回4
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewMemoCache[HString, interface{}](tt.capacity)
			for _, action := range tt.actions {
				switch action.actionType {
				case set:
					cache.Set(action.key, action.value)
				case get:
					if got, _ := cache.Get(action.key); got != action.expectedValue {
						t.Errorf("Get() = %v, want %v", got, action.expectedValue)
					}
				}
			}
		})
	}
}

func FuzzCache(f *testing.F) {
	// Define some seed values for initial scenarios
	for _, seed := range [][]byte{
		[]byte("7\x010\x0000000020"),
		{0, 0, 0, 0}, // Set key 0 to 0
		{1, 0, 0, 1}, // Set key 0 to 1
		{2, 0},       // Get key 0
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, in []byte) {
		if len(in) < 1 {
			t.Skip() // Skip the test if the input is less than 1 byte
		}

		cache := NewMemoCache[HInt, int](10) // Initialize a cache with the initial size

		expectedValues := make(map[HInt]int) // Map to store expected key-value pairs
		accessOrder := make([]HInt, 0)       // Slice to store the order of keys accessed

		for i := 0; i < len(in); {
			opCode := in[i] % 4 // Determine the operation: Set, Get, or Reset (added case for Reset)
			i++

			switch opCode {
			case 0, 1: // Set operation
				if i+3 > len(in) {
					t.Skip() // Not enough input to continue, so skip
				}

				key := HInt(binary.BigEndian.Uint16(in[i : i+2]))
				value := int(in[i+2])
				i += 3

				// If the key is already in accessOrder, we remove it and append it again later
				for index, accessedKey := range accessOrder {
					if accessedKey == key {
						accessOrder = append(accessOrder[:index], accessOrder[index+1:]...)
						break
					}
				}

				cache.Set(key, value) // Set the value in the cache
				expectedValues[key] = value
				accessOrder = append(accessOrder, key) // Add the key to the access order slice

				// If we exceeded the cache size, we need to evict the least recently used item
				if len(accessOrder) > cache.Capacity() {
					evictedKey := accessOrder[0]
					accessOrder = accessOrder[1:]
					delete(expectedValues, evictedKey) // Remove the evicted key from expected values
				}

			case 2: // Get operation
				if i >= len(in) {
					t.Skip() // Not enough input to continue, so skip
				}

				key := HInt(in[i])
				i++

				expectedValue, ok := expectedValues[key]
				if !ok {
					// If the key is not found, it means it was either evicted or never added
					expectedValue = 0 // The zero value, depends on your cache implementation
				} else {
					// If the key was accessed, move it to the end of the accessOrder to represent recent use
					for index, accessedKey := range accessOrder {
						if accessedKey == key {
							accessOrder = append(accessOrder[:index], accessOrder[index+1:]...)
							accessOrder = append(accessOrder, key)
							break
						}
					}
				}

				if got, _ := cache.Get(key); got != expectedValue {
					fmt.Fprintf(os.Stderr, "cache: capacity: %d, hashable: %v, cache: %v\n", cache.capacity, cache.hashableItems, cache.cache)
					t.Fatalf("Get(%v) = %v, want %v", key, got, expectedValue) // The values do not match
				}
			case 3: // Reset operation
				if i >= len(in) {
					t.Skip() // Not enough input to continue, so skip
				}

				newCacheSize := int(in[i]) // Read the new cache size from the input
				i++

				if newCacheSize == 0 {
					t.Skip() // If the size is zero, we skip this test
				}

				// Create a new cache with the specified size
				cache = NewMemoCache[HInt, int](newCacheSize)

				// clear and reinitialize the expected values
				expectedValues = make(map[HInt]int)
				accessOrder = make([]HInt, 0)
			}
		}
	})
}
