package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestSyncMap(t *testing.T) {
	m := new(sync.Map)

	c := make(chan int, 3)

	latch := new(sync.WaitGroup)

	latch.Add(1)

	go func() {
		latch.Wait()
		for i := 0; i < 10000; i++ {
			m.Store(fmt.Sprintf("i%d", i), fmt.Sprintf("i%d", i))
		}

		c <- 0
	}()

	go func() {
		latch.Wait()
		for j := 0; j < 10000; j++ {
			m.Store(fmt.Sprintf("j%d", j), fmt.Sprintf("j%d", j))
		}

		c <- 1
	}()

	go func() {
		latch.Wait()
		for k := 0; k < 10000; k++ {
			m.Store(fmt.Sprintf("k%d", k), fmt.Sprintf("k%d", k))
		}

		c <- 2
	}()

	t.Log("-------------------")

	latch.Done()

	<-c
	<-c
	<-c

	countMap := make(map[string]int)

	m.Range(func(key, value interface{}) bool {
		strKey := key.(string)

		k := strKey[0:1]

		countMap[k] = countMap[k] + 1

		return true
	})

	for k, v := range countMap {
		t.Logf("%s : %d", k, v)
	}
}
