package main

import (
	"log"
	"os"

	"github.com/abibby/airtag-tracker/process"
	"github.com/gliderlabs/ssh"
)

var (
	Password = os.Getenv("PASSWORD")
)

func main() {
	err := generateKeys()
	if err != nil {
		panic(err)
	}

	s := ssh.Server{
		Addr:            ":2222",
		Handler:         handle,
		PasswordHandler: PasswordHandler,
	}
	err = s.SetOption(ssh.HostKeyFile(keyPath))
	if err != nil {
		panic(err)
	}
	err = s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func handle(s ssh.Session) {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("connection failed: %v", err)
		}
	}()

	err := process.Handle(s)
	if err != nil {
		log.Print(err)
	}
}

func PasswordHandler(ctx ssh.Context, password string) bool {
	return password == Password
}
