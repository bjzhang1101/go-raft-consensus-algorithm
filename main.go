package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bjzhang1101/raft/server"
)

func main() {
	mainDone := make(chan os.Signal, 1)
	signal.Notify(mainDone, syscall.SIGINT, syscall.SIGTERM)

	s, err := server.NewServer()
	if err != nil {
		os.Exit(1)
	}

	go func() {
		addr := fmt.Sprintf(":%d", server.ClientPort)

		if err = s.ListenAndServe(addr); err != nil {
			os.Exit(1)
		}
	}()

	select {
	case <-mainDone:
	}
}
