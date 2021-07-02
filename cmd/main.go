package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/fangyi-zhou/mpst-examples/twobuyer"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	var protocol string
	if len(os.Args) == 1 {
		// Default TwoBuyer
		protocol = "twobuyer"
	} else {
		protocol = os.Args[1]
		protocol = strings.ToLower(protocol)
	}
	switch protocol {
	case "twobuyer":
		fmt.Println("Two Buyer Protocol:")
		twobuyer.RunAll()
	case "twobuyerbad":
		fmt.Println("Bad Two Buyer Protocol:")
		twobuyer.RunAllBad()
	}
}
