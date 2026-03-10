package tools

import "github.com/mark3labs/mcp-go/mcp"

func readOnlyToolOptions(opts ...mcp.ToolOption) []mcp.ToolOption {
	return append([]mcp.ToolOption{
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
	}, opts...)
}
