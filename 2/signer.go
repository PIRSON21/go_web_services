package main

import (
	"slices"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код


func ExecutePipeline(jobs ...job) {
	var in chan interface{}
	wg := &sync.WaitGroup{}
	for _, j := range jobs {
		out := make(chan interface{}, 100)
		wg.Add(1)
		go func(j job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			j(in, out)
		}(j, in, out)

		in = out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for n := range in {
		num := n.(int)
		data := strconv.Itoa(num)
		md5Data := DataSignerMd5(data)
		wg.Add(1)
		go singleHash(wg, data, out, md5Data)
	}
	wg.Wait()
}

func singleHash(outerWG *sync.WaitGroup, data string, out chan interface{}, md5Data string) {
	defer outerWG.Done()
	wg := &sync.WaitGroup{}
	var fst string
	wg.Add(1)
	go func() {
		defer wg.Done()
		fst = DataSignerCrc32(data)
	}()
	
	var scd string
	wg.Add(1)
	go func() {
		defer wg.Done()
		scd = DataSignerCrc32(md5Data)
	}()

	wg.Wait()
	out <- fst + "~" + scd
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for s := range in {
		data := s.(string)
		wg.Add(1)
		go multiHash(wg, data, out)
	}
	wg.Wait()
}

func multiHash(outerWg *sync.WaitGroup, data string, out chan interface{}) {
	defer outerWg.Done()
	var slice = make([]string, 6)
	wg := &sync.WaitGroup{}
	for i := range 6 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			slice[i] = DataSignerCrc32(strconv.Itoa(i) + data)
		}(i)
	}
	wg.Wait()
	out <- strings.Join(slice, "")
}

func CombineResults(in, out chan interface{}) {
	var slice []string
	for s := range in {
		data := s.(string)
		slice = append(slice, data)
	}

	slices.Sort(slice)

	res := strings.Join(slice, "_")
	out <- res
}