package cache

import "io"

// Cache - interface for a key value store
type Cache interface {
	Input(key string) (io.Writer, io.Closer, error)
	Output(key string) (io.Reader, io.Closer, error)
	OutputLastGoodCopy(key string) (io.Reader, io.Closer, error)
	Expire(key string, ttl int)
}
