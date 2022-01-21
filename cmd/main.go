package main

import (
	"fmt"
	"github.com/fangyi-zhou/mpst-examples/twobuttons"
	"log"
	"math/rand"
	"os"
	"strconv"
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
	case "twobuttons":
		fmt.Println("Two Buttons Protocol:")
		twobuttons.RunAll()
	case "twobuttonsmailbox":
		fmt.Println("Two Buttons Protocol:")
		twobuttons.RunAllMailbox()
	case "twobuttonsmailboxmulti":
		fmt.Println("Two Buttons Protocol:")
		iterations, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Panic(err)
		}
		twobuttons.RunAllMailboxMulti(iterations)
	}
}
