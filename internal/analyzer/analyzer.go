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
			// Check if it's a string literal or variable that might be a string
			stringCompletions := a.getStringCompletions(uri, objectName, textBeforeCursor)
			if len(stringCompletions) > 0 {
				completionItems = append(completionItems, stringCompletions...)
			} else {
				// Check if it's a grimoire instance or grimoire class
				grimoireCompletions := a.getGrimoireCompletions(uri, objectName)
				if len(grimoireCompletions) > 0 {
					completionItems = append(completionItems, grimoireCompletions...)
				} else {
					// Add completions for other objects based on symbol table
					objectCompletions := a.getObjectCompletions(uri, objectName)
					completionItems = append(completionItems, objectCompletions...)
				}
			}
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
		
		// Add built-in grimoires
		builtinGrimoires := a.getBuiltinGrimoireNames()
		completionItems = append(completionItems, builtinGrimoires...)
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

// getStringCompletions returns completions for string methods
func (a *CarrionAnalyzer) getStringCompletions(
	uri lsp.DocumentURI,
	objectName string,
	textBeforeCursor string,
) []lsp.CompletionItem {
	// Check if the object is likely a string (either a string literal or string variable)
	isStringLikely := false
	
	// Check for string literals (quoted strings)
	if strings.Contains(textBeforeCursor, "\"") || strings.Contains(textBeforeCursor, "'") {
		isStringLikely = true
	}
	
	// Check if variable is known to be a string from symbol table
	if !isStringLikely {
		symbol := a.symbolTable.LookupSymbol(objectName, string(uri))
		if symbol != nil && (symbol.ValueType == "string" || symbol.ValueType == "STRING") {
			isStringLikely = true
		}
	}
	
	// If we have evidence this might be a string, provide string method completions
	if !isStringLikely {
		return nil
	}
	
	// String grimoire methods from the Munin standard library
	stringMethods := []struct {
		name   string
		detail string
		doc    string
	}{
		{"length", "length() -> int", "Returns the length of the string"},
		{"upper", "upper() -> string", "Converts string to uppercase"},
		{"lower", "lower() -> string", "Converts string to lowercase"},
		{"reverse", "reverse() -> string", "Reverses the order of characters in the string"},
		{"find", "find(substring) -> int", "Finds the position of a substring, returns -1 if not found"},
		{"contains", "contains(substring) -> bool", "Checks if the string contains a substring"},
		{"char_at", "char_at(index) -> string", "Returns character at index with bounds checking"},
	}
	
	completions := []lsp.CompletionItem{}
	
	for _, method := range stringMethods {
		completions = append(completions, lsp.CompletionItem{
			Label:         method.name,
			Kind:          lsp.CompletionItemKindMethod,
			Detail:        method.detail,
			Documentation: method.doc,
		})
	}
	
	return completions
}

// getGrimoireCompletions returns completions for grimoire methods and fields
func (a *CarrionAnalyzer) getGrimoireCompletions(
	uri lsp.DocumentURI,
	objectName string,
) []lsp.CompletionItem {
	// First check built-in grimoires (like Time)
	if builtinCompletions := a.getBuiltinGrimoireCompletions(objectName); len(builtinCompletions) > 0 {
		return builtinCompletions
	}
	
	// Check if the object is a grimoire instance from the symbol table
	symbol := a.symbolTable.LookupSymbol(objectName, string(uri))
	if symbol != nil {
		// Case 1: Variable that's an instance of a grimoire
		if symbol.Type == "instance" && symbol.GrimoireName != "" {
			grimoire := a.symbolTable.LookupGrimoire(symbol.GrimoireName)
			if grimoire != nil {
				return a.getMethodsAndFieldsForGrimoire(grimoire)
			}
		}
		
		// Case 2: Direct grimoire class (for static methods)
		if symbol.Type == "Grimoire" {
			grimoire := a.symbolTable.LookupGrimoire(symbol.Name)
			if grimoire != nil {
				return a.getMethodsAndFieldsForGrimoire(grimoire)
			}
		}
	}
	
	// Case 3: Check if it's a known grimoire name directly
	grimoire := a.symbolTable.LookupGrimoire(objectName)
	if grimoire != nil {
		return a.getMethodsAndFieldsForGrimoire(grimoire)
	}
	
	return nil
}

// getBuiltinGrimoireCompletions returns completions for built-in grimoires
func (a *CarrionAnalyzer) getBuiltinGrimoireCompletions(objectName string) []lsp.CompletionItem {
	switch objectName {
	case "Time":
		return a.getTimeGrimoireCompletions()
	case "String":
		return a.getStringGrimoireCompletions()
	case "Array":
		return a.getArrayGrimoireCompletions()
	case "Math":
		return a.getMathGrimoireCompletions()
	case "File":
		return a.getFileGrimoireCompletions()
	case "OS":
		return a.getOSGrimoireCompletions()
	default:
		return nil
	}
}

// getTimeGrimoireCompletions returns method completions for the Time grimoire
func (a *CarrionAnalyzer) getTimeGrimoireCompletions() []lsp.CompletionItem {
	methods := []struct {
		name   string
		detail string
		doc    string
	}{
		{"now", "now() -> int", "Get current Unix timestamp (seconds since epoch)"},
		{"now_nano", "now_nano() -> int", "Get current Unix timestamp in nanoseconds"},
		{"sleep", "sleep(seconds)", "Sleep for specified number of seconds (can be float)"},
		{"format", "format(timestamp, format_str?) -> string", "Format Unix timestamp to string using Go time format"},
		{"parse", "parse(format_str, time_str) -> int", "Parse time string using format, returns Unix timestamp"},
		{"date", "date(timestamp?) -> array", "Get date components [year, month, day] from timestamp or current time"},
		{"add_duration", "add_duration(timestamp, seconds) -> int", "Add duration in seconds to timestamp, returns new timestamp"},
		{"diff", "diff(timestamp1, timestamp2) -> int", "Calculate difference between two timestamps in seconds"},
	}
	
	completions := []lsp.CompletionItem{}
	for _, method := range methods {
		completions = append(completions, lsp.CompletionItem{
			Label:         method.name,
			Kind:          lsp.CompletionItemKindMethod,
			Detail:        method.detail,
			Documentation: method.doc,
		})
	}
	
	return completions
}

// getStringGrimoireCompletions returns method completions for the String grimoire
func (a *CarrionAnalyzer) getStringGrimoireCompletions() []lsp.CompletionItem {
	methods := []struct {
		name   string
		detail string
		doc    string
	}{
		{"length", "length() -> int", "Returns the length of the string"},
		{"upper", "upper() -> string", "Converts string to uppercase"},
		{"lower", "lower() -> string", "Converts string to lowercase"},
		{"reverse", "reverse() -> string", "Reverses the order of characters in the string"},
		{"find", "find(substring) -> int", "Finds the position of a substring, returns -1 if not found"},
		{"contains", "contains(substring) -> bool", "Checks if the string contains a substring"},
		{"char_at", "char_at(index) -> string", "Returns character at index with bounds checking"},
	}
	
	completions := []lsp.CompletionItem{}
	for _, method := range methods {
		completions = append(completions, lsp.CompletionItem{
			Label:         method.name,
			Kind:          lsp.CompletionItemKindMethod,
			Detail:        method.detail,
			Documentation: method.doc,
		})
	}
	
	return completions
}

// getArrayGrimoireCompletions returns method completions for the Array grimoire
func (a *CarrionAnalyzer) getArrayGrimoireCompletions() []lsp.CompletionItem {
	methods := []struct {
		name   string
		detail string
		doc    string
	}{
		{"length", "length() -> int", "Returns the length of the array"},
		{"append", "append(item)", "Appends an item to the end of the array"},
		{"prepend", "prepend(item)", "Prepends an item to the beginning of the array"},
		{"pop", "pop() -> any", "Removes and returns the last item"},
		{"shift", "shift() -> any", "Removes and returns the first item"},
		{"contains", "contains(item) -> bool", "Checks if the array contains an item"},
		{"index", "index(item) -> int", "Returns the index of the first occurrence of item"},
		{"reverse", "reverse()", "Reverses the array in place"},
		{"sort", "sort()", "Sorts the array in place"},
	}
	
	completions := []lsp.CompletionItem{}
	for _, method := range methods {
		completions = append(completions, lsp.CompletionItem{
			Label:         method.name,
			Kind:          lsp.CompletionItemKindMethod,
			Detail:        method.detail,
			Documentation: method.doc,
		})
	}
	
	return completions
}

// getMathGrimoireCompletions returns method completions for the Math grimoire
func (a *CarrionAnalyzer) getMathGrimoireCompletions() []lsp.CompletionItem {
	methods := []struct {
		name   string
		detail string
		doc    string
	}{
		{"abs", "abs(number) -> number", "Returns absolute value"},
		{"sqrt", "sqrt(number) -> float", "Returns square root"},
		{"pow", "pow(base, exponent) -> number", "Returns base raised to exponent"},
		{"sin", "sin(radians) -> float", "Returns sine of angle in radians"},
		{"cos", "cos(radians) -> float", "Returns cosine of angle in radians"},
		{"tan", "tan(radians) -> float", "Returns tangent of angle in radians"},
		{"log", "log(number) -> float", "Returns natural logarithm"},
		{"ceil", "ceil(number) -> int", "Returns ceiling (round up)"},
		{"floor", "floor(number) -> int", "Returns floor (round down)"},
		{"round", "round(number) -> int", "Returns rounded value"},
		{"min", "min(...numbers) -> number", "Returns minimum value"},
		{"max", "max(...numbers) -> number", "Returns maximum value"},
	}
	
	completions := []lsp.CompletionItem{}
	for _, method := range methods {
		completions = append(completions, lsp.CompletionItem{
			Label:         method.name,
			Kind:          lsp.CompletionItemKindMethod,
			Detail:        method.detail,
			Documentation: method.doc,
		})
	}
	
	return completions
}

// getFileGrimoireCompletions returns method completions for the File grimoire
func (a *CarrionAnalyzer) getFileGrimoireCompletions() []lsp.CompletionItem {
	methods := []struct {
		name   string
		detail string
		doc    string
	}{
		{"read", "read(path) -> string", "Reads content from a file"},
		{"write", "write(path, content)", "Writes content to a file"},
		{"append", "append(path, content)", "Appends content to a file"},
		{"exists", "exists(path) -> bool", "Checks if a file exists"},
		{"delete", "delete(path)", "Deletes a file"},
		{"copy", "copy(source, destination)", "Copies a file"},
		{"size", "size(path) -> int", "Returns file size in bytes"},
	}
	
	completions := []lsp.CompletionItem{}
	for _, method := range methods {
		completions = append(completions, lsp.CompletionItem{
			Label:         method.name,
			Kind:          lsp.CompletionItemKindMethod,
			Detail:        method.detail,
			Documentation: method.doc,
		})
	}
	
	return completions
}

// getOSGrimoireCompletions returns method completions for the OS grimoire
func (a *CarrionAnalyzer) getOSGrimoireCompletions() []lsp.CompletionItem {
	methods := []struct {
		name   string
		detail string
		doc    string
	}{
		{"run", "run(command) -> string", "Executes a system command"},
		{"getenv", "getenv(name) -> string", "Gets environment variable"},
		{"setenv", "setenv(name, value)", "Sets environment variable"},
		{"getcwd", "getcwd() -> string", "Gets current working directory"},
		{"chdir", "chdir(path)", "Changes current directory"},
		{"listdir", "listdir(path) -> array", "Lists directory contents"},
		{"mkdir", "mkdir(path)", "Creates a directory"},
		{"rmdir", "rmdir(path)", "Removes a directory"},
		{"expandenv", "expandenv(path) -> string", "Expands environment variables in path"},
	}
	
	completions := []lsp.CompletionItem{}
	for _, method := range methods {
		completions = append(completions, lsp.CompletionItem{
			Label:         method.name,
			Kind:          lsp.CompletionItemKindMethod,
			Detail:        method.detail,
			Documentation: method.doc,
		})
	}
	
	return completions
}

// getMethodsAndFieldsForGrimoire returns method and field completions for a custom grimoire
func (a *CarrionAnalyzer) getMethodsAndFieldsForGrimoire(grimoire *symbols.GrimoireSymbol) []lsp.CompletionItem {
	completions := []lsp.CompletionItem{}
	
	// Add methods
	for _, method := range grimoire.Methods {
		var paramStrs []string
		for _, param := range method.Parameters {
			paramStr := param.Name
			if param.TypeHint != "" {
				paramStr += ": " + param.TypeHint
			}
			if param.DefaultValue != "" {
				paramStr += " = " + param.DefaultValue
			}
			paramStrs = append(paramStrs, paramStr)
		}
		
		signature := method.Name + "(" + strings.Join(paramStrs, ", ") + ")"
		
		completions = append(completions, lsp.CompletionItem{
			Label:         method.Name,
			Kind:          lsp.CompletionItemKindMethod,
			Detail:        signature,
			Documentation: method.Documentation,
		})
	}
	
	// Add fields
	for _, field := range grimoire.Fields {
		completions = append(completions, lsp.CompletionItem{
			Label:         field.Name,
			Kind:          lsp.CompletionItemKindField,
			Detail:        "field of " + grimoire.Name,
			Documentation: field.Documentation,
		})
	}
	
	return completions
}

// getBuiltinGrimoireNames returns built-in grimoire names for completion
func (a *CarrionAnalyzer) getBuiltinGrimoireNames() []lsp.CompletionItem {
	grimoires := []struct {
		name   string
		detail string
		doc    string
	}{
		{"Time", "Time grimoire", "Time-related functionality including timestamps, formatting, and date operations"},
		{"String", "String grimoire", "String manipulation methods including case conversion, searching, and character access"},
		{"Array", "Array grimoire", "Array operations including length, append, pop, sort, and search"},
		{"Math", "Math grimoire", "Mathematical functions including trigonometry, logarithms, and rounding"},
		{"File", "File grimoire", "File I/O operations including read, write, copy, and existence checks"},
		{"OS", "OS grimoire", "Operating system interface including environment variables and command execution"},
		{"Boolean", "Boolean grimoire", "Boolean value operations and logical functions"},
		{"Integer", "Integer grimoire", "Integer manipulation and conversion functions"},
		{"Float", "Float grimoire", "Floating-point number operations and formatting"},
		{"Debug", "Debug grimoire", "Debugging utilities and diagnostic functions"},
	}
	
	completions := []lsp.CompletionItem{}
	for _, grimoire := range grimoires {
		completions = append(completions, lsp.CompletionItem{
			Label:         grimoire.name,
			Kind:          lsp.CompletionItemKindClass,
			Detail:        grimoire.detail,
			Documentation: grimoire.doc,
		})
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
		{"bool", "bool(value) -> bool", "Converts a value to a boolean"},
		{"list", "list(iterable) -> array", "Converts an iterable to an array"},
		{"tuple", "tuple(iterable) -> tuple", "Converts an iterable to a tuple"},
		{"type", "type(object) -> string", "Returns the type of an object"},
		{"enumerate", "enumerate(iterable) -> array", "Returns an array of [index, value] pairs"},
		{"pairs", "pairs(hash) -> array", "Returns key-value pairs from a hash"},
		{"range", "range(start, stop?, step?) -> array", "Generates a range of numbers"},
		{"max", "max(...args) -> number", "Returns the maximum value"},
		{"abs", "abs(number) -> number", "Returns absolute value of a number"},
		{"ord", "ord(char) -> int", "Returns ASCII code of a single character"},
		{"chr", "chr(code) -> string", "Converts ASCII code (0-255) to character"},
		{"is_sametype", "is_sametype(obj1, obj2) -> bool", "Checks if two objects have the same type"},
		{"Error", "Error(name, message) -> Error", "Creates a new Error object"},
		{"help", "help() -> string", "Returns help information"},
		{"version", "version() -> string", "Returns version information"},
		{"modules", "modules() -> string", "Lists available modules"},
		{"fileRead", "fileRead(path) -> string", "Reads content from a file"},
		{"fileWrite", "fileWrite(path, content)", "Writes content to a file"},
		{"fileAppend", "fileAppend(path, content)", "Appends content to a file"},
		{"fileExists", "fileExists(path) -> bool", "Checks if a file exists"},
		{"osRunCommand", "osRunCommand(command) -> string", "Executes a system command"},
		{"osGetEnv", "osGetEnv(name) -> string", "Gets environment variable"},
		{"osSetEnv", "osSetEnv(name, value)", "Sets environment variable"},
		{"osGetCwd", "osGetCwd() -> string", "Gets current working directory"},
		{"osChdir", "osChdir(path)", "Changes current directory"},
		{"osSleep", "osSleep(seconds)", "Sleeps for specified seconds"},
		{"osListDir", "osListDir(path) -> array", "Lists directory contents"},
		{"osRemove", "osRemove(path)", "Removes a file or directory"},
		{"osMkdir", "osMkdir(path)", "Creates a directory"},
		{"osExpandEnv", "osExpandEnv(path) -> string", "Expands environment variables in path"},
		{"timeNow", "timeNow() -> int", "Returns current Unix timestamp"},
		{"timeNowNano", "timeNowNano() -> int", "Returns current timestamp in nanoseconds"},
		{"timeSleep", "timeSleep(seconds)", "Sleep for specified seconds"},
		{"timeFormat", "timeFormat(timestamp, format) -> string", "Formats timestamp using Go time format"},
		{"timeParse", "timeParse(format, timeString) -> int", "Parses time string to timestamp"},
		{"timeDate", "timeDate(timestamp) -> array", "Returns [year, month, day, hour, minute, second]"},
		{"timeAddDuration", "timeAddDuration(timestamp, seconds) -> int", "Adds duration to timestamp"},
		{"timeDiff", "timeDiff(timestamp1, timestamp2) -> int", "Calculate difference between timestamps"},
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

	// Check for string indexing bounds and syntax
	stringIndexDiagnostics := a.checkStringIndexing(program, doc)
	diagnostics = append(diagnostics, stringIndexDiagnostics...)

	// TODO: Implement additional semantic analysis
	// - Check for undefined variables
	// - Check for unused imports
	// - Check for type mismatches
	// - Check for unreachable code
	// - etc.

	return diagnostics
}

// checkStringIndexing validates string indexing operations using simple text analysis
func (a *CarrionAnalyzer) checkStringIndexing(program *ast.Program, doc *protocol.CarrionDocument) []lsp.Diagnostic {
	diagnostics := []lsp.Diagnostic{}
	
	// For now, we'll do simple text-based analysis for string indexing
	// This is a simplified approach until we can implement proper AST walking
	lines := strings.Split(doc.Text, "\n")
	for lineNum, line := range lines {
		// Look for string indexing patterns like s[0], "hello"[1], etc.
		if strings.Contains(line, "[") && strings.Contains(line, "]") {
			// Simple regex-based detection for string literals with indexing
			if strings.Contains(line, "\"") && strings.Contains(line, "[") {
				// This is a very basic check - in a real implementation, 
				// we'd need proper AST traversal
				a.logger.Debug("Found potential string indexing on line %d: %s", lineNum+1, line)
			}
		}
	}
	
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
		"bool": {
			signature: "bool(value)",
			doc:       "Converts a value to a boolean",
			params:    []string{"value"},
		},
		"list": {
			signature: "list(iterable)",
			doc:       "Converts an iterable to an array",
			params:    []string{"iterable"},
		},
		"tuple": {
			signature: "tuple(iterable)",
			doc:       "Converts an iterable to a tuple",
			params:    []string{"iterable"},
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
		"pairs": {
			signature: "pairs(hash)",
			doc:       "Returns key-value pairs from a hash",
			params:    []string{"hash"},
		},
		"range": {
			signature: "range(start, stop?, step?)",
			doc:       "Generates a range of numbers",
			params:    []string{"start", "stop?", "step?"},
		},
		"max": {
			signature: "max(...args)",
			doc:       "Returns the maximum value",
			params:    []string{"...args"},
		},
		"abs": {
			signature: "abs(number)",
			doc:       "Returns absolute value of a number",
			params:    []string{"number"},
		},
		"ord": {
			signature: "ord(char)",
			doc:       "Returns ASCII code of a single character",
			params:    []string{"char"},
		},
		"chr": {
			signature: "chr(code)",
			doc:       "Converts ASCII code (0-255) to character",
			params:    []string{"code"},
		},
		"is_sametype": {
			signature: "is_sametype(obj1, obj2)",
			doc:       "Checks if two objects have the same type",
			params:    []string{"obj1", "obj2"},
		},
		"Error": {
			signature: "Error(name, message)",
			doc:       "Creates a new Error object",
			params:    []string{"name", "message"},
		},
		"fileRead": {
			signature: "fileRead(path)",
			doc:       "Reads content from a file",
			params:    []string{"path"},
		},
		"fileWrite": {
			signature: "fileWrite(path, content)",
			doc:       "Writes content to a file",
			params:    []string{"path", "content"},
		},
		"fileAppend": {
			signature: "fileAppend(path, content)",
			doc:       "Appends content to a file",
			params:    []string{"path", "content"},
		},
		"fileExists": {
			signature: "fileExists(path)",
			doc:       "Checks if a file exists",
			params:    []string{"path"},
		},
		"osRunCommand": {
			signature: "osRunCommand(command)",
			doc:       "Executes a system command",
			params:    []string{"command"},
		},
		"osGetEnv": {
			signature: "osGetEnv(name)",
			doc:       "Gets environment variable",
			params:    []string{"name"},
		},
		"osSetEnv": {
			signature: "osSetEnv(name, value)",
			doc:       "Sets environment variable",
			params:    []string{"name", "value"},
		},
		"osGetCwd": {
			signature: "osGetCwd()",
			doc:       "Gets current working directory",
			params:    []string{},
		},
		"osChdir": {
			signature: "osChdir(path)",
			doc:       "Changes current directory",
			params:    []string{"path"},
		},
		"osSleep": {
			signature: "osSleep(seconds)",
			doc:       "Sleeps for specified seconds",
			params:    []string{"seconds"},
		},
		"osListDir": {
			signature: "osListDir(path)",
			doc:       "Lists directory contents",
			params:    []string{"path"},
		},
		"osRemove": {
			signature: "osRemove(path)",
			doc:       "Removes a file or directory",
			params:    []string{"path"},
		},
		"osMkdir": {
			signature: "osMkdir(path)",
			doc:       "Creates a directory",
			params:    []string{"path"},
		},
		"osExpandEnv": {
			signature: "osExpandEnv(path)",
			doc:       "Expands environment variables in path",
			params:    []string{"path"},
		},
		"timeNow": {
			signature: "timeNow()",
			doc:       "Returns current Unix timestamp",
			params:    []string{},
		},
		"timeNowNano": {
			signature: "timeNowNano()",
			doc:       "Returns current timestamp in nanoseconds",
			params:    []string{},
		},
		"timeSleep": {
			signature: "timeSleep(seconds)",
			doc:       "Sleep for specified seconds",
			params:    []string{"seconds"},
		},
		"timeFormat": {
			signature: "timeFormat(timestamp, format)",
			doc:       "Formats timestamp using Go time format",
			params:    []string{"timestamp", "format"},
		},
		"timeParse": {
			signature: "timeParse(format, timeString)",
			doc:       "Parses time string to timestamp",
			params:    []string{"format", "timeString"},
		},
		"timeDate": {
			signature: "timeDate(timestamp)",
			doc:       "Returns [year, month, day, hour, minute, second]",
			params:    []string{"timestamp"},
		},
		"timeAddDuration": {
			signature: "timeAddDuration(timestamp, seconds)",
			doc:       "Adds duration to timestamp",
			params:    []string{"timestamp", "seconds"},
		},
		"timeDiff": {
			signature: "timeDiff(timestamp1, timestamp2)",
			doc:       "Calculate difference between timestamps",
			params:    []string{"timestamp1", "timestamp2"},
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
