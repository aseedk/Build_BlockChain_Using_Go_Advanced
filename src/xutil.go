package src

import (
	"fmt"
	"os"
)

func Handle(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func FileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println(err)
		return false
	}
	return true
}
