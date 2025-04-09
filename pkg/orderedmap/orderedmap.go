package orderedmap

import (
	"container/list"
	"fmt"
	"iter"
)

// OrderedMap 是一个保持插入顺序的 Map 实现
type OrderedMap[K comparable, V any] struct {
	store    map[K]*list.Element
	list     *list.List
	keyOrder map[*list.Element]K
}

// entry 表示 Map 中的一个键值对
type entry[K comparable, V any] struct {
	key   K
	value V
}

// NewOrderedMap 创建一个新的 OrderedMap
func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		store:    make(map[K]*list.Element),
		list:     list.New(),
		keyOrder: make(map[*list.Element]K),
	}
}

// FromMap 从 map 创建一个新的 OrderedMap
func FromMap[K comparable, V any](m map[K]V) *OrderedMap[K, V] {
	ret := NewOrderedMap[K, V]()
	for k, v := range m {
		ret.Set(k, v)
	}
	return ret
}

// Set 设置键值对，如果键已存在则更新值
func (m *OrderedMap[K, V]) Set(key K, value V) {
	if element, exists := m.store[key]; exists {
		// 更新已存在的值
		element.Value.(*entry[K, V]).value = value
	} else {
		// 创建新的键值对
		entry := &entry[K, V]{key: key, value: value}
		element := m.list.PushBack(entry)
		m.store[key] = element
		m.keyOrder[element] = key
	}
}

// Get 获取键对应的值
func (m *OrderedMap[K, V]) Get(key K) (V, bool) {
	if element, exists := m.store[key]; exists {
		return element.Value.(*entry[K, V]).value, true
	}
	var zero V
	return zero, false
}

// MustGet 获取键对应的值
func (m *OrderedMap[K, V]) MustGet(key K) V {
	if element, exists := m.store[key]; exists {
		return element.Value.(*entry[K, V]).value
	}
	var zero V
	return zero
}

// Delete 删除键值对
func (m *OrderedMap[K, V]) Delete(key K) bool {
	if element, exists := m.store[key]; exists {
		m.list.Remove(element)
		delete(m.store, key)
		delete(m.keyOrder, element)
		return true
	}
	return false
}

// Len 返回 Map 中的键值对数量
func (m *OrderedMap[K, V]) Len() int {
	return len(m.store)
}

// Keys 返回所有键，按照插入顺序
func (m *OrderedMap[K, V]) Keys() []K {
	keys := make([]K, 0, m.Len())
	for e := m.list.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*entry[K, V]).key)
	}
	return keys
}

// Values 返回所有值，按照插入顺序
func (m *OrderedMap[K, V]) Values() []V {
	values := make([]V, 0, m.Len())
	for e := m.list.Front(); e != nil; e = e.Next() {
		values = append(values, e.Value.(*entry[K, V]).value)
	}
	return values
}

// String 返回 OrderedMap 的字符串表示
func (m *OrderedMap[K, V]) String() string {
	var result string
	result = "OrderedMap{"
	first := true
	for key, value := range m.store {
		if !first {
			result += ", "
		}
		result += fmt.Sprintf("%v: %v", key, value)
		first = false
	}
	result += "}"
	return result
}

// IterEntries 返回一个迭代器，按照插入顺序遍历所有键值对
func (m *OrderedMap[K, V]) IterEntries() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for e := m.list.Front(); e != nil; e = e.Next() {
			entry := e.Value.(*entry[K, V])
			if !yield(entry.key, entry.value) {
				return
			}
		}
	}
}

// IterKeys 返回一个迭代器，按照插入顺序遍历所有键
func (m *OrderedMap[K, V]) IterKeys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for e := m.list.Front(); e != nil; e = e.Next() {
			key := e.Value.(*entry[K, V]).key
			if !yield(key) {
				return
			}
		}
	}
}

// IterValues 返回一个迭代器，按照插入顺序遍历所有值
func (m *OrderedMap[K, V]) IterValues() iter.Seq[V] {
	return func(yield func(V) bool) {
		for e := m.list.Front(); e != nil; e = e.Next() {
			value := e.Value.(*entry[K, V]).value
			if !yield(value) {
				return
			}
		}
	}
}
