/*
 * Copyright (c) 2020. Vasyl Naumenko
 */

package main

import (
	"fmt"
	"time"
)

func main() {
	inputData := []int{0, 1, 1, 2, 3, 5, 8}
	// inputData := []int{0,1}

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				fmt.Println("cant convert result data to string")
			}
			fmt.Println(data)
		}),
	}

	start := time.Now()

	ExecutePipeline(hashSignJobs...)
	end := time.Since(start)

	fmt.Printf("execution took %s\n", end)
}
