package main

import (
	"fmt"
	"github.com/fangyi-zhou/mpst-examples/twobuyer"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Two Buyer Protocol:")
	twobuyer.RunAll()
}
