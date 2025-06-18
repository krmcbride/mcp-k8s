package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/server"

	"github.com/krmcbride/mcp-k8s/internal/prompts"
	"github.com/krmcbride/mcp-k8s/internal/resources"
	"github.com/krmcbride/mcp-k8s/internal/tools"
)

const (
	serverName    = "mcp-k8s"
	serverVersion = "0.0.0-dev"
)

// WARN: only log to stderr to prevent interference with stdio transport
// See: https://modelcontextprotocol.io/docs/tools/debugging#implementing-logging
func main() {
	var showHelp bool
	var showVersion bool

	flag.BoolVar(&showHelp, "help", false, "Show help information")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.Parse()

	if showHelp {
		fmt.Printf("%s - MCP server for Kubernetes cluster interaction\n\n", serverName)
		fmt.Println("This is an MCP (Model Context Protocol) server that provides tools for")
		fmt.Println("interacting with Kubernetes clusters. It exposes Kubernetes operations")
		fmt.Println("through MCP tools that can be used by Claude and other MCP clients.")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Printf("  %s [options]\n\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("The server runs over stdio and communicates using the MCP protocol.")
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("%s %s\n", serverName, serverVersion)
		os.Exit(0)
	}

	// Initialize the MCP server
	s := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(false, false),
		server.WithPromptCapabilities(false),
		server.WithRecovery(),
	)

	// Register prompts, resources, and tools
	prompts.RegisterMCPPrompts(s)
	resources.RegisterMCPResources(s)
	tools.RegisterMCPTools(s)

	// Set up signal handling
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel to receive server errors
	errChan := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		fmt.Fprintf(os.Stderr, "Starting MCP server %s %s\n", serverName, serverVersion)
		if err := server.ServeStdio(s); err != nil {
			errChan <- err
		}
	}()

	// Wait for either a signal or an error
	select {
	case sig := <-sigChan:
		fmt.Fprintf(os.Stderr, "Received signal %v, shutting down gracefully...\n", sig)
		cancel()

		// Give the server a moment to clean up
		time.Sleep(100 * time.Millisecond)

	case err := <-errChan:
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		cancel()
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Server shutdown complete\n")
}
