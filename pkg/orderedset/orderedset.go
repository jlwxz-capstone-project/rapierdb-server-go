package orderedset

import (
	"fmt"
	"iter"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/orderedmap"
)

// OrderedSet 是一个保持插入顺序的 Set 实现
type OrderedSet[T comparable] struct {
	store *orderedmap.OrderedMap[T, struct{}]
}

// NewOrderedSet 创建一个新的 OrderedSet
func NewOrderedSet[T comparable]() *OrderedSet[T] {
	return &OrderedSet[T]{
		store: orderedmap.NewOrderedMap[T, struct{}](),
	}
}

// Add 添加元素到集合中
func (s *OrderedSet[T]) Add(value T) {
	s.store.Set(value, struct{}{})
}

// Remove 从集合中移除元素
func (s *OrderedSet[T]) Remove(value T) bool {
	return s.store.Delete(value)
}

// Contains 检查元素是否在集合中
func (s *OrderedSet[T]) Contains(value T) bool {
	_, exists := s.store.Get(value)
	return exists
}

// Len 返回集合中的元素数量
func (s *OrderedSet[T]) Len() int {
	return s.store.Len()
}

// Elements 返回集合中的所有元素，按照插入顺序
func (s *OrderedSet[T]) Elements() []T {
	return s.store.Keys()
}

// String 返回 OrderedSet 的字符串表示
func (s *OrderedSet[T]) String() string {
	var result string
	result = "OrderedSet{"
	first := true
	for value := range s.IterValues() {
		if !first {
			result += ", "
		}
		result += fmt.Sprintf("%v", value)
		first = false
	}
	result += "}"
	return result
}

// Union 返回两个集合的并集
func (s *OrderedSet[T]) Union(other *OrderedSet[T]) *OrderedSet[T] {
	result := NewOrderedSet[T]()
	for value := range s.IterValues() {
		result.Add(value)
	}
	for value := range other.IterValues() {
		result.Add(value)
	}
	return result
}

// Intersection 返回两个集合的交集
func (s *OrderedSet[T]) Intersection(other *OrderedSet[T]) *OrderedSet[T] {
	result := NewOrderedSet[T]()
	for value := range s.IterValues() {
		if other.Contains(value) {
			result.Add(value)
		}
	}
	return result
}

// Difference 返回两个集合的差集（在 s 中但不在 other 中的元素）
func (s *OrderedSet[T]) Difference(other *OrderedSet[T]) *OrderedSet[T] {
	result := NewOrderedSet[T]()
	for value := range s.IterValues() {
		if !other.Contains(value) {
			result.Add(value)
		}
	}
	return result
}

// IsSubset 检查 s 是否是 other 的子集
func (s *OrderedSet[T]) IsSubset(other *OrderedSet[T]) bool {
	if s.Len() > other.Len() {
		return false
	}
	isSubset := true
	for value := range s.IterValues() {
		if !other.Contains(value) {
			isSubset = false
			break
		}
	}
	return isSubset
}

// Clear 清空集合
func (s *OrderedSet[T]) Clear() {
	s.store = orderedmap.NewOrderedMap[T, struct{}]()
}

// IterValues 返回一个迭代器，按照插入顺序遍历所有元素
func (s *OrderedSet[T]) IterValues() iter.Seq[T] {
	return s.store.IterKeys()
}
