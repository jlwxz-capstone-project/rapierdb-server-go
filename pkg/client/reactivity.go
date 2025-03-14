package client

type RefWatcher[T any] func(new, old T)

type Ref[T any] struct {
	Value      T
	watchCount int
	watchers   map[int]RefWatcher[T]
}

func NewRef[T any](value T) *Ref[T] {
	return &Ref[T]{
		Value:      value,
		watchCount: 0,
		watchers:   make(map[int]RefWatcher[T]),
	}
}

func (r *Ref[T]) Set(value T) {
	old := r.Value
	r.Value = value
	for _, observer := range r.watchers {
		observer(value, old)
	}
}

func (r *Ref[T]) Get() T {
	return r.Value
}

func (r *Ref[T]) Watch(onChange RefWatcher[T]) func() {
	cnt := r.watchCount
	r.watchers[cnt] = onChange
	r.watchCount++
	return func() {
		delete(r.watchers, cnt)
	}
}
