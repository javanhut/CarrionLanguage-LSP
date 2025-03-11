package protocol

import (
	"github.com/carrionlang-lsp/lsp/internal/util"
	"go.lsp.dev/protocol"
)

// CarrionDocument represents a Carrion source file that is being edited
type CarrionDocument struct {
	URI        protocol.DocumentURI
	Text       string
	Version    int32
	LanguageID string
}

// CarrionDocumentStore keeps track of all open documents
type CarrionDocumentStore struct {
	Documents map[protocol.DocumentURI]*CarrionDocument
	Logger    *util.Logger
}

// NewDocumentStore creates a new document store
func NewDocumentStore(logger *util.Logger) *CarrionDocumentStore {
	return &CarrionDocumentStore{
		Documents: make(map[protocol.DocumentURI]*CarrionDocument),
		Logger:    logger,
	}
}

// GetDocument retrieves a document by URI
func (s *CarrionDocumentStore) GetDocument(uri protocol.DocumentURI) *CarrionDocument {
	doc, ok := s.Documents[uri]
	if !ok {
		return nil
	}
	return doc
}

// AddDocument adds a document to the store
func (s *CarrionDocumentStore) AddDocument(
	uri protocol.DocumentURI,
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
	uri protocol.DocumentURI,
	changes []protocol.TextDocumentContentChangeEvent,
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
	if changes[0].Range == nil {
		doc.Text = changes[0].Text
		doc.Version = version
		s.Logger.Debug("Full update of document: %s (version: %d)", uri, version)
		return doc
	}

	// Handle incremental updates (not implemented yet)
	s.Logger.Warn("Incremental updates not fully implemented yet")
	doc.Text = changes[0].Text // Treating as full update for now
	doc.Version = version

	return doc
}

// RemoveDocument removes a document from the store
func (s *CarrionDocumentStore) RemoveDocument(uri protocol.DocumentURI) {
	if _, ok := s.Documents[uri]; ok {
		delete(s.Documents, uri)
		s.Logger.Debug("Removed document: %s", uri)
	}
}

// DiagnosticSeverity maps error types to LSP severity levels
var DiagnosticSeverity = map[string]protocol.DiagnosticSeverity{
	"error":   protocol.DiagnosticSeverityError,
	"warning": protocol.DiagnosticSeverityWarning,
	"info":    protocol.DiagnosticSeverityInformation,
	"hint":    protocol.DiagnosticSeverityHint,
}

// CompletionItemKind maps Carrion symbol types to LSP completion item kinds
var CompletionItemKind = map[string]protocol.CompletionItemKind{
	"keyword":   protocol.CompletionItemKindKeyword,
	"spellbook": protocol.CompletionItemKindClass,
	"spell":     protocol.CompletionItemKindFunction,
	"variable":  protocol.CompletionItemKindVariable,
	"field":     protocol.CompletionItemKindField,
	"method":    protocol.CompletionItemKindMethod,
	"const":     protocol.CompletionItemKindConstant,
	"import":    protocol.CompletionItemKindModule,
}
