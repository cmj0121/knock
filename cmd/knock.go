package main

import (
	"fmt"

	"github.com/cmj0121/knock"
)

func main() {
	agent := knock.New()
	if err := agent.Run(); err != nil {
		fmt.Println(err)
	}
}
