package main

import (
	"fmt"
	"sync"
)

const N = 10

func add(a, b, c []int, wg *sync.WaitGroup, id int) {
	defer wg.Done()
	if id < N {
		c[id] = a[id] + b[id]
	}
}

func main() {
	a := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b := []int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	c := make([]int, N)

	var wg sync.WaitGroup

	for i := 0; i < N; i++ {
		wg.Add(1)
		go add(a, b, c, &wg, i)
	}

	wg.Wait()

	fmt.Println("Result:", c)
}
