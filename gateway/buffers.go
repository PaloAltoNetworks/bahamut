package gateway

import "sync"

// bufferPool implements the interface of httputil.BufferPool in order
// to improve memory utilization in the reverse proxy.
type bufferPool struct {
	s sync.Pool
}

func newPool(size int) *bufferPool {
	return &bufferPool{
		s: sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		},
	}
}

// Get gets a buffer from the pool.
func (b *bufferPool) Get() []byte {
	return b.s.Get().([]byte)
}

// Put returns the buffer to the pool.
func (b *bufferPool) Put(buf []byte) {
	b.s.Put(buf) // nolint
}
