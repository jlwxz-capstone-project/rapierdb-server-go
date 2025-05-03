package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
)

// create-db <db-path> <db-schema> <db-permissions>
func main() {
	// Define command line arguments
	var (
		dbPath      string
		schemaPath  string
		permPath    string
		showHelp    bool
		showVersion bool
	)

	// Set up command line flags
	flag.StringVar(&dbPath, "path", "", "Database storage path")
	flag.StringVar(&schemaPath, "schema", "", "Database schema file path")
	flag.StringVar(&permPath, "permissions", "", "Database permissions file path")
	flag.BoolVar(&showHelp, "help", false, "Show help information")
	flag.BoolVar(&showVersion, "version", false, "Show version information")

	// Parse command line arguments
	flag.Parse()

	// Show help information
	if showHelp {
		fmt.Println("Usage: create-db -path <db-path> -schema <db-schema> -permissions <db-permissions>")
		flag.PrintDefaults()
		return
	}

	// Show version information
	if showVersion {
		fmt.Println("RapierDB Creation Tool v0.1.0")
		return
	}

	// Check required parameters
	if dbPath == "" || schemaPath == "" || permPath == "" {
		fmt.Println("Error: Database path, schema file, and permissions file must be provided")
		fmt.Println("Usage: create-db -path <db-path> -schema <db-schema> -permissions <db-permissions>")
		os.Exit(1)
	}

	// Read schema file
	schemaContent, err := os.ReadFile(schemaPath)
	if err != nil {
		fmt.Printf("Failed to read schema file: %v\n", err)
		os.Exit(1)
	}

	// Read permissions file
	permContent, err := os.ReadFile(permPath)
	if err != nil {
		fmt.Printf("Failed to read permissions file: %v\n", err)
		os.Exit(1)
	}

	// Parse schema
	dbSchema, err := storage_engine.NewDatabaseSchemaFromJs(string(schemaContent))
	if err != nil {
		fmt.Printf("Failed to parse database schema: %v\n", err)
		os.Exit(1)
	}

	// Create database
	err = storage_engine.CreateNewDatabase(dbPath, dbSchema, string(permContent))
	if err != nil {
		fmt.Printf("Failed to create database: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Database successfully created at %s\n", dbPath)
}
