package utils_test

import (
	"fmt"
	"./utils"
)

func ExampleGetStringUntil() {
	s := utils.GetStringUntil("abc")
	fmt.Println("s = " + s)
	// Output:
	// apa
}
