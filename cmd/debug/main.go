package main

import (
	"fmt"
	"os"
)

func main() {
	for _, s := range os.Environ() {
		fmt.Println(s)
	}
}
