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

// Build-time variables set by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const (
	serverName = "mcp-k8s"
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
		fmt.Printf("%s %s\n", serverName, version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built: %s\n", date)
		os.Exit(0)
	}

	// Initialize the MCP server
	s := server.NewMCPServer(
		serverName,
		version,
		server.WithInstructions(`
This MCP server provides safe, read-only access to Kubernetes clusters through structured tools and resources.

**Key Features:**
- Safe by design: All operations are read-only, no cluster modifications possible
- No kubectl required: Direct API access through kubeconfig contexts
- Context discovery: Use 'kubeconfig://contexts' MCP resource to find available clusters
- Comprehensive analysis: Built-in prompts for memory pressure and workload instability analysis

**Available Tools:**
- list_k8s_resources: List and filter Kubernetes resources with smart formatting
- list_k8s_api_resources: Discover available API resource types (like kubectl api-resources)
- get_k8s_resource: Fetch individual resources with optional Go template formatting
- get_k8s_metrics: Get CPU/memory metrics for nodes and pods (like kubectl top)
- get_k8s_pod_logs: Retrieve pod logs with filtering options

**Context Usage:**
Instead of running kubectl commands, use the kubeconfig://contexts MCP resource to discover available cluster contexts. This server resolves cluster aliases (like 'prod', 'staging') to actual kubeconfig contexts automatically.

**Analysis Prompts:**
- memory_pressure_analysis: Systematic analysis of pod memory usage and OOM issues
- workload_instability_analysis: Investigation of Events and logs for instability patterns

All tools support CRDs and custom resources automatically through dynamic client discovery.`),
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
		fmt.Fprintf(os.Stderr, "Starting MCP server %s %s\n", serverName, version)
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
