package memory

import "sync"

// MapPool provides a pool of map[string]interface{} to reduce GC pressure.
// It is safe for concurrent use.
type MapPool struct {
	pool sync.Pool
}

var globalPool = NewMapPool()

// NewMapPool creates a new MapPool.
func NewMapPool() *MapPool {
	return &MapPool{
		pool: sync.Pool{
			New: func() interface{} {
				// Allocate with initial capacity to avoid early resizing
				return make(map[string]interface{}, 8)
			},
		},
	}
}

// Get retrieves a map from the pool.
// The map contains arbitrary data and should be cleared before use if expected to be empty.
// However, standard usage pattern is to overwrite keys.
// To be safe, we rely on the user to populate it completely or clear it if needed?
// Actually, standard pool pattern for maps usually involves clearing ON PUT or ON GET.
// Go 1.21 `clear` makes it cheap.
// We will clear on Put to ensure Get always returns a clean map.
func (p *MapPool) Get() map[string]interface{} {
	return p.pool.Get().(map[string]interface{})
}

// Put returns a map to the pool.
// It clears the map before putting it back.
func (p *MapPool) Put(m map[string]interface{}) {
	if m == nil {
		return
	}
	clear(m) // Go 1.21+ builtin
	p.pool.Put(m)
}

// GetMap is a helper to get from the global pool.
func GetMap() map[string]interface{} {
	return globalPool.Get()
}

// PutMap is a helper to return to the global pool.
func PutMap(m map[string]interface{}) {
	globalPool.Put(m)
}
