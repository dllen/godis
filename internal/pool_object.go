package internal

import "github.com/go-redis/redis"

// PoolObject is a pack struct for addr and the appropriate redis pool.
type PooledObject struct {
	Addr   string
	Client *redis.Client
}

// NewPoolObject return the pack struct for addr and the appropriate redis pool.
func NewPooledObject(addr string, client *redis.Client) *PooledObject {
	pooledObject := &PooledObject{
		Addr:   addr,
		Client: client,
	}
	return pooledObject
}

