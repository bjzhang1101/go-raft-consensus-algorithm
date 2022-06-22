package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"

	"github.com/bjzhang1101/raft/node"
	"github.com/bjzhang1101/raft/server"
)

var (
	configFile = pflag.String("config-file", "", "path of the configuration file")
)

func main() {
	mainDone := make(chan os.Signal, 1)
	signal.Notify(mainDone, syscall.SIGINT, syscall.SIGTERM)

	c, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("failed to load config file: %v", err)
	}

	n := node.NewNode(c.id, c.quorum)
	exit := func(err error) {
		n.SetState(node.Down)
		// TODO: change to log.
		fmt.Printf("failed to load config file: %v", err)
		mainDone <- syscall.SIGTERM
	}

	ctx, cancel := context.WithCancel(context.Background())
	nodeDone := make(<-chan struct{})
	// Goroutine for node operations.
	go func() {
		nodeDone = n.Start(ctx)
	}()

	// Server to respond client's Http requests.
	s, err := server.NewServer(n)
	if err != nil {
		exit(err)
	}

	// Goroutine for the Http server.
	go func() {
		addr := fmt.Sprintf(":%d", server.ClientPort)

		if err = s.ListenAndServe(addr); err != nil {
			exit(err)
		}
	}()

	// TODO(bjzhang1101): Handle graceful shutdown, add wait group.
	select {
	case <-mainDone:
	case <-nodeDone:
	}
	cancel()
}
