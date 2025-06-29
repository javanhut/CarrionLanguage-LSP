package analyzer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/javanhut/TheCarrionLanguage/src/ast"
	"github.com/javanhut/TheCarrionLanguage/src/lexer"
	"github.com/javanhut/TheCarrionLanguage/src/parser"
	lsp "go.lsp.dev/protocol"

	"github.com/carrionlang-lsp/lsp/internal/protocol"
	"github.com/carrionlang-lsp/lsp/internal/symbols"
	"github.com/carrionlang-lsp/lsp/internal/util"
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
			diagnostic := a.createDiagnosticFromError(err)
			diagnostics = append(diagnostics, diagnostic)
		}
	}

	// Additional semantic analysis when parsing succeeds
	if len(p.Errors()) == 0 && program != nil {
		// Check for undefined variables, unused imports, etc.
		semanticDiagnostics := a.performSemanticAnalysis(program, doc)
		diagnostics = append(diagnostics, semanticDiagnostics...)
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
			// Handle self completions (methods and fields of current Grimoire)
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

		// Add built-in function completions
		builtinCompletions := a.getBuiltinCompletions()
		completionItems = append(completionItems, builtinCompletions...)

		// Add local variable completions
		localCompletions := a.getLocalCompletions(uri, position)
		completionItems = append(completionItems, localCompletions...)

		// Add global completions (Grimoires, etc.)
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
	// Find current Grimoire context
	currentGrimoire := a.symbolTable.GetCurrentGrimoire(string(uri), int(position.Line))
	if currentGrimoire == nil {
		return nil
	}

	completions := []lsp.CompletionItem{}

	// Add Grimoire methods
	for _, method := range currentGrimoire.Methods {
		completions = append(completions, lsp.CompletionItem{
			Label:         method.Name,
			Kind:          lsp.CompletionItemKindMethod,
			Detail:        "method of " + currentGrimoire.Name,
			Documentation: method.Documentation,
		})
	}

	// Add Grimoire fields
	for _, field := range currentGrimoire.Fields {
		completions = append(completions, lsp.CompletionItem{
			Label:         field.Name,
			Kind:          lsp.CompletionItemKindField,
			Detail:        "field of " + currentGrimoire.Name,
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

	// Get the object's type (Grimoire)
	if symbol.Type == "instance" || symbol.Type == "variable" {
		Grimoire := a.symbolTable.LookupGrimoire(symbol.GrimoireName)
		if Grimoire != nil {
			// Add methods
			for _, method := range Grimoire.Methods {
				completions = append(completions, lsp.CompletionItem{
					Label:         method.Name,
					Kind:          lsp.CompletionItemKindMethod,
					Detail:        "method of " + Grimoire.Name,
					Documentation: method.Documentation,
				})
			}

			// Add fields
			for _, field := range Grimoire.Fields {
				completions = append(completions, lsp.CompletionItem{
					Label:         field.Name,
					Kind:          lsp.CompletionItemKindField,
					Detail:        "field of " + Grimoire.Name,
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
		"spell", "grim", "init", "self", "if", "else", "otherwise",
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

// getBuiltinCompletions returns built-in function completions
func (a *CarrionAnalyzer) getBuiltinCompletions() []lsp.CompletionItem {
	builtins := []struct {
		name   string
		detail string
		doc    string
	}{
		{"len", "len(object) -> int", "Returns the length of a string, array, tuple, or hash"},
		{"print", "print(...args)", "Prints arguments to stdout"},
		{"input", "input(prompt?: string) -> string", "Reads user input from stdin"},
		{"int", "int(value) -> int", "Converts a value to an integer"},
		{"float", "float(value) -> float", "Converts a value to a float"},
		{"str", "str(value) -> string", "Converts a value to a string"},
		{"type", "type(object) -> string", "Returns the type of an object"},
		{"enumerate", "enumerate(iterable) -> array", "Returns an array of [index, value] pairs"},
		{"help", "help() -> string", "Returns help information"},
		{"version", "version() -> string", "Returns version information"},
		{"modules", "modules() -> string", "Lists available modules"},
	}

	completions := []lsp.CompletionItem{}

	for _, builtin := range builtins {
		completions = append(completions, lsp.CompletionItem{
			Label:         builtin.name,
			Kind:          lsp.CompletionItemKindFunction,
			Detail:        builtin.detail,
			Documentation: builtin.doc,
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
		if symbol.GrimoireName != "" {
			detail = detail + " of " + symbol.GrimoireName
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
	// Get all Grimoires and global functions from the symbol table
	globals := a.symbolTable.GetGlobalSymbols()

	completions := []lsp.CompletionItem{}

	for _, symbol := range globals {
		var kind lsp.CompletionItemKind

		switch symbol.Type {
		case "Grimoire":
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
	case "Grimoire":
		content = fmt.Sprintf("**Grimoire** %s\n\n%s", symbol.Name, symbol.Documentation)
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
			symbol.GrimoireName,
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
			symbol.GrimoireName,
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
		"spell": true, "grim": true, "init": true, "self": true,
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
		"grim":        "**Grimoire** - Defines a class in Carrion.",
		"init":        "**init** - Special method for initializing a Grimoire instance.",
		"self":        "**self** - References the current instance within a Grimoire method.",
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
		"arcane":      "**arcane** - Defines an abstract Grimoire (similar to 'abstract class').",
		"arcanespell": "**arcanespell** - Defines an abstract method in a Grimoire.",
		"super":       "**super** - References the parent Grimoire's implementation.",
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

// createDiagnosticFromError creates a diagnostic from a parser error
func (a *CarrionAnalyzer) createDiagnosticFromError(err string) lsp.Diagnostic {
	// Extract line and column information if available
	lineNum, colNum := 1, 1
	errMsg := err

	// Try to extract line/column info using a simple heuristic
	if parts := strings.Split(err, " at line "); len(parts) > 1 {
		if locParts := strings.Split(parts[1], ", column "); len(locParts) > 1 {
			lineNumStr := locParts[0]
			colNumStr := strings.Split(locParts[1], " ")[0]

			// Attempt to parse as integers
			if ln, err := strconv.Atoi(lineNumStr); err == nil {
				lineNum = ln
			}
			if cn, err := strconv.Atoi(colNumStr); err == nil {
				colNum = cn
			}

			// Clean up the error message
			errMsg = parts[0]
		}
	}

	// Create a diagnostic for the error
	return lsp.Diagnostic{
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
}

// performSemanticAnalysis performs additional semantic checks on the AST
func (a *CarrionAnalyzer) performSemanticAnalysis(program *ast.Program, doc *protocol.CarrionDocument) []lsp.Diagnostic {
	diagnostics := []lsp.Diagnostic{}

	// TODO: Implement semantic analysis
	// - Check for undefined variables
	// - Check for unused imports
	// - Check for type mismatches
	// - Check for unreachable code
	// - etc.

	return diagnostics
}

// GetSignatureHelp returns signature help for function calls
func (a *CarrionAnalyzer) GetSignatureHelp(
	uri lsp.DocumentURI,
	position lsp.Position,
) *lsp.SignatureHelp {
	doc := a.documentStore.GetDocument(uri)
	if doc == nil {
		a.logger.Warn("Cannot get signature help for non-existent document: %s", uri)
		return nil
	}

	// Get the current line
	lines := strings.Split(doc.Text, "\n")
	if int(position.Line) >= len(lines) {
		return nil
	}

	line := lines[position.Line]
	if int(position.Character) > len(line) {
		return nil
	}

	// Find the function call context
	funcName, paramIndex := a.getFunctionCallContext(line, position.Character)
	if funcName == "" {
		return nil
	}

	// Look up the function in the symbol table
	symbol := a.symbolTable.LookupSymbol(funcName, string(uri))
	if symbol == nil || (symbol.Type != "spell" && symbol.Type != "method") {
		// Check if it's a built-in function
		return a.getBuiltinSignatureHelp(funcName, paramIndex)
	}

	// Create signature information
	params := make([]string, 0, len(symbol.Parameters))
	paramInfos := make([]lsp.ParameterInformation, 0, len(symbol.Parameters))
	
	for _, param := range symbol.Parameters {
		paramStr := param.Name
		if param.TypeHint != "" {
			paramStr += ": " + param.TypeHint
		}
		if param.DefaultValue != "" {
			paramStr += " = " + param.DefaultValue
		}
		params = append(params, paramStr)
		
		paramInfos = append(paramInfos, lsp.ParameterInformation{
			Label: paramStr,
		})
	}

	signature := symbol.Name + "(" + strings.Join(params, ", ") + ")"
	
	return &lsp.SignatureHelp{
		Signatures: []lsp.SignatureInformation{
			{
				Label:         signature,
				Documentation: symbol.Documentation,
				Parameters:    paramInfos,
			},
		},
		ActiveSignature: 0,
		ActiveParameter: uint32(paramIndex),
	}
}

// getFunctionCallContext analyzes the current position to find function call context
func (a *CarrionAnalyzer) getFunctionCallContext(line string, charPos uint32) (string, int) {
	// Find the opening parenthesis
	parenPos := -1
	parenDepth := 0
	
	for i := int(charPos) - 1; i >= 0; i-- {
		if line[i] == ')' {
			parenDepth++
		} else if line[i] == '(' {
			if parenDepth == 0 {
				parenPos = i
				break
			}
			parenDepth--
		}
	}
	
	if parenPos == -1 {
		return "", 0
	}

	// Extract function name before the parenthesis
	nameEnd := parenPos - 1
	for nameEnd >= 0 && line[nameEnd] == ' ' {
		nameEnd--
	}
	
	nameStart := nameEnd
	for nameStart >= 0 && isIdentifierChar(line[nameStart]) {
		nameStart--
	}
	nameStart++
	
	if nameStart > nameEnd {
		return "", 0
	}
	
	funcName := line[nameStart:nameEnd+1]
	
	// Count commas to determine parameter index
	paramIndex := 0
	for i := parenPos + 1; i < int(charPos); i++ {
		if line[i] == ',' {
			paramIndex++
		}
	}
	
	return funcName, paramIndex
}

// getBuiltinSignatureHelp returns signature help for built-in functions
func (a *CarrionAnalyzer) getBuiltinSignatureHelp(funcName string, paramIndex int) *lsp.SignatureHelp {
	builtinSigs := map[string]struct {
		signature string
		doc       string
		params    []string
	}{
		"print": {
			signature: "print(...args)",
			doc:       "Prints arguments to stdout",
			params:    []string{"...args"},
		},
		"len": {
			signature: "len(object)",
			doc:       "Returns the length of a string, array, tuple, or hash",
			params:    []string{"object"},
		},
		"input": {
			signature: "input(prompt?)",
			doc:       "Reads user input from stdin",
			params:    []string{"prompt?"},
		},
		"int": {
			signature: "int(value)",
			doc:       "Converts a value to an integer",
			params:    []string{"value"},
		},
		"float": {
			signature: "float(value)",
			doc:       "Converts a value to a float",
			params:    []string{"value"},
		},
		"str": {
			signature: "str(value)",
			doc:       "Converts a value to a string",
			params:    []string{"value"},
		},
		"type": {
			signature: "type(object)",
			doc:       "Returns the type of an object",
			params:    []string{"object"},
		},
		"enumerate": {
			signature: "enumerate(iterable)",
			doc:       "Returns an array of [index, value] pairs",
			params:    []string{"iterable"},
		},
	}

	if sig, ok := builtinSigs[funcName]; ok {
		paramInfos := make([]lsp.ParameterInformation, 0, len(sig.params))
		for _, param := range sig.params {
			paramInfos = append(paramInfos, lsp.ParameterInformation{
				Label: param,
			})
		}

		activeParam := uint32(paramIndex)
		if activeParam >= uint32(len(sig.params)) && len(sig.params) > 0 {
			activeParam = uint32(len(sig.params) - 1)
		}

		return &lsp.SignatureHelp{
			Signatures: []lsp.SignatureInformation{
				{
					Label:         sig.signature,
					Documentation: sig.doc,
					Parameters:    paramInfos,
				},
			},
			ActiveSignature: 0,
			ActiveParameter: activeParam,
		}
	}

	return nil
}
