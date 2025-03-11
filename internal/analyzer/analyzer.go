package analyzer

import (
	"fmt"
	"strings"

	"github.com/carrionlang-lsp/lsp/internal/protocol"
	"github.com/carrionlang-lsp/lsp/internal/symbols"
	"github.com/carrionlang-lsp/lsp/internal/util"

	"github.com/javanhut/TheCarrionLanguage/src/lexer"
	"github.com/javanhut/TheCarrionLanguage/src/parser"

	lsp "go.lsp.dev/protocol"
)

// CarrionAnalyzer provides language analysis for Carrion language files
type CarrionAnalyzer struct {
	logger        *util.Logger
	documentStore *protocol.CarrionDocumentStore
	symbolTable   *symbols.SymbolTable
}

// NewCarrionAnalyzer creates a new analyzer
func NewCarrionAnalyzer(
	logger *util.Logger,
	docStore *protocol.CarrionDocumentStore,
) *CarrionAnalyzer {
	return &CarrionAnalyzer{
		logger:        logger,
		documentStore: docStore,
		symbolTable:   symbols.NewSymbolTable(),
	}
}

// AnalyzeDocument analyzes a document and returns diagnostics
func (a *CarrionAnalyzer) AnalyzeDocument(uri lsp.DocumentURI) []lsp.Diagnostic {
	doc := a.documentStore.GetDocument(uri)
	if doc == nil {
		a.logger.Warn("Cannot analyze non-existent document: %s", uri)
		return nil
	}

	// Parse the document
	l := lexer.New(doc.Text)
	p := parser.New(l)
	program := p.ParseProgram()

	// Collect diagnostics from parser errors
	diagnostics := make([]lsp.Diagnostic, 0)

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			// Extract line and column information if available
			lineNum, colNum := 1, 1
			errMsg := err

			// Try to extract line/column info using a simple heuristic
			// This would be more robust in a real implementation
			if parts := strings.Split(err, " at line "); len(parts) > 1 {
				if locParts := strings.Split(parts[1], ", column "); len(locParts) > 1 {
					lineNumStr := locParts[0]
					colNumStr := strings.Split(locParts[1], " ")[0]

					// Attempt to parse as integers
					if ln, err := util.ParseInt(lineNumStr); err == nil {
						lineNum = ln
					}
					if cn, err := util.ParseInt(colNumStr); err == nil {
						colNum = cn
					}

					// Clean up the error message
					errMsg = parts[0]
				}
			}

			// Create a diagnostic for the error
			diagnostic := lsp.Diagnostic{
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      uint32(lineNum - 1),
						Character: uint32(colNum - 1),
					},
					End: lsp.Position{
						Line:      uint32(lineNum - 1),
						Character: uint32(colNum + 10), // Estimate end position
					},
				},
				Severity: protocol.DiagnosticSeverity["error"],
				Source:   "carrion-lsp",
				Message:  errMsg,
			}

			diagnostics = append(diagnostics, diagnostic)
		}
	}

	// Build symbol table for the document
	if len(p.Errors()) == 0 {
		a.symbolTable.BuildFromAST(program, string(uri))
	}

	return diagnostics
}

// GetCompletions returns completion items at the given position
func (a *CarrionAnalyzer) GetCompletions(
	uri lsp.DocumentURI,
	position lsp.Position,
) []lsp.CompletionItem {
	doc := a.documentStore.GetDocument(uri)
	if doc == nil {
		a.logger.Warn("Cannot get completions for non-existent document: %s", uri)
		return nil
	}

	// Get the current line and position context
	lines := strings.Split(doc.Text, "\n")
	if int(position.Line) >= len(lines) {
		return nil
	}

	line := lines[position.Line]
	if int(position.Character) > len(line) {
		position.Character = uint32(len(line))
	}

	// Get text before cursor position to determine context
	textBeforeCursor := line[:position.Character]

	// Determine if we're after a dot (e.g., object.???)
	isDotCompletion := false
	objectName := ""
	if idx := strings.LastIndex(textBeforeCursor, "."); idx >= 0 {
		isDotCompletion = true
		prefix := strings.TrimSpace(textBeforeCursor[:idx])
		// Extract the object name
		fields := strings.Fields(prefix)
		if len(fields) > 0 {
			objectName = fields[len(fields)-1]
		}
	}

	// Collect completion items based on context
	var completionItems []lsp.CompletionItem

	if isDotCompletion {
		// Add method/field completions for the object
		if objectName == "self" {
			// Handle self completions (methods and fields of current spellbook)
			selfCompletions := a.getSelfCompletions(uri, position)
			completionItems = append(completionItems, selfCompletions...)
		} else {
			// Add completions for other objects based on symbol table
			objectCompletions := a.getObjectCompletions(uri, objectName)
			completionItems = append(completionItems, objectCompletions...)
		}
	} else {
		// Add keyword completions
		keywordCompletions := a.getKeywordCompletions()
		completionItems = append(completionItems, keywordCompletions...)

		// Add local variable completions
		localCompletions := a.getLocalCompletions(uri, position)
		completionItems = append(completionItems, localCompletions...)

		// Add global completions (spellbooks, etc.)
		globalCompletions := a.getGlobalCompletions()
		completionItems = append(completionItems, globalCompletions...)
	}

	return completionItems
}

// getSelfCompletions returns completions for 'self.' context
func (a *CarrionAnalyzer) getSelfCompletions(
	uri lsp.DocumentURI,
	position lsp.Position,
) []lsp.CompletionItem {
	// Find current spellbook context
	currentSpellbook := a.symbolTable.GetCurrentSpellbook(string(uri), int(position.Line))
	if currentSpellbook == nil {
		return nil
	}

	completions := []lsp.CompletionItem{}

	// Add spellbook methods
	for _, method := range currentSpellbook.Methods {
		completions = append(completions, lsp.CompletionItem{
			Label:         method.Name,
			Kind:          lsp.CompletionItemKindMethod,
			Detail:        "method of " + currentSpellbook.Name,
			Documentation: method.Documentation,
		})
	}

	// Add spellbook fields
	for _, field := range currentSpellbook.Fields {
		completions = append(completions, lsp.CompletionItem{
			Label:         field.Name,
			Kind:          lsp.CompletionItemKindField,
			Detail:        "field of " + currentSpellbook.Name,
			Documentation: field.Documentation,
		})
	}

	return completions
}

// getObjectCompletions returns completions for 'object.' context
func (a *CarrionAnalyzer) getObjectCompletions(
	uri lsp.DocumentURI,
	objectName string,
) []lsp.CompletionItem {
	// Find the object's type in the symbol table
	symbol := a.symbolTable.LookupSymbol(objectName, string(uri))
	if symbol == nil {
		return nil
	}

	completions := []lsp.CompletionItem{}

	// Get the object's type (spellbook)
	if symbol.Type == "instance" || symbol.Type == "variable" {
		spellbook := a.symbolTable.LookupSpellbook(symbol.SpellbookName)
		if spellbook != nil {
			// Add methods
			for _, method := range spellbook.Methods {
				completions = append(completions, lsp.CompletionItem{
					Label:         method.Name,
					Kind:          lsp.CompletionItemKindMethod,
					Detail:        "method of " + spellbook.Name,
					Documentation: method.Documentation,
				})
			}

			// Add fields
			for _, field := range spellbook.Fields {
				completions = append(completions, lsp.CompletionItem{
					Label:         field.Name,
					Kind:          lsp.CompletionItemKindField,
					Detail:        "field of " + spellbook.Name,
					Documentation: field.Documentation,
				})
			}
		}
	}

	return completions
}

// getKeywordCompletions returns Carrion language keyword completions
func (a *CarrionAnalyzer) getKeywordCompletions() []lsp.CompletionItem {
	keywords := []string{
		"spell", "spellbook", "init", "self", "if", "else", "otherwise",
		"for", "in", "while", "stop", "skip", "ignore", "return",
		"import", "match", "case", "attempt", "resolve", "ensnare",
		"raise", "as", "arcane", "arcanespell", "super",
		"check", "True", "False", "None", "and", "or", "not",
	}

	completions := []lsp.CompletionItem{}

	for _, keyword := range keywords {
		completions = append(completions, lsp.CompletionItem{
			Label: keyword,
			Kind:  lsp.CompletionItemKindKeyword,
		})
	}

	return completions
}

// getLocalCompletions returns local variable completions
func (a *CarrionAnalyzer) getLocalCompletions(
	uri lsp.DocumentURI,
	position lsp.Position,
) []lsp.CompletionItem {
	// Get local variables from the symbol table
	locals := a.symbolTable.GetLocalSymbols(string(uri), int(position.Line))

	completions := []lsp.CompletionItem{}

	for _, symbol := range locals {
		kind := lsp.CompletionItemKindVariable
		if symbol.Type == "parameter" {
			kind = lsp.CompletionItemKindVariable
		}

		detail := symbol.Type
		if symbol.SpellbookName != "" {
			detail = detail + " of " + symbol.SpellbookName
		}

		completions = append(completions, lsp.CompletionItem{
			Label:         symbol.Name,
			Kind:          kind,
			Detail:        detail,
			Documentation: symbol.Documentation,
		})
	}

	return completions
}

// getGlobalCompletions returns global symbol completions
func (a *CarrionAnalyzer) getGlobalCompletions() []lsp.CompletionItem {
	// Get all spellbooks and global functions from the symbol table
	globals := a.symbolTable.GetGlobalSymbols()

	completions := []lsp.CompletionItem{}

	for _, symbol := range globals {
		var kind lsp.CompletionItemKind

		switch symbol.Type {
		case "spellbook":
			kind = lsp.CompletionItemKindClass
		case "spell":
			kind = lsp.CompletionItemKindFunction
		case "variable":
			kind = lsp.CompletionItemKindVariable
		case "const":
			kind = lsp.CompletionItemKindConstant
		default:
			kind = lsp.CompletionItemKindVariable
		}

		completions = append(completions, lsp.CompletionItem{
			Label:         symbol.Name,
			Kind:          kind,
			Detail:        symbol.Type,
			Documentation: symbol.Documentation,
		})
	}

	return completions
}

// FindDefinition returns the location of the definition for a symbol at the given position
func (a *CarrionAnalyzer) FindDefinition(
	uri lsp.DocumentURI,
	position lsp.Position,
) []lsp.Location {
	doc := a.documentStore.GetDocument(uri)
	if doc == nil {
		a.logger.Warn("Cannot find definition in non-existent document: %s", uri)
		return nil
	}

	// Get the word at the current position
	lines := strings.Split(doc.Text, "\n")
	if int(position.Line) >= len(lines) {
		return nil
	}

	line := lines[position.Line]
	if int(position.Character) >= len(line) {
		return nil
	}

	// Extract the symbol name at the current position
	symbolName, _ := a.getSymbolAtPosition(line, position.Character)
	if symbolName == "" {
		return nil
	}

	// Look up the symbol in the symbol table
	symbol := a.symbolTable.LookupSymbol(symbolName, string(uri))
	if symbol == nil {
		return nil
	}

	// Create a location for the definition
	locations := []lsp.Location{
		{
			URI: lsp.DocumentURI(symbol.DefinitionURI),
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      uint32(symbol.DefinitionLine),
					Character: uint32(symbol.DefinitionColumn),
				},
				End: lsp.Position{
					Line:      uint32(symbol.DefinitionLine),
					Character: uint32(symbol.DefinitionColumn + len(symbolName)),
				},
			},
		},
	}

	return locations
}

// getSymbolAtPosition extracts the symbol name at the given position in a line
func (a *CarrionAnalyzer) getSymbolAtPosition(line string, charPos uint32) (string, lsp.Range) {
	if int(charPos) >= len(line) {
		return "", lsp.Range{}
	}

	// Find the start of the symbol (non-whitespace, non-punctuation)
	start := int(charPos)
	for start > 0 && isIdentifierChar(line[start-1]) {
		start--
	}

	// Find the end of the symbol
	end := int(charPos)
	for end < len(line) && isIdentifierChar(line[end]) {
		end++
	}

	// Extract the symbol name
	if start < end {
		symbolRange := lsp.Range{
			Start: lsp.Position{Character: uint32(start)},
			End:   lsp.Position{Character: uint32(end)},
		}
		return line[start:end], symbolRange
	}

	return "", lsp.Range{}
}

// isIdentifierChar returns true if the character is valid in an identifier
func isIdentifierChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// Replace the Hover function in analyzer.go with this version:

// GetHoverInfo returns hover information for a symbol at the given position
func (a *CarrionAnalyzer) GetHoverInfo(uri lsp.DocumentURI, position lsp.Position) *lsp.Hover {
	doc := a.documentStore.GetDocument(uri)
	if doc == nil {
		a.logger.Warn("Cannot get hover info for non-existent document: %s", uri)
		return nil
	}

	// Get the word at the current position
	lines := strings.Split(doc.Text, "\n")
	if int(position.Line) >= len(lines) {
		return nil
	}

	line := lines[position.Line]
	if int(position.Character) >= len(line) {
		return nil
	}

	// Extract the symbol name at the current position
	symbolName, symbolRange := a.getSymbolAtPosition(line, position.Character)
	if symbolName == "" {
		return nil
	}

	// Check if it's a keyword
	if isCarrionKeyword(symbolName) {
		return &lsp.Hover{
			Contents: formatDocumentation(formatKeywordDocumentation(symbolName)),
			Range:    &symbolRange,
		}
	}

	// Look up the symbol in the symbol table
	symbol := a.symbolTable.LookupSymbol(symbolName, string(uri))
	if symbol == nil {
		return nil
	}

	// Create hover content based on symbol type
	var content string
	switch symbol.Type {
	case "spellbook":
		content = fmt.Sprintf("**spellbook** %s\n\n%s", symbol.Name, symbol.Documentation)
	case "spell":
		var params string
		if symbol.Parameters != nil {
			paramStrs := make([]string, 0, len(symbol.Parameters))
			for _, param := range symbol.Parameters {
				paramStrs = append(paramStrs, param.Name)
			}
			params = strings.Join(paramStrs, ", ")
		}
		content = fmt.Sprintf("**spell** %s(%s)\n\n%s", symbol.Name, params, symbol.Documentation)
	case "variable":
		content = fmt.Sprintf(
			"**variable** %s: %s\n\n%s",
			symbol.Name,
			symbol.ValueType,
			symbol.Documentation,
		)
	case "field":
		content = fmt.Sprintf(
			"**field** %s of %s\n\n%s",
			symbol.Name,
			symbol.SpellbookName,
			symbol.Documentation,
		)
	case "method":
		var params string
		if symbol.Parameters != nil {
			paramStrs := make([]string, 0, len(symbol.Parameters))
			for _, param := range symbol.Parameters {
				paramStrs = append(paramStrs, param.Name)
			}
			params = strings.Join(paramStrs, ", ")
		}
		content = fmt.Sprintf(
			"**method** %s.%s(%s)\n\n%s",
			symbol.SpellbookName,
			symbol.Name,
			params,
			symbol.Documentation,
		)
	default:
		content = fmt.Sprintf("**%s** %s\n\n%s", symbol.Type, symbol.Name, symbol.Documentation)
	}

	return &lsp.Hover{
		Contents: formatDocumentation(content),
		Range:    &symbolRange,
	}
}

// formatDocumentation creates a MarkupContent with Markdown
func formatDocumentation(content string) lsp.MarkupContent {
	return lsp.MarkupContent{
		Kind:  "markdown", // use string literal instead of constant
		Value: content,
	}
}

// isCarrionKeyword checks if a string is a Carrion language keyword
func isCarrionKeyword(word string) bool {
	keywords := map[string]bool{
		"spell": true, "spellbook": true, "init": true, "self": true,
		"if": true, "else": true, "otherwise": true, "for": true,
		"in": true, "while": true, "stop": true, "skip": true,
		"ignore": true, "return": true, "import": true, "match": true,
		"case": true, "attempt": true, "resolve": true, "ensnare": true,
		"raise": true, "as": true, "arcane": true, "arcanespell": true,
		"super": true, "check": true, "True": true, "False": true,
		"None": true, "and": true, "or": true, "not": true,
	}

	return keywords[word]
}

// formatKeywordDocumentation returns markdown documentation for a keyword
func formatKeywordDocumentation(keyword string) string {
	docs := map[string]string{
		"spell":       "**spell** - Defines a function in Carrion.",
		"spellbook":   "**spellbook** - Defines a class in Carrion.",
		"init":        "**init** - Special method for initializing a spellbook instance.",
		"self":        "**self** - References the current instance within a spellbook method.",
		"if":          "**if** - Conditional statement that executes a block if the condition is true.",
		"else":        "**else** - Alternative block of an if statement.",
		"otherwise":   "**otherwise** - Alternative condition in an if-chain (similar to 'else if').",
		"for":         "**for** - Loop that iterates over a sequence.",
		"in":          "**in** - Used with 'for' to specify the sequence to iterate over.",
		"while":       "**while** - Loop that executes as long as a condition is true.",
		"stop":        "**stop** - Exits a loop (similar to 'break').",
		"skip":        "**skip** - Skips to the next iteration of a loop (similar to 'continue').",
		"ignore":      "**ignore** - No-operation statement.",
		"return":      "**return** - Returns a value from a function.",
		"import":      "**import** - Imports functionality from another module.",
		"match":       "**match** - Pattern matching statement (similar to 'switch').",
		"case":        "**case** - Defines a pattern in a match statement.",
		"attempt":     "**attempt** - Try block for exception handling.",
		"resolve":     "**resolve** - Always executed after an attempt block (similar to 'finally').",
		"ensnare":     "**ensnare** - Catches exceptions in an attempt block (similar to 'catch').",
		"raise":       "**raise** - Throws an exception.",
		"as":          "**as** - Alias for imports or caught exceptions.",
		"arcane":      "**arcane** - Defines an abstract spellbook (similar to 'abstract class').",
		"arcanespell": "**arcanespell** - Defines an abstract method in a spellbook.",
		"super":       "**super** - References the parent spellbook's implementation.",
		"check":       "**check** - Assertion statement.",
		"True":        "**True** - Boolean true value.",
		"False":       "**False** - Boolean false value.",
		"None":        "**None** - Null value.",
		"and":         "**and** - Logical AND operator.",
		"or":          "**or** - Logical OR operator.",
		"not":         "**not** - Logical NOT operator.",
	}

	if doc, ok := docs[keyword]; ok {
		return doc
	}

	return fmt.Sprintf("**%s** - Carrion keyword", keyword)
}
