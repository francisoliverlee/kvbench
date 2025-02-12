package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/smallnest/kvbench"
	"github.com/smallnest/log"
)

var (
	duration = flag.Duration("d", time.Minute, "test duration for each case")
	c        = flag.Int("c", runtime.NumCPU(), "concurrent goroutines")
	size     = flag.Int("size", 256, "data size")
	fsync    = flag.Bool("fsync", false, "fsync")
	s        = flag.String("s", "map", "store type")
	data     = make([]byte, *size)
)

func main() {

	flag.Parse()
	fmt.Printf("duration=%v, c=%d size=%d\n", *duration, *c, *size)

	var memory bool
	var path string
	if strings.HasSuffix(*s, "/memory") {
		memory = true
		path = ":memory:"
		*s = strings.TrimSuffix(*s, "/memory")
	}

	store, path, err := getStore(*s, *fsync, path)
	if err != nil {
		panic(err)
	}
	if !memory {
		defer os.RemoveAll(path)
	}

	defer store.Close()
	name := *s
	if memory {
		name = name + "/memory"
	}
	if *fsync {
		name = name + "/fsync"
	} else {
		name = name + "/nofsync"
	}

	testBatchWrite(name, store)
	testSet(name, store)
	testGet(name, store)
	testGetSet(name, store)
	testDelete(name, store)
}

// test batch writes
func testBatchWrite(name string, store kvbench.Store) {
	var wg sync.WaitGroup
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	var total uint64
	for i := 0; i < *c; i++ {
		wg.Add(1)
		go func(proc int) {
			batchSize := uint64(1000)
			var keyList, valList [][]byte
			for i := uint64(0); i < batchSize; i++ {
				keyList = append(keyList, genKey(i))
				valList = append(valList, make([]byte, *size))
			}
		LOOP:
			for {
				select {
				case <-ctx.Done():
					break LOOP
				default:
					// Fill random keys and values.
					for i := range keyList {
						rand.Read(keyList[i])
						rand.Read(valList[i])
					}
					err := store.PSet(keyList, valList)
					if err != nil {
						panic(err)
					}
					atomic.AddUint64(&total, uint64(len(keyList)))
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Printf("%s batch write test inserted: %d entries; took: %s s\n", name, total, time.Since(start))
}

// test get
func testGet(name string, store kvbench.Store) {
	var wg sync.WaitGroup
	wg.Add(*c)

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	counts := make([]int, *c)
	start := time.Now()
	for j := 0; j < *c; j++ {
		index := uint64(j)
		go func() {
			var count int
			i := index
		LOOP:
			for {
				select {
				case <-ctx.Done():
					break LOOP
				default:
					_, ok, _ := store.Get(genKey(i))
					if !ok {
						i = index
					}
					i += uint64(*c)
					count++
				}
			}
			counts[index] = count
			wg.Done()
		}()
	}
	wg.Wait()
	dur := time.Since(start)
	d := int64(dur)
	var n int
	for _, count := range counts {
		n += count
	}
	fmt.Printf("%s get rate: %d op/s, mean: %d ns, took: %d s\n", name, int64(n)*1e6/(d/1e3), d/int64((n)*(*c)), int(dur.Seconds()))
}

// test multiple get/one set
func testGetSet(name string, store kvbench.Store) {
	var wg sync.WaitGroup
	wg.Add(*c)

	ch := make(chan struct{})

	var setCount uint64

	go func() {
		i := uint64(0)
		for {
			select {
			case <-ch:
				return
			default:
				store.Set(genKey(i), data)
				setCount++
				i++
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	counts := make([]int, *c)
	start := time.Now()
	for j := 0; j < *c; j++ {
		index := uint64(j)
		go func() {
			var count int
			i := index
		LOOP:
			for {
				select {
				case <-ctx.Done():
					break LOOP
				default:
					_, ok, _ := store.Get(genKey(i))
					if !ok {
						i = index
					}
					i += uint64(*c)
					count++
				}
			}
			counts[index] = count
			wg.Done()
		}()
	}
	wg.Wait()
	close(ch)
	dur := time.Since(start)
	d := int64(dur)
	var n int
	for _, count := range counts {
		n += count
	}

	if setCount == 0 {
		fmt.Printf("%s setmixed rate: -1 op/s, mean: -1 ns, took: %d s\n", name, int(dur.Seconds()))
	} else {
		fmt.Printf("%s setmixed rate: %d op/s, mean: %d ns, took: %d s\n", name, int64(setCount)*1e6/(d/1e3), d/int64(setCount), int(dur.Seconds()))
	}
	fmt.Printf("%s getmixed rate: %d op/s, mean: %d ns, took: %d s\n", name, int64(n)*1e6/(d/1e3), d/int64((n)*(*c)), int(dur.Seconds()))
}

func testSet(name string, store kvbench.Store) {
	var wg sync.WaitGroup
	wg.Add(*c)

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	counts := make([]int, *c)
	start := time.Now()
	for j := 0; j < *c; j++ {
		index := uint64(j)
		go func() {
			count := 0
			i := index
		LOOP:
			for {
				select {
				case <-ctx.Done():
					break LOOP
				default:
					store.Set(genKey(i), data)
					i += uint64(*c)
					count++
				}
			}
			counts[index] = count
			wg.Done()
		}()
	}
	wg.Wait()
	dur := time.Since(start)
	d := int64(dur)
	var n int
	for _, count := range counts {
		n += count
	}
	fmt.Printf("%s set rate: %d op/s, mean: %d ns, took: %d s\n", name, int64(n)*1e6/(d/1e3), d/int64((n)*(*c)), int(dur.Seconds()))
}

func testDelete(name string, store kvbench.Store) {
	var wg sync.WaitGroup
	wg.Add(*c)

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	counts := make([]int, *c)
	start := time.Now()
	for j := 0; j < *c; j++ {
		index := uint64(j)
		go func() {
			var count int
			i := index
		LOOP:
			for {
				select {
				case <-ctx.Done():
					break LOOP
				default:
					store.Del(genKey(i))
					i += uint64(*c)
					count++
				}
			}
			counts[index] = count
			wg.Done()
		}()
	}
	wg.Wait()
	dur := time.Since(start)
	d := int64(dur)
	var n int
	for _, count := range counts {
		n += count
	}

	fmt.Printf("%s del rate: %d op/s, mean: %d ns, took: %d s\n", name, int64(n)*1e6/(d/1e3), d/int64((n)*(*c)), int(dur.Seconds()))
}

func genKey(i uint64) []byte {
	r := make([]byte, 9)
	r[0] = 'k'
	binary.BigEndian.PutUint64(r[1:], i)
	return r
}

func getStore(s string, fsync bool, path string) (kvbench.Store, string, error) {
	var store kvbench.Store
	var err error
	switch s {
	default:
		err = fmt.Errorf("unknown store type: %v", s)
	case "map":
		if path == "" {
			path = "map.db"
		}
		store, err = kvbench.NewMapStore(path, fsync)
	case "btree":
		if path == "" {
			path = "btree.db"
		}
		store, err = kvbench.NewBTreeStore(path, fsync)
	case "bolt":
		if path == "" {
			path = "bolt.db"
		}
		store, err = kvbench.NewBoltStore(path, fsync)
	case "bbolt":
		if path == "" {
			path = "bbolt.db"
		}
		store, err = kvbench.NewBboltStore(path, fsync)
	case "leveldb":
		if path == "" {
			path = "leveldb.db"
		}
		store, err = kvbench.NewLevelDBStore(path, fsync)
	case "kv":
		log.Warningf("kv store is unstable")
		if path == "" {
			path = "kv.db"
		}
		store, err = kvbench.NewKVStore(path, fsync)
	case "badger":
		if path == "" {
			path = "badger.db"
		}
		store, err = kvbench.NewBadgerStore(path, fsync)
	case "buntdb":
		if path == "" {
			path = "buntdb.db"
		}
		store, err = kvbench.NewBuntdbStore(path, fsync)
	//case "rocksdb":
	//	if path == "" {
	//		path = "rocksdb.db"
	//	}
	//	store, err = kvbench.NewRocksdbStore(path, fsync)
	case "pebble":
		if path == "" {
			path = "pebble.db"
		}
		store, err = kvbench.NewPebbleStore(path, fsync)
	case "pogreb":
		if path == "" {
			path = "pogreb.db"
		}
		store, err = kvbench.NewPogrebStore(path, fsync)
	case "nutsdb":
		if path == "" {
			path = "nutsdb.db"
		}
		store, err = kvbench.NewNutsdbStore(path, fsync)
	}

	return store, path, err
}
