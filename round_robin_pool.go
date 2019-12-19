package godis

import (
	"github.com/dllen/godis/internal"
	"github.com/go-redis/redis"
	"log"
	"sort"
	"sync/atomic"
)

// RoundRobinPool is a round-robin redis client pool for connecting multiple codis proxies based on
// redis-go.
type RoundRobinPool struct {
	pools   atomic.Value
	hosts   []string
	options redis.Options
	nextIdx int64
}

// NewRoundRobinPool return a round-robin redis client pool specified by
// proxy hosts and redis options.
func NewRoundRobinPool(hosts []string, options redis.Options) (*RoundRobinPool, error) {
	pool := &RoundRobinPool{
		nextIdx: -1,
		pools:   atomic.Value{},
		hosts:   hosts,
	}
	pool.pools.Store([]*internal.PooledObject{})
	pool.ResetPools(hosts)
	return pool, nil
}

func (p *RoundRobinPool) ResetPools(hosts []string) {
	p.hosts = hosts
	sort.Strings(hosts)
	pools := p.pools.Load().([]*internal.PooledObject)
	addr2Pool := make(map[string]*internal.PooledObject, len(pools))
	for _, pool := range pools {
		addr2Pool[pool.Addr] = pool
	}
	newPools := make([]*internal.PooledObject, 0)
	for _, host := range hosts {
		proxyInfo := internal.ProxyInfo{
			Addr: host,
		}
		addr := proxyInfo.Addr
		if pooledObject, ok := addr2Pool[addr]; ok {
			newPools = append(newPools, pooledObject)
			delete(addr2Pool, addr)
		} else {
			options := p.cloneOptions()
			options.Addr = addr
			options.Network = "tcp"
			pooledObject := internal.NewPooledObject(
				addr,
				redis.NewClient(&options),
			)
			newPools = append(newPools, pooledObject)
			log.Printf("Add new proxy: %s", addr)
		}
	}

	p.pools.Store(newPools)
	for _, pooledObject := range addr2Pool {
		log.Printf("Remove proxy: %s", pooledObject.Addr)
		err := pooledObject.Client.Close()
		if err != nil {
			log.Printf("Close client err: %v", err)
		}
	}

}

// GetClient can get a redis client from pool with round-robin policy.
// It's safe for concurrent use by multiple goroutines.
func (p *RoundRobinPool) GetClient() *redis.Client {
	pools := p.pools.Load().([]*internal.PooledObject)
	for {
		current := atomic.LoadInt64(&p.nextIdx)
		var next int64
		if (current) >= (int64)(len(pools))-1 {
			next = 0
		} else {
			next = current + 1
		}
		if atomic.CompareAndSwapInt64(&p.nextIdx, current, next) {
			return pools[next].Client
		}
	}
}

func (p *RoundRobinPool) cloneOptions() redis.Options {
	options := redis.Options{
		Network:            p.options.Network,
		Addr:               p.options.Addr,
		Dialer:             p.options.Dialer,
		OnConnect:          p.options.OnConnect,
		Password:           p.options.Password,
		DB:                 p.options.DB,
		MaxRetries:         p.options.MaxRetries,
		MinRetryBackoff:    p.options.MinRetryBackoff,
		MaxRetryBackoff:    p.options.MaxRetryBackoff,
		DialTimeout:        p.options.DialTimeout,
		ReadTimeout:        p.options.ReadTimeout,
		WriteTimeout:       p.options.WriteTimeout,
		PoolSize:           p.options.PoolSize,
		PoolTimeout:        p.options.PoolTimeout,
		IdleTimeout:        p.options.IdleTimeout,
		IdleCheckFrequency: p.options.IdleCheckFrequency,
		TLSConfig:          p.options.TLSConfig,
	}
	return options
}

// Close closes the pool, releasing all resources except zookeeper client.
func (p *RoundRobinPool) Close() {
	pools := p.pools.Load().([]*internal.PooledObject)
	for _, pool := range pools {
		err := pool.Client.Close()
		if err != nil {
			log.Printf("Client close err: %v", err)
		}
	}
}
