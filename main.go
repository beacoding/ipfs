package main

import (
	"log"

	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	s, err := server.New()
	if err != nil {
		return err
	}
	return s.Listen(":0")
}
