// Package memoization 是一个内部包，为文本区域提供简单的记忆化缓存功能
// 记忆化缓存（Memoization）是一种优化技术，用于存储昂贵函数调用的结果，避免重复计算
package memoization

import (
	"container/list"
	"crypto/sha256"
	"fmt"
	"sync"
)

// Hasher 是一个接口，要求实现Hash方法
// Hash方法应返回对象的哈希值的字符串表示形式
type Hasher interface {
	Hash() string
}

// entry 是一个结构体，用于存储键值对
// 它作为MemoCache的evictionList中的元素使用
type entry[T any] struct {
	key   string // 键（哈希值）
	value T      // 值
}

// MemoCache 是一个结构体，表示具有固定容量的缓存
// 它使用LRU（最近最少使用）淘汰策略，并且是线程安全的
type MemoCache[H Hasher, T any] struct {
	capacity      int                      // 缓存容量
	mutex         sync.Mutex               // 互斥锁，用于并发访问控制
	cache         map[string]*list.Element // 存储缓存结果的映射
	evictionList  *list.List               // 用于跟踪LRU淘汰顺序的列表
	hashableItems map[string]T             // 存储原始可哈希项的映射（可选）
}

// NewMemoCache 是一个函数，用于创建一个具有指定容量的新MemoCache
// 返回指向创建的MemoCache的指针
func NewMemoCache[H Hasher, T any](capacity int) *MemoCache[H, T] {
	return &MemoCache[H, T]{
		capacity:      capacity,                       // 缓存容量
		cache:         make(map[string]*list.Element), // 初始化缓存映射
		evictionList:  list.New(),                     // 初始化LRU淘汰列表
		hashableItems: make(map[string]T),             // 初始化可哈希项映射
	}
}

// Capacity 是一个方法，返回MemoCache的容量
func (m *MemoCache[H, T]) Capacity() int {
	return m.capacity
}

// Size 是一个方法，返回MemoCache的当前大小
// 即当前存储在缓存中的项目数量
func (m *MemoCache[H, T]) Size() int {
	m.mutex.Lock()              // 加锁，确保并发安全
	defer m.mutex.Unlock()      // 函数返回时解锁
	return m.evictionList.Len() // 返回LRU列表的长度，即缓存中的项目数量
}

// Get 是一个方法，返回与给定可哈希项关联的值
// 如果没有对应的值，返回零值和false
func (m *MemoCache[H, T]) Get(h H) (T, bool) {
	m.mutex.Lock()         // 加锁，确保并发安全
	defer m.mutex.Unlock() // 函数返回时解锁

	hashedKey := h.Hash() // 获取可哈希项的哈希值
	// 检查缓存中是否存在该哈希值
	if element, found := m.cache[hashedKey]; found {
		m.evictionList.MoveToFront(element)          // 将元素移到列表头部，表示最近使用过
		return element.Value.(*entry[T]).value, true // 返回缓存的值和true
	}
	var result T
	return result, false // 缓存未命中，返回零值和false
}

// Set 是一个方法，为给定的可哈希项设置值
// 如果缓存已满，会先淘汰最近最少使用的项目，然后再添加新项目
func (m *MemoCache[H, T]) Set(h H, value T) {
	m.mutex.Lock()         // 加锁，确保并发安全
	defer m.mutex.Unlock() // 函数返回时解锁

	hashedKey := h.Hash() // 获取可哈希项的哈希值
	// 检查缓存中是否已存在该哈希值
	if element, found := m.cache[hashedKey]; found {
		m.evictionList.MoveToFront(element)     // 将元素移到列表头部，表示最近使用过
		element.Value.(*entry[T]).value = value // 更新缓存的值
		return                                  // 缓存已存在，更新后返回
	}

	// 检查缓存是否已满
	if m.evictionList.Len() >= m.capacity {
		// 淘汰最近最少使用的项目
		toEvict := m.evictionList.Back() // 获取列表尾部的元素（最近最少使用）
		if toEvict != nil {
			evictedEntry := m.evictionList.Remove(toEvict).(*entry[T]) // 从列表中移除
			delete(m.cache, evictedEntry.key)                          // 从缓存映射中删除
			delete(m.hashableItems, evictedEntry.key)                  // 从可哈希项映射中删除（如果启用）
		}
	}

	// 将新值添加到缓存和LRU列表
	newEntry := &entry[T]{
		key:   hashedKey, // 哈希值作为键
		value: value,     // 要缓存的值
	}
	element := m.evictionList.PushFront(newEntry) // 将新元素添加到列表头部
	m.cache[hashedKey] = element                  // 将元素添加到缓存映射
	m.hashableItems[hashedKey] = value            // 将原始值添加到可哈希项映射（如果启用）
}

// HString 是一个类型，为字符串实现了Hasher接口
type HString string

// Hash 是一个方法，返回字符串的哈希值
func (h HString) Hash() string {
	// 使用SHA256算法计算哈希值，并返回十六进制字符串
	return fmt.Sprintf("%x", sha256.Sum256([]byte(h)))
}

// HInt 是一个类型，为整数实现了Hasher接口
type HInt int

// Hash 是一个方法，返回整数的哈希值
func (h HInt) Hash() string {
	// 将整数转换为字符串，然后使用SHA256算法计算哈希值
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d", h))))
}
