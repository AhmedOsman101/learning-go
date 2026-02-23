package exercises

import "fmt"

func FizzBuzz(n int) (result string, wasModified bool) {
	wasModified = true
	switch {
	case n%15 == 0:
		result = "FizzBuzz"
	case n%3 == 0:
		result = "Fizz"
	case n%5 == 0:
		result = "Buzz"
	default:
		result = fmt.Sprint(n)
		wasModified = false
	}

	return
}
