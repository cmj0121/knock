package main

import (
	"os"

	"github.com/cmj0121/knock"
)

func main() {
	agent := knock.New()
	os.Exit(agent.Run())
}
