package symbols

import (
	"github.com/javanhut/TheCarrionLanguage/src/ast"
	"github.com/javanhut/TheCarrionLanguage/src/token"
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
	Type             string
	GrimoireName     string
	ValueType        string
	Documentation    string
	Parameters       []Parameter
	DefinitionURI    string
	DefinitionLine   int
	DefinitionColumn int
	Scope            *Scope
}

// GrimoireSymbol represents a Grimoire declaration
type GrimoireSymbol struct {
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
	Grimoire  *GrimoireSymbol
	URI       string
}

// SymbolTable maintains symbols for the entire codebase
type SymbolTable struct {
	Global     *Scope
	Grimoires  map[string]*GrimoireSymbol
	FileScopes map[string]*Scope
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
		Grimoires:  make(map[string]*GrimoireSymbol),
		FileScopes: make(map[string]*Scope),
	}
}

// BuildFromAST builds the symbol table from an AST
func (st *SymbolTable) BuildFromAST(program *ast.Program, uri string) {
	fileScope := &Scope{
		Parent:    st.Global,
		Symbols:   make(map[string]*Symbol),
		StartLine: 0,
		URI:       uri,
	}
	st.FileScopes[uri] = fileScope
	st.CurrentURI = uri

	// First pass: collect Grimoires and top-level spells
	for _, stmt := range program.Statements {
		switch node := stmt.(type) {
		case *ast.GrimoireDefinition:
			st.processGrimoireDefinition(node, fileScope)
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

// processGrimoireDefinition processes a Grimoire definition
func (st *SymbolTable) processGrimoireDefinition(node *ast.GrimoireDefinition, scope *Scope) {
	GrimoireName := node.Name.Value
	line := 0
	column := 0
	tokenPos := extractPositionFromToken(node.Token)
	line = tokenPos.Line
	column = tokenPos.Column

	Grimoire := &GrimoireSymbol{
		Name:           GrimoireName,
		Methods:        make([]*Symbol, 0),
		Fields:         make([]*Symbol, 0),
		DefinitionURI:  st.CurrentURI,
		DefinitionLine: line,
	}

	if node.Inherits != nil {
		Grimoire.ParentName = node.Inherits.Value
	}

	if node.DocString != nil {
		Grimoire.Documentation = node.DocString.Value
	}

	GrimoireScope := &Scope{
		Parent:   scope,
		Symbols:  make(map[string]*Symbol),
		Grimoire: Grimoire,
		URI:      st.CurrentURI,
	}

	GrimoireSymbol := &Symbol{
		Name:             GrimoireName,
		Type:             "Grimoire",
		Documentation:    Grimoire.Documentation,
		DefinitionURI:    st.CurrentURI,
		DefinitionLine:   line,
		DefinitionColumn: column,
		Scope:            GrimoireScope,
	}

	scope.Symbols[GrimoireName] = GrimoireSymbol
	st.Grimoires[GrimoireName] = Grimoire

	for _, method := range node.Methods {
		st.processMethod(method, GrimoireScope, Grimoire)
	}

	if node.InitMethod != nil {
		st.processMethod(node.InitMethod, GrimoireScope, Grimoire)
	}
}

func extractPositionFromToken(tok token.Token) struct{ Line, Column int } {
	// Extract line and column from the token
	// Tokens in Carrion have Line and Column fields
	return struct{ Line, Column int }{
		Line:   tok.Line - 1,   // Convert to 0-based for LSP
		Column: tok.Column - 1, // Convert to 0-based for LSP
	}
}

// processMethod processes a method definition within a Grimoire
func (st *SymbolTable) processMethod(
	node *ast.FunctionDefinition,
	scope *Scope,
	Grimoire *GrimoireSymbol,
) {
	methodName := node.Name.Value
	line := 0
	column := 0

	tokenPos := extractPositionFromToken(node.Token)
	line = tokenPos.Line
	column = tokenPos.Column

	params := make([]Parameter, 0, len(node.Parameters))
	for _, p := range node.Parameters {
		// Type assert the Expression to *Parameter
		if paramNode, ok := p.(*ast.Parameter); ok {
			param := Parameter{
				Name: paramNode.Name.Value,
			}

			if paramNode.TypeHint != nil {
				if typeIdent, ok := paramNode.TypeHint.(*ast.Identifier); ok {
					param.TypeHint = typeIdent.Value
				}
			}

			if paramNode.DefaultValue != nil {
				param.DefaultValue = paramNode.DefaultValue.String()
			}

			params = append(params, param)
		}
	}

	methodSymbol := &Symbol{
		Name:             methodName,
		Type:             "method",
		GrimoireName:     Grimoire.Name,
		Parameters:       params,
		DefinitionURI:    st.CurrentURI,
		DefinitionLine:   line,
		DefinitionColumn: column,
	}

	if node.DocString != nil {
		methodSymbol.Documentation = node.DocString.Value
	}

	Grimoire.Methods = append(Grimoire.Methods, methodSymbol)

	methodScope := &Scope{
		Parent:    scope,
		Symbols:   make(map[string]*Symbol),
		StartLine: line,
		URI:       st.CurrentURI,
	}

	methodSymbol.Scope = methodScope

	selfSymbol := &Symbol{
		Name:           "self",
		Type:           "parameter",
		GrimoireName:   Grimoire.Name,
		DefinitionURI:  st.CurrentURI,
		DefinitionLine: line,
	}
	methodScope.Symbols["self"] = selfSymbol

	for _, param := range params {
		paramSymbol := &Symbol{
			Name:           param.Name,
			Type:           "parameter",
			ValueType:      param.TypeHint,
			DefinitionURI:  st.CurrentURI,
			DefinitionLine: line,
		}
		methodScope.Symbols[param.Name] = paramSymbol
	}

	if node.Body != nil {
		for _, stmt := range node.Body.Statements {
			st.processStatementForSymbols(stmt, methodScope)
		}
	}
}

// processFunctionDefinition processes a top-level function definition
func (st *SymbolTable) processFunctionDefinition(node *ast.FunctionDefinition, scope *Scope) {
	funcName := node.Name.Value
	line := 0
	column := 0

	tokenPos := extractPositionFromToken(node.Token)
	line = tokenPos.Line
	column = tokenPos.Column

	params := make([]Parameter, 0, len(node.Parameters))
	for _, p := range node.Parameters {
		// Type assert the Expression to *Parameter
		if paramNode, ok := p.(*ast.Parameter); ok {
			param := Parameter{
				Name: paramNode.Name.Value,
			}

			if paramNode.TypeHint != nil {
				if typeIdent, ok := paramNode.TypeHint.(*ast.Identifier); ok {
					param.TypeHint = typeIdent.Value
				}
			}

			if paramNode.DefaultValue != nil {
				param.DefaultValue = paramNode.DefaultValue.String()
			}

			params = append(params, param)
		}
	}

	funcScope := &Scope{
		Parent:    scope,
		Symbols:   make(map[string]*Symbol),
		StartLine: line,
		URI:       st.CurrentURI,
	}

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

	scope.Symbols[funcName] = funcSymbol

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
	funcSymbol, ok := scope.Symbols[node.Name.Value]
	if !ok {
		return
	}

	if node.Body != nil {
		for _, stmt := range node.Body.Statements {
			st.processStatementForSymbols(stmt, funcSymbol.Scope)
		}
	}
}

// processAssignStatement processes an assignment statement
func (st *SymbolTable) processAssignStatement(node *ast.AssignStatement, scope *Scope) {
	switch target := node.Name.(type) {
	case *ast.Identifier:
		line := 0
		column := 0

		tokenPos := extractPositionFromToken(node.Token)
		line = tokenPos.Line
		column = tokenPos.Column

		varName := target.Value

		if scope.Grimoire != nil {
			fieldSymbol := &Symbol{
				Name:             varName,
				Type:             "field",
				GrimoireName:     scope.Grimoire.Name,
				DefinitionURI:    st.CurrentURI,
				DefinitionLine:   line,
				DefinitionColumn: column,
			}

			scope.Symbols[varName] = fieldSymbol
			scope.Grimoire.Fields = append(scope.Grimoire.Fields, fieldSymbol)
		} else {
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

			if node.Value != nil {
				if Grimoire, ok := node.Value.(*ast.CallExpression); ok {
					if ident, ok := Grimoire.Function.(*ast.Identifier); ok {
						if sb, exists := st.Grimoires[ident.Value]; exists {
							varSymbol.GrimoireName = sb.Name
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
		blockScope := &Scope{
			Parent:  scope,
			Symbols: make(map[string]*Symbol),
			URI:     st.CurrentURI,
		}

		for _, blockStmt := range node.Statements {
			st.processStatementForSymbols(blockStmt, blockScope)
		}

	case *ast.IfStatement:
		if node.Consequence != nil {
			for _, s := range node.Consequence.Statements {
				st.processStatementForSymbols(s, scope)
			}
		}

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
		switch v := node.Variable.(type) {
		case *ast.Identifier:
			line := 0
			tokenPos := extractPositionFromToken(node.Token)
			line = tokenPos.Line

			loopVarSymbol := &Symbol{
				Name:           v.Value,
				Type:           "variable",
				DefinitionURI:  st.CurrentURI,
				DefinitionLine: line,
			}
			scope.Symbols[v.Value] = loopVarSymbol

		case *ast.TupleLiteral:
			line := 0
			tokenPos := extractPositionFromToken(node.Token)
			line = tokenPos.Line

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

		if node.Body != nil {
			for _, s := range node.Body.Statements {
				st.processStatementForSymbols(s, scope)
			}
		}

		if node.Alternative != nil {
			for _, s := range node.Alternative.Statements {
				st.processStatementForSymbols(s, scope)
			}
		}

	case *ast.WhileStatement:
		if node.Body != nil {
			for _, s := range node.Body.Statements {
				st.processStatementForSymbols(s, scope)
			}
		}

	case *ast.AttemptStatement:
		if node.TryBlock != nil {
			for _, s := range node.TryBlock.Statements {
				st.processStatementForSymbols(s, scope)
			}
		}

		for _, ensnare := range node.EnsnareClauses {
			if ensnare.Alias != nil {
				line := 0
				tokenPos := extractPositionFromToken(ensnare.Token)
				line = tokenPos.Line

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

		if node.ResolveBlock != nil {
			for _, s := range node.ResolveBlock.Statements {
				st.processStatementForSymbols(s, scope)
			}
		}
	}
}

// GetCurrentGrimoire returns the Grimoire at the given position, if any
func (st *SymbolTable) GetCurrentGrimoire(uri string, line int) *GrimoireSymbol {
	fileScope, ok := st.FileScopes[uri]
	if !ok {
		return nil
	}

	for _, symbol := range fileScope.Symbols {
		if symbol.Type == "Grimoire" && symbol.Scope != nil && symbol.Scope.Grimoire != nil {
			if line >= symbol.DefinitionLine {
				return symbol.Scope.Grimoire
			}
		}
	}

	return nil
}

// GetLocalSymbols returns all symbols in scope at the given position
func (st *SymbolTable) GetLocalSymbols(uri string, line int) []*Symbol {
	fileScope, ok := st.FileScopes[uri]
	if !ok {
		return nil
	}

	scope := st.findScopeAtPosition(fileScope, line)
	if scope == nil {
		scope = fileScope
	}

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
	if scope.StartLine > line || (scope.EndLine > 0 && scope.EndLine < line) {
		return nil
	}

	for _, symbol := range scope.Symbols {
		if symbol.Scope != nil {
			if childScope := st.findScopeAtPosition(symbol.Scope, line); childScope != nil {
				return childScope
			}
		}
	}

	return scope
}

// GetGlobalSymbols returns all global symbols
func (st *SymbolTable) GetGlobalSymbols() []*Symbol {
	symbols := make([]*Symbol, 0)

	for _, symbol := range st.Global.Symbols {
		symbols = append(symbols, symbol)
	}

	for name, Grimoire := range st.Grimoires {
		for _, fileScope := range st.FileScopes {
			if symbol, ok := fileScope.Symbols[name]; ok && symbol.Type == "Grimoire" {
				symbols = append(symbols, symbol)
				break
			}
		}

		for _, method := range Grimoire.Methods {
			symbols = append(symbols, method)
		}
	}

	return symbols
}

// LookupSymbol finds a symbol by name in the appropriate scope
func (st *SymbolTable) LookupSymbol(name string, uri string) *Symbol {
	if symbol, ok := st.Global.Symbols[name]; ok {
		return symbol
	}

	if fileScope, ok := st.FileScopes[uri]; ok {
		if symbol, ok := fileScope.Symbols[name]; ok {
			return symbol
		}

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
	if symbol, ok := scope.Symbols[name]; ok {
		return symbol
	}

	for _, s := range scope.Symbols {
		if s.Scope != nil {
			if symbol := st.lookupSymbolInScope(name, s.Scope); symbol != nil {
				return symbol
			}
		}
	}

	return nil
}

// LookupGrimoire finds a Grimoire by name
func (st *SymbolTable) LookupGrimoire(name string) *GrimoireSymbol {
	Grimoire, ok := st.Grimoires[name]
	if !ok {
		return nil
	}
	return Grimoire
}
