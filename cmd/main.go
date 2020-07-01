package main

import (
	"dedicated-container-ingress-controller/cmd/dedicated-container-ingress-controller/cmd"
	"os"
)

func main() {
	rootCmd := cmd.GetRootCmd(os.Args[1:])

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
