package main

import (
	"os"

	"chatgpt-cli/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Args[1:], os.Stdout, os.Stderr))
}
