package main

import (
	"flag"
	"os"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/auth"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/synchronizer"
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
	flag.StringVar(&opts.sseEndpoint, "sse", opts.sseEndpoint, "SSE endpoint path")
	flag.StringVar(&opts.apiEndpoint, "api", opts.apiEndpoint, "API endpoint path")
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

	// Open database engine
	dbEngine, err := storage_engine.OpenStorageEngine(&storage_engine.StorageEngineOptions{
		Path: opts.dbPath,
	})
	if err != nil {
		log.Error("Unable to open database: %v", err)
		os.Exit(1)
	}

	<-dbEngine.WaitForStatus(storage_engine.StorageEngineStatusOpen)
	log.Info("Database opened successfully")

	server := network_server.NewRapierDbHTTPServer(&network_server.RapierDbHTTPServerOption{
		Addr:        opts.addr,
		SseEndpoint: opts.sseEndpoint,
		ApiEndpoint: opts.apiEndpoint,
	})
	server.SetAuthProvider(&auth.HttpMockAuthProvider{})

	log.Info("Starting server...")
	err = server.Start()
	if err != nil {
		log.Error("Failed to start server: %v", err)
		os.Exit(1)
	}
	<-server.WaitForStatus(network_server.NetworkRunning)

	// Start synchronizer
	channel := server.GetChannel()
	synchronizerConfig := &synchronizer.SynchronizerConfig{}
	s := synchronizer.NewSynchronizer(dbEngine, channel, synchronizerConfig)
	err = s.Start()
	if err != nil {
		log.Error("Failed to start synchronizer: %v", err)
		os.Exit(1)
	}
	<-s.WaitForStatus(synchronizer.SynchronizerStatusRunning)
	log.Info("Synchronizer started successfully")

	<-server.WaitForStatus(network_server.NetworkStopped)
	log.Info("Server stopped")
}
