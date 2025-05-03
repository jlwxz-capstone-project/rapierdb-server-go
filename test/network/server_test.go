package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"

	network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
)

func TestHttpServerNetwork(t *testing.T) {
	opts := &network_server.HttpNetworkOptions{
		BaseUrl:         "localhost:8088",
		ReceiveEndpoint: "/api",
		SendEndpoint:    "/sse",
	}
	server := network_server.NewHttpNetwork(opts)

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
