package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"

	"github.com/bjzhang1101/raft/grpc"
	"github.com/bjzhang1101/raft/http"
	"github.com/bjzhang1101/raft/node"
)

var (
	configFile = pflag.String("config-file", "", "path of the configuration file")
	ip         = pflag.String("ip", "", "ip of the dest")
)

func main() {
	pflag.Parse()

	mainDone := make(chan os.Signal, 1)
	signal.Notify(mainDone, syscall.SIGINT, syscall.SIGTERM)

	c, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("failed to load config file: %v", err)
	}

	n := node.NewNode(c.ID, c.Quorum)
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

	// Server to respond http requests.
	httpServer := http.NewServer(n, *ip)
	go func() {
		addr := fmt.Sprintf(":%d", http.DefaultPort)

		if err = httpServer.ListenAndServe(addr); err != nil {
			exit(err)
		}
	}()

	// Server to respond gRPC requests.
	grpcServer := grpc.NewServer(n)
	go func() {
		lis, err := net.Listen("tcp4", fmt.Sprintf(":%d", grpc.DefaultPort))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		if err = grpcServer.Serve(lis); err != nil {
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
