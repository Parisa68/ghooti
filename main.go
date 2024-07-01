package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	run()
}

func run() {
	command := exec.Command("/bin/bash")
	err := command.Run()
	if err != nil {
		log.Fatalf("command failed with %s\n", err)
	}
	fmt.Printf("Command executed successfully.\n")
}
