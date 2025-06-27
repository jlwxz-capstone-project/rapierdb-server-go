package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"

	network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
)

func TestHttpServerNetwork(t *testing.T) {
	opts := &network_server.WebSocketNetworkOptions{
		BaseUrl:       "localhost:8088",
		WebSocketPath: "/ws",
	}
	ctx, _ := context.WithCancel(context.Background())
	server := network_server.NewWebSocketNetworkWithContext(opts, ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	server.SetMsgHandler(func(clientId string, msg []byte) {
		fmt.Println("receive from " + clientId + ": " + string(msg))
		server.Send(clientId, []byte("hello client "+clientId))
	})

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	<-sigChan

	err = server.Stop()
	if err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}
}
