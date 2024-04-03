package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/gliderlabs/ssh"
)

func main() {
	s := ssh.Server{
		Addr:    ":2222",
		Handler: handle,
		// PasswordHandler: PasswordHandler,
	}
	err := s.SetOption(ssh.HostKeyFile("/.ssh/id_rsa"))
	if err != nil {
		panic(err)
	}
	err = s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func handle(s ssh.Session) {
	// f, err := os.OpenFile("img.png", os.O_CREATE|os.O_WRONLY, 0o644)
	// if err != nil {
	// 	panic(err)
	// }
	// _, err = io.Copy(f, s)
	// if errors.Is(err, io.EOF) {
	// 	// fallthrough
	// } else if err != nil {
	// 	panic(err)
	// }

	b, err := io.ReadAll(s)
	if err != nil {
		log.Printf("failed to read image: %v", err)
		return
	}

	txt, err := extract(b)
	if err != nil {
		log.Printf("failed to process image: %v", err)
		return
	}

	fmt.Sprintln(txt)

	log.Println("end")
}

func PasswordHandler(ctx ssh.Context, password string) bool {
	return password == os.Getenv("PASSWORD")
}

func extract(imageBytes []byte) (string, error) {
	cmd := exec.Command("tesseract", "-", "-", "-l", "eng")
	cmd.Stdin = bytes.NewBuffer(imageBytes)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
