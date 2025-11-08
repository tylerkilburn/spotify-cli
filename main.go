/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"spotify-cli/cmd"
	"spotify-cli/config"
)

func main() {
	fmt.Printf("Port %v", config.Get().Http.Port)
	cmd.Execute()
}
