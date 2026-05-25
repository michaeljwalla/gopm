package main

import (
	"fmt"
	"math/rand"

	"golang.org/x/exp/constraints"
)

func simple_sort_asc[T constraints.Ordered](nums []T) {
	for i := len(nums) - 1; i >= 0; i-- {
		var nodiff bool = true
		for j := 0; j < i; j++ {
			if nums[j] <= nums[j+1] {
				continue
			}
			nums[j], nums[j+1] = nums[j+1], nums[j]
			nodiff = false
		}
		if nodiff {
			break
		}
	}
}

func main() {
	var n, min, max int
	fmt.Print("Num. elems: ")
	fmt.Scan(&n)
	fmt.Print("Minimum (int): ")
	fmt.Scan(&min)
	fmt.Print("Maximum (int): ")
	fmt.Scan(&max)

	my_nums := make([]int, n)
	for i := 0; i < n; i++ {
		my_nums[i] = rand.Intn(max-min) + min
	}
	fmt.Println("My slice: ", my_nums)
	simple_sort_asc(my_nums)
	fmt.Println("Sorted: ", my_nums)
}
