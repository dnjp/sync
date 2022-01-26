package main

import (
	"fmt"
	"os"
)

func main() {
	c := &client{}
	err := c.rootCmd().Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
