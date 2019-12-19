# godis

##Godis - Golang client for codis

支持功能列表:

- 动态添加hosts
- 保证codis proxy 负载均衡

## Download

```
go get -u github.com/dllen/godis
```

## Example
```go
package main

import (
	"fmt"
	"github.com/dllen/godis"
	"github.com/go-redis/redis"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

func main() {
	var hosts []string
	options := redis.Options{
		DB: 0,
	}
	hosts = append(hosts, "192.168.33.131:6379")
	pool, err := godis.NewRoundRobinPool(hosts, options)
	if err != nil {
		fmt.Printf("failed to create pool, %v", err)
	}
	defer pool.Close()

	var wg sync.WaitGroup
	max := 100
	wg.Add(max)
	expire := time.Second
	for i := 0; i < max; i++ {
		go func() {
			key := strconv.Itoa(rand.Int())
			val := strconv.Itoa(rand.Int())
			pool.GetClient().Set(key, val, expire)
			v1, _ := pool.GetClient().Get(key).Result()
			if val != v1 {
				fmt.Printf("failed get set, %s:%s", val, v1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
```