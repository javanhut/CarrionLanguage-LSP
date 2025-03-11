package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/carrionlang-lsp/lsp/internal/handler"
	"github.com/carrionlang-lsp/lsp/internal/langserver"
	"github.com/carrionlang-lsp/lsp/internal/util"
	"go.lsp.dev/jsonrpc2"
)

var (
	version = "dev"
	addr    = flag.String(
		"addr",
		"localhost:7777",
		"TCP address to listen on (e.g., localhost:7777)",
	)
	stdio   = flag.Bool("stdio", false, "Use stdio for communication (default: TCP)")
	logFile = flag.String("log", "", "Path to log file (default: disabled)")
)

func main() {
	flag.Parse()
	// Configure logging
	var logWriter util.LogWriter = util.StderrLogger{}
	if *logFile != "" {
		dir := filepath.Dir(*logFile)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			log.Fatalf("Failed to create log directory: %v", err)
		}
		f, err := os.Create(*logFile)
		if err != nil {
			log.Fatalf("Failed to create log file: %v", err)
		}
		defer f.Close()
		logWriter = util.FileLogger{File: f}
	}
	logger := util.NewLogger(logWriter)
	logger.Info("Starting Carrion Language Server %s", version)

	if *stdio {
		logger.Info("Using stdio for JSON-RPC communication")
		stream := jsonrpc2.NewStream(util.NewStdioReadWriteCloser())
		runServer(context.Background(), stream, logger)
		return
	}

	logger.Info("Using TCP for JSON-RPC communication, listening at %s", *addr)
	err := langserver.RunTCPServer(
		context.Background(),
		*addr,
		func(ctx context.Context, conn jsonrpc2.Conn) {
			h := handler.NewHandler(logger, conn)

			// Create a function adapter that matches jsonrpc2.Handler type
			handlerFunc := func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
				result, err := h.Handle(ctx, req)
				return reply(ctx, result, err)
			}

			conn.Go(ctx, handlerFunc)
			<-conn.Done()
			if err := conn.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Connection closed with error: %v\n", err)
			}
		},
	)
	if err != nil {
		logger.Error("Server error: %v", err)
		os.Exit(1)
	}
}

func runServer(ctx context.Context, stream jsonrpc2.Stream, logger *util.Logger) {
	conn := jsonrpc2.NewConn(stream)
	h := handler.NewHandler(logger, conn)

	// Create a function adapter that matches jsonrpc2.Handler type
	handlerFunc := func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		result, err := h.Handle(ctx, req)
		return reply(ctx, result, err)
	}

	conn.Go(ctx, handlerFunc)
	<-conn.Done()
	if err := conn.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Connection closed with error: %v\n", err)
	}
}
