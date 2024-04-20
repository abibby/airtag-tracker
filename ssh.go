package main

import (
	"fmt"
	"os"
	"os/exec"
)

const keyPath = "/.ssh/id_rsa"

func generateKeys() error {
	_, err := os.Stat(keyPath)
	if err == nil {
		return nil
	}
	b, err := exec.Command("ssh-keygen", "-t", "rsa", "-N", "", "-f", keyPath).CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", b)
		return err
	}
	return nil
}
