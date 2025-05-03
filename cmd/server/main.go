package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/auth"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/synchronizer"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

type serverOptions struct {
	dbPath      string
	addr        string
	sseEndpoint string
	apiEndpoint string
	showHelp    bool
}

func parseServerOptions() *serverOptions {
	// Define command line parameters
	opts := &serverOptions{
		addr:        "localhost:8080",
		sseEndpoint: "/sse",
		apiEndpoint: "/api",
	}

	// Set command line flags
	flag.StringVar(&opts.dbPath, "path", "", "Database storage path")
	flag.StringVar(&opts.addr, "addr", opts.addr, "Server address, e.g.: localhost:8080")
	flag.StringVar(&opts.sseEndpoint, "sse", opts.sseEndpoint, "SSE endpoint path (server sends events to clients)")
	flag.StringVar(&opts.apiEndpoint, "api", opts.apiEndpoint, "API endpoint path (clients send requests to server)")
	flag.BoolVar(&opts.showHelp, "help", false, "Show help information")

	// Parse command line arguments
	flag.Parse()

	return opts
}

func ensureServerOptionsValid(opts *serverOptions) {
	if opts.dbPath == "" {
		log.Error("Database path must be provided")
		log.Info("Usage: server -path <db-path> [-addr <address>] [-sse <sse-endpoint>] [-api <api-endpoint>]")
		os.Exit(1)
	}
}

func main() {
	// Parse command line arguments
	opts := parseServerOptions()

	// Show help information
	if opts.showHelp {
		log.Info("Usage: server -path <db-path> [-addr <address>] [-sse <sse-endpoint>] [-api <api-endpoint>]")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Ensure command line arguments are valid
	ensureServerOptionsValid(opts)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for termination
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Infof("Received signal: %v. Initiating shutdown...", sig)
		cancel()
	}()

	// Open database engine
	// Note: Consider using DefaultStorageEngineOptions if needed, like in tests
	dbEngine, err := storage_engine.OpenStorageEngine(&storage_engine.StorageEngineOptions{
		Path: opts.dbPath,
	})
	if err != nil {
		log.Errorf("Unable to open database: %v", err) // Use Errorf for formatted error
		os.Exit(1)
	}

	<-dbEngine.WaitForStatus(storage_engine.StorageEngineStatusOpen)
	log.Info("Database opened successfully")

	// Configure and create the HTTP network server
	serverOptions := &network_server.HttpNetworkOptions{
		BaseUrl:         opts.addr,
		ReceiveEndpoint: opts.apiEndpoint,             // Clients send API requests here
		SendEndpoint:    opts.sseEndpoint,             // Server sends SSE events from here
		Authenticator:   &auth.HttpMockAuthProvider{}, // Using mock auth for now
		ShutdownTimeout: 5 * time.Second,              // Default shutdown timeout
	}
	server := network_server.NewHttpNetworkWithContext(serverOptions, ctx)

	log.Info("Starting server...")
	err = server.Start()
	if err != nil {
		log.Errorf("Failed to start server: %v", err) // Use Errorf
		os.Exit(1)
	}

	// Wait for server to be running using status subscription
	serverStatusCh := server.SubscribeStatusChange()
	cleanupServerWait := func() {
		server.UnsubscribeStatusChange(serverStatusCh)
	}
	<-util.WaitForStatus(server.GetStatus, network_server.NetworkRunning, serverStatusCh, cleanupServerWait, 0)
	log.Info("Server started successfully")

	// Start synchronizer
	synchronizerConfig := &synchronizer.SynchronizerConfig{}
	// Pass context for graceful shutdown
	s := synchronizer.NewSynchronizerWithContext(ctx, dbEngine, server, synchronizerConfig)
	err = s.Start()
	if err != nil {
		log.Errorf("Failed to start synchronizer: %v", err) // Use Errorf
		// Consider cancelling context here if synchronizer fails to start?
		os.Exit(1)
	}
	<-s.WaitForStatus(synchronizer.SynchronizerStatusRunning)
	log.Info("Synchronizer started successfully")

	// Wait for shutdown signal
	<-ctx.Done()

	// maybe more shutdown logic here

	// Wait for server to actually stop before exiting the program
	<-util.WaitForStatus(server.GetStatus, network_server.NetworkStopped, serverStatusCh, cleanupServerWait, 0)
	log.Info("Server stopped gracefully")
}
