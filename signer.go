package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// сюда писать код

// CombineResults combine all MultiHash hashes into one string
func CombineResults(in, out chan interface{}) {
	var hashes []string
	for data := range in {
		hashes = append(hashes, data.(string))
	}
	sort.Strings(hashes)
	out <- strings.Join(hashes, "_")
}

// MultiHash adds iteration hashes to the SingleHash output
func MultiHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	hash := HashMaker{}
	mHash := func(str string, wg *sync.WaitGroup) {
		out <- hash.GetMultiHash(str)
		wg.Done()
	}
	for data := range in {
		wg.Add(1)
		go mHash(data.(string), &wg)
	}
	wg.Wait()
}

// SingleHash outputs hash based on int number
func SingleHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	hash := HashMaker{}
	hashed := func(i int, wg *sync.WaitGroup) {
		out <- hash.GetHash(i)
		wg.Done()
	}

	for i := range in {
		wg.Add(1)
		go hashed(i.(int), &wg)
	}
	wg.Wait()
}

// ExecutePipeline executes jobs and provides data flow from one job to the next one
func ExecutePipeline(jobs ...job) {
	in := make(chan interface{})
	var wg sync.WaitGroup
	for _, j := range jobs {
		in = pipe(j, in, &wg)
	}
	wg.Wait()
}

// pipe gets data from input (in) channel, runs the job, send results to the next job input (in2) channel
func pipe(j job, in chan interface{}, wg *sync.WaitGroup) chan interface{} {
	out := make(chan interface{})
	in2 := make(chan interface{})

	go func() {
		j(in, out)
		close(out)
	}()

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		for i := range out {
			in2 <- i
		}
		close(in2)
		wg.Done()
	}(wg)

	return in2
}

// HashMaker used for making hashes
type HashMaker struct {
	sync.Mutex
}

// GetHash returns hash for int number. It takes 1s 10ms for run
func (h *HashMaker) GetHash(i int) string {
	var wg sync.WaitGroup
	chHash := make(chan string)

	hash1 := func(str string, wg *sync.WaitGroup) {
		chHash <- DataSignerCrc32(str) // takes 1 sek
		wg.Done()
	}

	hash2 := func(str string, wg *sync.WaitGroup) {
		h.Lock()
		str = DataSignerMd5(str) // takes 10ms, overheat if 2 runs simultaneously
		<-time.After(10 * time.Millisecond)
		h.Unlock()
		chHash <- DataSignerCrc32(str) // takes 1 sek
		wg.Done()
	}

	str := strconv.Itoa(i)

	wg.Add(2)
	go hash1(str, &wg)
	go hash2(str, &wg)

	go func() {
		wg.Wait()
		close(chHash)
	}()

	var hashes []string
	for h := range chHash {
		hashes = append(hashes, h)
	}

	return strings.Join(hashes, "~")
}

// GetMultiHash returns advanced hash. It takes 1s for run
func (h *HashMaker) GetMultiHash(str string) string {
	const iterations = 6
	var combined [iterations]string
	var wg sync.WaitGroup
	type ihash struct {
		i int    // index
		s string // hash
	}
	chHash := make(chan ihash)
	hashed := func(i int, str string, wg *sync.WaitGroup) {
		chHash <- ihash{
			i: i,
			s: DataSignerCrc32(strconv.Itoa(i) + str), // takes 1 sek
		}
		wg.Done()
	}

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go hashed(i, str, &wg)
	}

	go func() {
		wg.Wait()
		close(chHash)
	}()

	for h := range chHash {
		combined[h.i] = h.s
	}
	hashes := combined[:]

	return strings.Join(hashes, "")
}
