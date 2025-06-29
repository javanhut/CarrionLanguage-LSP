package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/carrionlang-lsp/lsp/internal/analyzer"
	"github.com/carrionlang-lsp/lsp/internal/formatter"
	"github.com/carrionlang-lsp/lsp/internal/protocol"
	"github.com/carrionlang-lsp/lsp/internal/util"

	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
)

// ErrorCodes defined in the JSON-RPC spec
const (
	CodeMethodNotFound = -32601
)

type Handler struct {
	conn          jsonrpc2.Conn
	logger        *util.Logger
	documentStore *protocol.CarrionDocumentStore
	analyzer      *analyzer.CarrionAnalyzer
	formatter     *formatter.CarrionFormatter
	capabilities  lsp.ServerCapabilities
	workspace     lsp.WorkspaceFolder
	initialized   bool
}

func NewHandler(logger *util.Logger, conn jsonrpc2.Conn) *Handler {
	docStore := protocol.NewDocumentStore(logger)
	analyzer := analyzer.NewCarrionAnalyzer(logger, docStore)
	formatter := formatter.NewCarrionFormatter(logger)

	return &Handler{
		conn:          conn,
		logger:        logger,
		documentStore: docStore,
		analyzer:      analyzer,
		formatter:     formatter,
		initialized:   false,
	}
}

// Handle implements jsonrpc2.Handler interface
func (h *Handler) Handle(
	ctx context.Context,
	req jsonrpc2.Request,
) (result interface{}, err error) {
	h.logger.Debug("Received request: %s", req.Method())

	// Allow initialize even if not initialized yet
	if !h.initialized && req.Method() != "initialize" {
		return nil, fmt.Errorf("server not initialized")
	}

	switch req.Method() {
	case "initialize":
		return h.handleInitialize(ctx, req)
	case "initialized":
		return h.handleInitialized(ctx, req)
	case "shutdown":
		return h.handleShutdown(ctx, req)
	case "exit":
		return h.handleExit(ctx, req)
	case "textDocument/didOpen":
		return h.handleTextDocumentDidOpen(ctx, req)
	case "textDocument/didChange":
		return h.handleTextDocumentDidChange(ctx, req)
	case "textDocument/didClose":
		return h.handleTextDocumentDidClose(ctx, req)
	case "textDocument/completion":
		return h.handleTextDocumentCompletion(ctx, req)
	case "textDocument/formatting":
		return h.handleTextDocumentFormatting(ctx, req)
	case "textDocument/definition":
		return h.handleTextDocumentDefinition(ctx, req)
	case "textDocument/hover":
		return h.handleTextDocumentHover(ctx, req)
	case "textDocument/signatureHelp":
		return h.handleTextDocumentSignatureHelp(ctx, req)
	default:
		h.logger.Warn("Unsupported method: %s", req.Method())
		return nil, &jsonrpc2.Error{
			Code:    CodeMethodNotFound,
			Message: fmt.Sprintf("method not supported: %s", req.Method()),
		}
	}
}

func (h *Handler) handleInitialize(
	ctx context.Context,
	req jsonrpc2.Request,
) (interface{}, error) {
	h.logger.Info("Initializing Carrion Language Server")

	var params lsp.InitializeParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, err
	}

	if params.WorkspaceFolders != nil && len(params.WorkspaceFolders) > 0 {
		h.workspace = params.WorkspaceFolders[0]
	}

	// Set server capabilities
	h.capabilities = lsp.ServerCapabilities{
		TextDocumentSync: &lsp.TextDocumentSyncOptions{
			OpenClose: true,
			Change:    lsp.TextDocumentSyncKindFull, // We currently only support full syncs
		},
		CompletionProvider: &lsp.CompletionOptions{
			TriggerCharacters: []string{".", ":"},
			ResolveProvider:   false,
		},
		HoverProvider:              true,
		DefinitionProvider:         true,
		DocumentFormattingProvider: true,
		SignatureHelpProvider: &lsp.SignatureHelpOptions{
			TriggerCharacters: []string{"(", ","},
		},
	}

	h.initialized = true

	return lsp.InitializeResult{
		Capabilities: h.capabilities,
		ServerInfo: &lsp.ServerInfo{
			Name:    "carrion-language-server",
			Version: "0.1.0",
		},
	}, nil
}

func (h *Handler) handleInitialized(
	ctx context.Context,
	req jsonrpc2.Request,
) (interface{}, error) {
	h.logger.Info("Server initialized")
	return nil, nil
}

func (h *Handler) handleShutdown(ctx context.Context, req jsonrpc2.Request) (interface{}, error) {
	h.logger.Info("Shutting down")
	h.initialized = false
	return nil, nil
}

func (h *Handler) handleExit(ctx context.Context, req jsonrpc2.Request) (interface{}, error) {
	h.logger.Info("Exiting")
	return nil, nil
}

func (h *Handler) handleTextDocumentDidOpen(
	ctx context.Context,
	req jsonrpc2.Request,
) (interface{}, error) {
	var params lsp.DidOpenTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, err
	}

	h.logger.Debug("Document opened: %s", params.TextDocument.URI)

	h.documentStore.AddDocument(
		params.TextDocument.URI,
		string(params.TextDocument.LanguageID),
		params.TextDocument.Text,
		params.TextDocument.Version,
	)

	// Run diagnostics on open
	diagnostics := h.analyzer.AnalyzeDocument(params.TextDocument.URI)
	h.sendDiagnostics(ctx, params.TextDocument.URI, diagnostics)

	return nil, nil
}

func (h *Handler) handleTextDocumentDidChange(
	ctx context.Context,
	req jsonrpc2.Request,
) (interface{}, error) {
	var params lsp.DidChangeTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, err
	}

	h.logger.Debug("Document changed: %s", params.TextDocument.URI)
	h.documentStore.UpdateDocument(
		params.TextDocument.URI,
		params.ContentChanges,
		params.TextDocument.Version,
	)

	// Run diagnostics on change
	diagnostics := h.analyzer.AnalyzeDocument(params.TextDocument.URI)
	h.sendDiagnostics(ctx, params.TextDocument.URI, diagnostics)

	return nil, nil
}

func (h *Handler) handleTextDocumentDidClose(
	ctx context.Context,
	req jsonrpc2.Request,
) (interface{}, error) {
	var params lsp.DidCloseTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, err
	}

	h.logger.Debug("Document closed: %s", params.TextDocument.URI)
	h.documentStore.RemoveDocument(params.TextDocument.URI)

	// Clear diagnostics for closed document
	h.sendDiagnostics(ctx, params.TextDocument.URI, nil)

	return nil, nil
}

func (h *Handler) handleTextDocumentCompletion(
	ctx context.Context,
	req jsonrpc2.Request,
) (interface{}, error) {
	var params lsp.CompletionParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, err
	}

	h.logger.Debug(
		"Completion requested at position %v in %s",
		params.Position,
		params.TextDocument.URI,
	)
	completions := h.analyzer.GetCompletions(params.TextDocument.URI, params.Position)

	return completions, nil
}

func (h *Handler) handleTextDocumentFormatting(
	ctx context.Context,
	req jsonrpc2.Request,
) (interface{}, error) {
	var params lsp.DocumentFormattingParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, err
	}

	h.logger.Debug("Formatting requested for %s", params.TextDocument.URI)

	doc := h.documentStore.GetDocument(params.TextDocument.URI)
	if doc == nil {
		return nil, fmt.Errorf("document not found: %s", params.TextDocument.URI)
	}

	edits := h.formatter.Format(doc)
	return edits, nil
}

func (h *Handler) handleTextDocumentDefinition(
	ctx context.Context,
	req jsonrpc2.Request,
) (interface{}, error) {
	var params lsp.DefinitionParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, err
	}

	h.logger.Debug(
		"Definition requested at position %v in %s",
		params.Position,
		params.TextDocument.URI,
	)

	locations := h.analyzer.FindDefinition(params.TextDocument.URI, params.Position)
	return locations, nil
}

func (h *Handler) handleTextDocumentHover(
	ctx context.Context,
	req jsonrpc2.Request,
) (interface{}, error) {
	var params lsp.HoverParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, err
	}

	h.logger.Debug("Hover requested at position %v in %s", params.Position, params.TextDocument.URI)

	hoverInfo := h.analyzer.GetHoverInfo(params.TextDocument.URI, params.Position)
	return hoverInfo, nil
}

func (h *Handler) handleTextDocumentSignatureHelp(
	ctx context.Context,
	req jsonrpc2.Request,
) (interface{}, error) {
	var params lsp.SignatureHelpParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, err
	}

	h.logger.Debug("Signature help requested at position %v in %s", params.Position, params.TextDocument.URI)

	signatureHelp := h.analyzer.GetSignatureHelp(params.TextDocument.URI, params.Position)
	return signatureHelp, nil
}

func (h *Handler) sendDiagnostics(
	ctx context.Context,
	uri lsp.DocumentURI,
	diagnostics []lsp.Diagnostic,
) {
	// Send diagnostics notification
	err := h.conn.Notify(ctx, "textDocument/publishDiagnostics", lsp.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
	if err != nil {
		h.logger.Error("Failed to publish diagnostics: %v", err)
	}
}
