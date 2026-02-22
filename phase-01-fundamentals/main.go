package main

import (
	"errors"
	"fmt"
)

func square(a int) int {
	return a * a
}

func divide(a, b int) (result float64, err error) {
	if b == 0 {
		err = errors.New("Cannot divide by zero!")
	}

	result = float64(a) / float64(b)
	return
}

func add(nums ...int) int {
	sum := 0
	for _, num := range nums {
		sum += num
	}
	return sum
}

// apply() takes an array of int, and a function to apply on the array
// returns the new array with applied function on each element
func apply(nums []int, fn func(int) int) []int {
	// make an array of int with the same length as `nums`
	result := make([]int, len(nums))
	// loop over nums
	for i, val := range nums {
		// apply `fn` and store in result[i]
		result[i] = fn(val)
	}
	// return result
	return result
}

// Closures (generator), returns a function
func idGenerator(start int) func() int {
	// take the start from the user
	id := start - 1
	// return a function that increments then returns the id
	return func() int {
		id++
		return id
	}
}

func main() {
	nums := []int{1, 2, 3, 4}
	result := apply(nums, func(i int) int { return i * i })
	fmt.Println(result)

	count := idGenerator(1)
	fmt.Println(count())
	fmt.Println(count())
}

// var x int = 5
// y := 6
// const PI float64 = 3.14
// val, err := divide(1, 2)
// if err != nil {
// 	fmt.Println(err)
// 	return
// }

// fmt.Println(
// 	"Hello world",
// 	x, y, square(y),
// 	val, add(5, 2, 3),
// )
