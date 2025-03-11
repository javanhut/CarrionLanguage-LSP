package protocol

import (
	"github.com/carrionlang-lsp/lsp/internal/util"
	lsp "go.lsp.dev/protocol"
)

// CarrionDocument represents a Carrion source file that is being edited
type CarrionDocument struct {
	URI        lsp.DocumentURI
	Text       string
	Version    int32
	LanguageID string
}

// CarrionDocumentStore keeps track of all open documents
type CarrionDocumentStore struct {
	Documents map[lsp.DocumentURI]*CarrionDocument
	Logger    *util.Logger
}

// NewDocumentStore creates a new document store
func NewDocumentStore(logger *util.Logger) *CarrionDocumentStore {
	return &CarrionDocumentStore{
		Documents: make(map[lsp.DocumentURI]*CarrionDocument),
		Logger:    logger,
	}
}

// GetDocument retrieves a document by URI
func (s *CarrionDocumentStore) GetDocument(uri lsp.DocumentURI) *CarrionDocument {
	doc, ok := s.Documents[uri]
	if !ok {
		return nil
	}
	return doc
}

// AddDocument adds a document to the store
func (s *CarrionDocumentStore) AddDocument(
	uri lsp.DocumentURI,
	languageID string,
	text string,
	version int32,
) *CarrionDocument {
	doc := &CarrionDocument{
		URI:        uri,
		Text:       text,
		Version:    version,
		LanguageID: languageID,
	}
	s.Documents[uri] = doc
	s.Logger.Debug("Added document: %s (version: %d)", uri, version)
	return doc
}

// UpdateDocument updates a document's content
func (s *CarrionDocumentStore) UpdateDocument(
	uri lsp.DocumentURI,
	changes []lsp.TextDocumentContentChangeEvent,
	version int32,
) *CarrionDocument {
	doc := s.GetDocument(uri)
	if doc == nil {
		s.Logger.Error("Cannot update non-existent document: %s", uri)
		return nil
	}

	// Update the document based on the changes
	if len(changes) == 0 {
		s.Logger.Warn("Document update with no changes: %s", uri)
		return doc
	}

	// Handle full document update
	// Note: We're only supporting full content updates for now
	doc.Text = changes[0].Text
	doc.Version = version
	s.Logger.Debug("Full update of document: %s (version: %d)", uri, version)
	return doc
}

// RemoveDocument removes a document from the store
func (s *CarrionDocumentStore) RemoveDocument(uri lsp.DocumentURI) {
	if _, ok := s.Documents[uri]; ok {
		delete(s.Documents, uri)
		s.Logger.Debug("Removed document: %s", uri)
	}
}

// DiagnosticSeverity maps error types to LSP severity levels
var DiagnosticSeverity = map[string]lsp.DiagnosticSeverity{
	"error":   lsp.DiagnosticSeverityError,
	"warning": lsp.DiagnosticSeverityWarning,
	"info":    lsp.DiagnosticSeverityInformation,
	"hint":    lsp.DiagnosticSeverityHint,
}

// CompletionItemKind maps Carrion symbol types to LSP completion item kinds
var CompletionItemKind = map[string]lsp.CompletionItemKind{
	"keyword":   lsp.CompletionItemKindKeyword,
	"spellbook": lsp.CompletionItemKindClass,
	"spell":     lsp.CompletionItemKindFunction,
	"variable":  lsp.CompletionItemKindVariable,
	"field":     lsp.CompletionItemKindField,
	"method":    lsp.CompletionItemKindMethod,
	"const":     lsp.CompletionItemKindConstant,
	"import":    lsp.CompletionItemKindModule,
}
