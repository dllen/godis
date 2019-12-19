package godis

import (
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func Test(t *testing.T) {
	options := redis.Options{
		DB: 0,
	}

	server1, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start redis server, %v", err)
	}
	defer server1.Close()

	server2, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start redis server, %v", err)
	}
	defer server2.Close()

	var hosts []string

	hosts = append(hosts, server1.Addr())

	pool, err := NewRoundRobinPool(hosts, options)
	if err != nil {
		t.Fatalf("failed to create pool, %v", err)
	}
	defer pool.Close()

	pool.GetClient().Set("k1", "v1", 0)
	v1, _ := server1.Get("k1")
	assert.Equal(t, "v1", v1)

	time.Sleep(time.Second)
	pool.GetClient().Set("k2", "v2", 0)
	v2, _ := server1.Get("k2")
	assert.Equal(t, "v2", v2)

	time.Sleep(time.Second)

	hosts = hosts[:0]
	hosts = append(hosts, server2.Addr())
	pool.ResetPools(hosts)

	pool.GetClient().Set("k3", "v3", 0)
	v3, _ := server2.Get("k3")
	assert.Equal(t, "v3", v3)

	time.Sleep(time.Second)

	pool.GetClient().Set("k4", "v4", 0)
	v4, _ := server2.Get("k4")
	assert.Equal(t, "v4", v4)
}

func BenchmarkNewRoundRobinPool(b *testing.B) {
	b.N = 1000
	var hosts []string
	options := redis.Options{
		DB: 0,
	}
	hosts = append(hosts, "192.168.33.131:6379")
	pool, err := NewRoundRobinPool(hosts, options)
	if err != nil {
		b.Fatalf("failed to create pool, %v", err)
	}
	defer pool.Close()
	b.ResetTimer()

	var wg sync.WaitGroup
	max := b.N
	wg.Add(max)
	expire := time.Second
	for i := 0; i < max; i++ {
		go func() {
			key := strconv.Itoa(rand.Int())
			val := strconv.Itoa(rand.Int())
			pool.GetClient().Set(key, val, expire)
			v1, _ := pool.GetClient().Get(key).Result()
			assert.Equal(b, val, v1)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestNewRoundRobinPool(t *testing.T) {
	options := redis.Options{
		DB: 0,
	}

	var hosts []string

	hosts = append(hosts, "192.168.33.131:6379")

	pool, err := NewRoundRobinPool(hosts, options)
	if err != nil {
		t.Fatalf("failed to create pool, %v", err)
	}

	defer pool.Close()


	pool.GetClient().Set("k1", "v1", 0)
	v1, _ := pool.GetClient().Get("k1").Result()
	assert.Equal(t, "v1", v1)

	time.Sleep(time.Second)
	pool.GetClient().Set("k2", "v2", 0)
	v2, _ := pool.GetClient().Get("k2").Result()
	assert.Equal(t, "v2", v2)

	time.Sleep(time.Second)

	pool.GetClient().Set("k3", "v3", 0)
	v3, _ := pool.GetClient().Get("k3").Result()
	assert.Equal(t, "v3", v3)

	time.Sleep(time.Second)

	pool.GetClient().Set("k4", "v4", 0)
	v4, _ := pool.GetClient().Get("k4").Result()
	assert.Equal(t, "v4", v4)
}
