package cache

// Cache - interface for a key value store
type Cache interface {
	Set(key string, value []byte)
	Get(key string) ([]byte, error)
	GetLastGoodCopy(key string) ([]byte, error)
	Refresh(key string)
	Expire(key string, ttl int)
}
