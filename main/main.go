package main

import (
	"fmt"
	"os"

	"github.com/object88/gha-publish-release-assets/cmd"
)

func main() {
	rootCmd := cmd.InitializeCommands()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
