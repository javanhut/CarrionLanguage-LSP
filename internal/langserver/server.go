package langserver

import (
	"context"
	"fmt"
	"net"
	"os"

	"go.lsp.dev/jsonrpc2"
)

// HandlerFunc is a function that handles a connection
type HandlerFunc func(ctx context.Context, conn jsonrpc2.Conn)

// RunTCPServer starts a TCP server for JSON-RPC communication
func RunTCPServer(ctx context.Context, addr string, handler HandlerFunc) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	defer ln.Close()

	fmt.Fprintf(os.Stderr, "Carrion Language Server listening on %s\n", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %w", err)
		}

		stream := jsonrpc2.NewStream(conn)
		jsonConn := jsonrpc2.NewConn(stream)

		go handler(ctx, jsonConn)
	}
}
