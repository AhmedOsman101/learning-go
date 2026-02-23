package exercises

import (
	"fmt"
	"os"
)

func Greet() {
	username := os.Args[1]
	fmt.Println("Hello", username)
}
