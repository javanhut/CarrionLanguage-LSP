package symbols

import (
	"fmt"
	"strings"

	"github.com/javanhut/TheCarrionLanguage/src/ast"
)

// Parameter represents a function parameter
type Parameter struct {
	Name         string
	TypeHint     string
	DefaultValue string
}

// Symbol represents a symbol in the code
type Symbol struct {
	Name             string
	Type             string // "spellbook", "spell", "variable", "parameter", "field", "method"
	SpellbookName    string // For methods and fields
	ValueType        string // Type hint if available
	Documentation    string
	Parameters       []Parameter // For spells and methods
	DefinitionURI    string
	DefinitionLine   int
	DefinitionColumn int
	Scope            *Scope
}

// SpellbookSymbol represents a spellbook declaration
type SpellbookSymbol struct {
	Name           string
	Methods        []*Symbol
	Fields         []*Symbol
	ParentName     string
	Documentation  string
	DefinitionURI  string
	DefinitionLine int
}

// Scope represents a lexical scope in the code
type Scope struct {
	Parent    *Scope
	Symbols   map[string]*Symbol
	StartLine int
	EndLine   int
	Spellbook *SpellbookSymbol // If this scope is inside a spellbook
	URI       string
}

// SymbolTable maintains symbols for the entire codebase
type SymbolTable struct {
	Global     *Scope
	Spellbooks map[string]*SpellbookSymbol
	FileScopes map[string]*Scope // Top-level scopes for each file
	CurrentURI string
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	globalScope := &Scope{
		Parent:  nil,
		Symbols: make(map[string]*Symbol),
	}

	return &SymbolTable{
		Global:     globalScope,
		Spellbooks: make(map[string]*SpellbookSymbol),
		FileScopes: make(map[string]*Scope),
	}
}

// BuildFromAST builds the symbol table from an AST
func (st *SymbolTable) BuildFromAST(program *ast.Program, uri string) {
	// Create file scope
	fileScope := &Scope{
		Parent:    st.Global,
		Symbols:   make(map[string]*Symbol),
		StartLine: 0,
		URI:       uri,
	}
	st.FileScopes[uri] = fileScope
	st.CurrentURI = uri

	// First pass: collect spellbooks and top-level spells
	for _, stmt := range program.Statements {
		switch node := stmt.(type) {
		case *ast.SpellbookDefinition:
			st.processSpellbookDefinition(node, fileScope)
		case *ast.FunctionDefinition:
			st.processFunctionDefinition(node, fileScope)
		case *ast.AssignStatement:
			st.processAssignStatement(node, fileScope)
		}
	}

	// Second pass: process function bodies and resolve references
	for _, stmt := range program.Statements {
		switch node := stmt.(type) {
		case *ast.FunctionDefinition:
			st.processFunctionBody(node, fileScope)
		}
	}
}

// processSpellbookDefinition processes a spellbook definition
func (st *SymbolTable) processSpellbookDefinition(node *ast.SpellbookDefinition, scope *Scope) {
	spellbookName := node.Name.Value

	// Get position info from token
	line := 0
	column := 0
	if node.Token.Line > 0 {
		line = node.Token.Line - 1 // Convert to 0-based
	}
	if node.Token.Column > 0 {
		column = node.Token.Column - 1 // Convert to 0-based
	}

	spellbook := &SpellbookSymbol{
		Name:           spellbookName,
		Methods:        make([]*Symbol, 0),
		Fields:         make([]*Symbol, 0),
		DefinitionURI:  st.CurrentURI,
		DefinitionLine: line,
	}

	if node.Inherits != nil {
		spellbook.ParentName = node.Inherits.Value
	}

	if node.DocString != nil {
		spellbook.Documentation = node.DocString.Value
	}

	// Create a scope for the spellbook
	spellbookScope := &Scope{
		Parent:    scope,
		Symbols:   make(map[string]*Symbol),
		Spellbook: spellbook,
		URI:       st.CurrentURI,
	}

	// Add the spellbook to the symbol table
	spellbookSymbol := &Symbol{
		Name:             spellbookName,
		Type:             "spellbook",
		Documentation:    spellbook.Documentation,
		DefinitionURI:    st.CurrentURI,
		DefinitionLine:   line,
		DefinitionColumn: column,
		Scope:            spellbookScope,
	}

	scope.Symbols[spellbookName] = spellbookSymbol
	st.Spellbooks[spellbookName] = spellbook

	// Process methods
	for _, method := range node.Methods {
		st.processMethod(method, spellbookScope, spellbook)
	}

	// Process initializer if present
	if node.InitMethod != nil {
		st.processMethod(node.InitMethod, spellbookScope, spellbook)
	}
}

// processMethod processes a method definition within a spellbook
func (st *SymbolTable) processMethod(
	node *ast.FunctionDefinition,
	scope *Scope,
	spellbook *SpellbookSymbol,
) {
	methodName := node.Name.Value

	// Get position info from token
	line := 0
	column := 0
	if node.Token.Line > 0 {
		line = node.Token.Line - 1 // Convert to 0-based
	}
	if node.Token.Column > 0 {
		column = node.Token.Column - 1 // Convert to 0-based
	}

	// Create parameters
	params := make([]Parameter, 0, len(node.Parameters))
	for _, p := range node.Parameters {
		param := Parameter{
			Name: p.Name.Value,
		}

		if p.TypeHint != nil {
			if typeIdent, ok := p.TypeHint.(*ast.Identifier); ok {
				param.TypeHint = typeIdent.Value
			}
		}

		if p.DefaultValue != nil {
			param.DefaultValue = p.DefaultValue.String()
		}

		params = append(params, param)
	}

	// Create symbol for the method
	methodSymbol := &Symbol{
		Name:             methodName,
		Type:             "method",
		SpellbookName:    spellbook.Name,
		Parameters:       params,
		DefinitionURI:    st.CurrentURI,
		DefinitionLine:   line,
		DefinitionColumn: column,
	}

	if node.DocString != nil {
		methodSymbol.Documentation = node.DocString.Value
	}

	// Add to spellbook's methods
	spellbook.Methods = append(spellbook.Methods, methodSymbol)

	// Create method scope
	methodScope := &Scope{
		Parent:    scope,
		Symbols:   make(map[string]*Symbol),
		StartLine: line,
		URI:       st.CurrentURI,
	}

	methodSymbol.Scope = methodScope

	// Add 'self' parameter to method scope
	selfSymbol := &Symbol{
		Name:           "self",
		Type:           "parameter",
		SpellbookName:  spellbook.Name,
		DefinitionURI:  st.CurrentURI,
		DefinitionLine: line,
	}
	methodScope.Symbols["self"] = selfSymbol

	// Add method parameters to scope
	for i, param := range params {
		paramSymbol := &Symbol{
			Name:           param.Name,
			Type:           "parameter",
			ValueType:      param.TypeHint,
			DefinitionURI:  st.CurrentURI,
			DefinitionLine: line,
		}
		methodScope.Symbols[param.Name] = paramSymbol
	}

	// Process method body to find local variables
	if node.Body != nil {
		for _, stmt := range node.Body.Statements {
			st.processStatementForSymbols(stmt, methodScope)
		}
	}
}

// processFunctionDefinition processes a top-level function definition
func (st *SymbolTable) processFunctionDefinition(node *ast.FunctionDefinition, scope *Scope) {
	funcName := node.Name.Value

	// Get position info from token
	line := 0
	column := 0
	if node.Token.Line > 0 {
		line = node.Token.Line - 1 // Convert to 0-based
	}
	if node.Token.Column > 0 {
		column = node.Token.Column - 1 // Convert to 0-based
	}

	// Create parameters
	params := make([]Parameter, 0, len(node.Parameters))
	for _, p := range node.Parameters {
		param := Parameter{
			Name: p.Name.Value,
		}

		if p.TypeHint != nil {
			if typeIdent, ok := p.TypeHint.(*ast.Identifier); ok {
				param.TypeHint = typeIdent.Value
			}
		}

		if p.DefaultValue != nil {
			param.DefaultValue = p.DefaultValue.String()
		}

		params = append(params, param)
	}

	// Create function scope
	funcScope := &Scope{
		Parent:    scope,
		Symbols:   make(map[string]*Symbol),
		StartLine: line,
		URI:       st.CurrentURI,
	}

	// Create symbol for the function
	funcSymbol := &Symbol{
		Name:             funcName,
		Type:             "spell",
		Parameters:       params,
		DefinitionURI:    st.CurrentURI,
		DefinitionLine:   line,
		DefinitionColumn: column,
		Scope:            funcScope,
	}

	if node.DocString != nil {
		funcSymbol.Documentation = node.DocString.Value
	}

	// Add to scope
	scope.Symbols[funcName] = funcSymbol

	// Add function parameters to scope
	for _, param := range params {
		paramSymbol := &Symbol{
			Name:           param.Name,
			Type:           "parameter",
			ValueType:      param.TypeHint,
			DefinitionURI:  st.CurrentURI,
			DefinitionLine: line,
		}
		funcScope.Symbols[param.Name] = paramSymbol
	}
}

// processFunctionBody processes the body of a function to collect local variables
func (st *SymbolTable) processFunctionBody(node *ast.FunctionDefinition, scope *Scope) {
	// Find the function symbol
	funcSymbol, ok := scope.Symbols[node.Name.Value]
	if !ok {
		return
	}

	// Process the function body
	if node.Body != nil {
		for _, stmt := range node.Body.Statements {
			st.processStatementForSymbols(stmt, funcSymbol.Scope)
		}
	}
}

// processAssignStatement processes an assignment statement
func (st *SymbolTable) processAssignStatement(node *ast.AssignStatement, scope *Scope) {
	// Handle different types of assignments
	switch target := node.Name.(type) {
	case *ast.Identifier:
		line := 0
		column := 0
		if node.Token.Line > 0 {
			line = node.Token.Line - 1 // Convert to 0-based
		}
		if node.Token.Column > 0 {
			column = node.Token.Column - 1 // Convert to 0-based
		}

		varName := target.Value

		// Check if this is a spellbook field or method
		if scope.Spellbook != nil {
			// This is a field of a spellbook
			fieldSymbol := &Symbol{
				Name:             varName,
				Type:             "field",
				SpellbookName:    scope.Spellbook.Name,
				DefinitionURI:    st.CurrentURI,
				DefinitionLine:   line,
				DefinitionColumn: column,
			}

			scope.Symbols[varName] = fieldSymbol
			scope.Spellbook.Fields = append(scope.Spellbook.Fields, fieldSymbol)
		} else {
			// This is a variable
			varSymbol := &Symbol{
				Name:             varName,
				Type:             "variable",
				DefinitionURI:    st.CurrentURI,
				DefinitionLine:   line,
				DefinitionColumn: column,
			}

			if node.TypeHint != nil {
				if typeIdent, ok := node.TypeHint.(*ast.Identifier); ok {
					varSymbol.ValueType = typeIdent.Value
				}
			}

			// Try to determine the type from the value
			if node.Value != nil {
				if spellbook, ok := node.Value.(*ast.CallExpression); ok {
					if ident, ok := spellbook.Function.(*ast.Identifier); ok {
						if sb, exists := st.Spellbooks[ident.Value]; exists {
							varSymbol.SpellbookName = sb.Name
							varSymbol.Type = "instance"
						}
					}
				}
			}

			scope.Symbols[varName] = varSymbol
		}
	}
}

// processStatementForSymbols processes statements to collect symbols
func (st *SymbolTable) processStatementForSymbols(stmt ast.Statement, scope *Scope) {
	switch node := stmt.(type) {
	case *ast.AssignStatement:
		st.processAssignStatement(node, scope)

	case *ast.BlockStatement:
		// Create a new scope for blocks
		blockScope := &Scope{
			Parent:  scope,
			Symbols: make(map[string]*Symbol),
			URI:     st.CurrentURI,
		}

		for _, blockStmt := range node.Statements {
			st.processStatementForSymbols(blockStmt, blockScope)
		}

	case *ast.IfStatement:
		// Process both consequence and alternative
		if node.Consequence != nil {
			for _, s := range node.Consequence.Statements {
				st.processStatementForSymbols(s, scope)
			}
		}

		// Process "otherwise" branches
		for _, branch := range node.OtherwiseBranches {
			if branch.Consequence != nil {
				for _, s := range branch.Consequence.Statements {
					st.processStatementForSymbols(s, scope)
				}
			}
		}

		if node.Alternative != nil {
			for _, s := range node.Alternative.Statements {
				st.processStatementForSymbols(s, scope)
			}
		}

	case *ast.ForStatement:
		// Process loop variables
		switch v := node.Variable.(type) {
		case *ast.Identifier:
			line := 0
			if node.Token.Line > 0 {
				line = node.Token.Line - 1
			}

			loopVarSymbol := &Symbol{
				Name:           v.Value,
				Type:           "variable",
				DefinitionURI:  st.CurrentURI,
				DefinitionLine: line,
			}
			scope.Symbols[v.Value] = loopVarSymbol

		case *ast.TupleLiteral:
			line := 0
			if node.Token.Line > 0 {
				line = node.Token.Line - 1
			}

			// Handle tuple unpacking in a for loop (for x, y in items:)
			for _, elem := range v.Elements {
				if ident, ok := elem.(*ast.Identifier); ok {
					loopVarSymbol := &Symbol{
						Name:           ident.Value,
						Type:           "variable",
						DefinitionURI:  st.CurrentURI,
						DefinitionLine: line,
					}
					scope.Symbols[ident.Value] = loopVarSymbol
				}
			}
		}

		// Process loop body
		if node.Body != nil {
			for _, s := range node.Body.Statements {
				st.processStatementForSymbols(s, scope)
			}
		}

		// Process else clause if it exists
		if node.Alternative != nil {
			for _, s := range node.Alternative.Statements {
				st.processStatementForSymbols(s, scope)
			}
		}

	case *ast.WhileStatement:
		// Process while body
		if node.Body != nil {
			for _, s := range node.Body.Statements {
				st.processStatementForSymbols(s, scope)
			}
		}

	case *ast.AttemptStatement:
		// Process attempt block
		if node.TryBlock != nil {
			for _, s := range node.TryBlock.Statements {
				st.processStatementForSymbols(s, scope)
			}
		}

		// Process ensnare clauses
		for _, ensnare := range node.EnsnareClauses {
			if ensnare.Alias != nil {
				// Add the exception variable
				line := 0
				if ensnare.Token.Line > 0 {
					line = ensnare.Token.Line - 1
				}

				exceptionSymbol := &Symbol{
					Name:           ensnare.Alias.Value,
					Type:           "variable",
					DefinitionURI:  st.CurrentURI,
					DefinitionLine: line,
				}
				scope.Symbols[ensnare.Alias.Value] = exceptionSymbol
			}

			if ensnare.Consequence != nil {
				for _, s := range ensnare.Consequence.Statements {
					st.processStatementForSymbols(s, scope)
				}
			}
		}

		// Process resolve block
		if node.ResolveBlock != nil {
			for _, s := range node.ResolveBlock.Statements {
				st.processStatementForSymbols(s, scope)
			}
		}
	}
}

// GetCurrentSpellbook returns the spellbook at the given position, if any
func (st *SymbolTable) GetCurrentSpellbook(uri string, line int) *SpellbookSymbol {
	// Check if we have a file scope for this URI
	fileScope, ok := st.FileScopes[uri]
	if !ok {
		return nil
	}

	// Iterate through symbols to find a spellbook that contains this line
	for _, symbol := range fileScope.Symbols {
		if symbol.Type == "spellbook" && symbol.Scope != nil && symbol.Scope.Spellbook != nil {
			if line >= symbol.DefinitionLine {
				// Assuming the spellbook continues to the end of the file or until another spellbook
				return symbol.Scope.Spellbook
			}
		}
	}

	return nil
}

// GetLocalSymbols returns all symbols in scope at the given position
func (st *SymbolTable) GetLocalSymbols(uri string, line int) []*Symbol {
	// Get the file scope
	fileScope, ok := st.FileScopes[uri]
	if !ok {
		return nil
	}

	// Find the most specific scope for this position
	scope := st.findScopeAtPosition(fileScope, line)
	if scope == nil {
		scope = fileScope
	}

	// Collect all symbols from this scope and its parents
	symbols := make([]*Symbol, 0)
	for s := scope; s != nil; s = s.Parent {
		for _, symbol := range s.Symbols {
			symbols = append(symbols, symbol)
		}
	}

	return symbols
}

// findScopeAtPosition finds the most specific scope at the given position
func (st *SymbolTable) findScopeAtPosition(scope *Scope, line int) *Scope {
	// Check if this scope contains the line
	if scope.StartLine > line || (scope.EndLine > 0 && scope.EndLine < line) {
		return nil
	}

	// Check child scopes first (they are more specific)
	for _, symbol := range scope.Symbols {
		if symbol.Scope != nil {
			if childScope := st.findScopeAtPosition(symbol.Scope, line); childScope != nil {
				return childScope
			}
		}
	}

	// This scope contains the position
	return scope
}

// GetGlobalSymbols returns all global symbols
func (st *SymbolTable) GetGlobalSymbols() []*Symbol {
	symbols := make([]*Symbol, 0)

	// Add all symbols from the global scope
	for _, symbol := range st.Global.Symbols {
		symbols = append(symbols, symbol)
	}

	// Add all spellbooks
	for name, spellbook := range st.Spellbooks {
		// Find the spellbook symbol
		for _, fileScope := range st.FileScopes {
			if symbol, ok := fileScope.Symbols[name]; ok && symbol.Type == "spellbook" {
				symbols = append(symbols, symbol)
				break
			}
		}

		// Add all methods as global functions
		for _, method := range spellbook.Methods {
			symbols = append(symbols, method)
		}
	}

	return symbols
}

// LookupSymbol finds a symbol by name in the appropriate scope
func (st *SymbolTable) LookupSymbol(name string, uri string) *Symbol {
	// Check global scope first
	if symbol, ok := st.Global.Symbols[name]; ok {
		return symbol
	}

	// Check file scope
	if fileScope, ok := st.FileScopes[uri]; ok {
		if symbol, ok := fileScope.Symbols[name]; ok {
			return symbol
		}

		// Check all symbols in this file scope for nested scopes
		for _, s := range fileScope.Symbols {
			if s.Scope != nil {
				if symbol := st.lookupSymbolInScope(name, s.Scope); symbol != nil {
					return symbol
				}
			}
		}
	}

	return nil
}

// lookupSymbolInScope looks for a symbol in a specific scope and its children
func (st *SymbolTable) lookupSymbolInScope(name string, scope *Scope) *Symbol {
	// Check this scope
	if symbol, ok := scope.Symbols[name]; ok {
		return symbol
	}

	// Check child scopes
	for _, s := range scope.Symbols {
		if s.Scope != nil {
			if symbol := st.lookupSymbolInScope(name, s.Scope); symbol != nil {
				return symbol
			}
		}
	}

	return nil
}

// LookupSpellbook finds a spellbook by name
func (st *SymbolTable) LookupSpellbook(name string) *SpellbookSymbol {
	spellbook, ok := st.Spellbooks[name]
	if !ok {
		return nil
	}
	return spellbook
}
