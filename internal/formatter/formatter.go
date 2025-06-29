package formatter

import (
	"strings"

	"github.com/javanhut/TheCarrionLanguage/src/lexer"
	"github.com/javanhut/TheCarrionLanguage/src/parser"
	lsp "go.lsp.dev/protocol"

	"github.com/carrionlang-lsp/lsp/internal/protocol"
	"github.com/carrionlang-lsp/lsp/internal/util"
)

// CarrionFormatter provides formatting services for Carrion files
type CarrionFormatter struct {
	logger *util.Logger
}

// NewCarrionFormatter creates a new formatter
func NewCarrionFormatter(logger *util.Logger) *CarrionFormatter {
	return &CarrionFormatter{
		logger: logger,
	}
}

// Format formats a document and returns text edits
func (f *CarrionFormatter) Format(doc *protocol.CarrionDocument) []lsp.TextEdit {
	if doc == nil {
		f.logger.Warn("Cannot format nil document")
		return nil
	}

	f.logger.Debug("Formatting document: %s", doc.URI)

	// Parse the document to check for syntax errors
	l := lexer.New(doc.Text)
	p := parser.New(l)
	_ = p.ParseProgram() // Just check for errors, not using the AST

	// Check for parser errors
	if len(p.Errors()) > 0 {
		f.logger.Warn("Cannot format document with parser errors: %v", p.Errors())
		return nil
	}

	// Format the document
	formatted := f.formatDocument(doc.Text)
	if formatted == doc.Text {
		f.logger.Debug("Document is already properly formatted")
		return nil
	}

	// Create a single text edit for the entire document
	return []lsp.TextEdit{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   calculateEndPosition(doc.Text),
			},
			NewText: formatted,
		},
	}
}

// calculateEndPosition calculates the position of the end of the document
func calculateEndPosition(text string) lsp.Position {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return lsp.Position{Line: 0, Character: 0}
	}

	lastLineNum := len(lines) - 1
	lastLineLen := len(lines[lastLineNum])

	return lsp.Position{
		Line:      uint32(lastLineNum),
		Character: uint32(lastLineLen),
	}
}

// formatterContext keeps track of formatting state
type formatterContext struct {
	indentLevel      int
	indentStr        string
	output           *strings.Builder
	inGrimoire       bool
	inSpell          bool
	lastLineEmpty    bool
	lastLineIndented bool
	blockStack       []string // Tracks nested block types for context-aware formatting
}

// formatDocument formats a Carrion document
func (f *CarrionFormatter) formatDocument(text string) string {
	// First, handle multi-line comments
	text = f.normalizeBlockComments(text)
	
	// Create a formatter context
	ctx := &formatterContext{
		indentLevel:   0,
		indentStr:     "    ", // 4 spaces for indentation
		output:        &strings.Builder{},
		inGrimoire:    false,
		inSpell:       false,
		lastLineEmpty: false,
		blockStack:    make([]string, 0),
	}

	// Split the document into lines
	lines := strings.Split(text, "\n")

	// Process each line with correct indentation
	for i, line := range lines {
		// Get next line if available, otherwise empty string
		nextLine := ""
		if i < len(lines)-1 {
			nextLine = lines[i+1]
		}
		f.formatLine(line, i, ctx, nextLine)
	}

	return ctx.output.String()
}

// normalizeBlockComments handles multi-line comment formatting
func (f *CarrionFormatter) normalizeBlockComments(text string) string {
	// Handle /* */ block comments
	text = f.normalizeSlashStarComments(text)
	
	// Handle ``` triple backtick comments
	text = f.normalizeTripleBacktickComments(text)
	
	return text
}

// normalizeSlashStarComments formats /* */ style block comments
func (f *CarrionFormatter) normalizeSlashStarComments(text string) string {
	var result strings.Builder
	inComment := false
	i := 0
	
	for i < len(text) {
		if !inComment && i < len(text)-1 && text[i] == '/' && text[i+1] == '*' {
			inComment = true
			result.WriteString("/*")
			i += 2
		} else if inComment && i < len(text)-1 && text[i] == '*' && text[i+1] == '/' {
			inComment = false
			result.WriteString("*/")
			i += 2
		} else {
			result.WriteByte(text[i])
			i++
		}
	}
	
	return result.String()
}

// normalizeTripleBacktickComments formats ``` style block comments
func (f *CarrionFormatter) normalizeTripleBacktickComments(text string) string {
	var result strings.Builder
	inComment := false
	i := 0
	
	for i < len(text) {
		if !inComment && i < len(text)-2 && text[i] == '`' && text[i+1] == '`' && text[i+2] == '`' {
			inComment = true
			result.WriteString("```")
			i += 3
		} else if inComment && i < len(text)-2 && text[i] == '`' && text[i+1] == '`' && text[i+2] == '`' {
			inComment = false
			result.WriteString("```")
			i += 3
		} else {
			result.WriteByte(text[i])
			i++
		}
	}
	
	return result.String()
}

// formatLine formats a single line of code
func (f *CarrionFormatter) formatLine(
	line string,
	lineNum int,
	ctx *formatterContext,
	nextLine string,
) {
	// Skip empty lines but keep track of them
	trimmedLine := strings.TrimSpace(line)
	if trimmedLine == "" {
		// Don't output consecutive empty lines
		if !ctx.lastLineEmpty {
			ctx.output.WriteString("\n")
			ctx.lastLineEmpty = true
		}
		return
	}

	// Track if this line starts a new block or ends one
	isStart := isBlockStart(trimmedLine)
	isEnd := isBlockEnd(trimmedLine)

	// Determine if this is a Grimoire or spell declaration
	isGrimoire := strings.HasPrefix(trimmedLine, "grim ")
	isSpell := strings.HasPrefix(trimmedLine, "spell ")

	// Determine if this is a control flow statement
	isControlFlow := isControlFlowStart(trimmedLine)

	// Check if this line should exit spell context (top-level statements)
	isTopLevelStatement := !isGrimoire && !isSpell && !isControlFlow && !isStart && !isEnd && 
		!strings.HasPrefix(trimmedLine, "return ") && !strings.HasPrefix(trimmedLine, "ignore") &&
		!strings.HasPrefix(trimmedLine, "#") && !strings.HasPrefix(trimmedLine, "```") &&
		(strings.Contains(trimmedLine, "=") || strings.Contains(trimmedLine, "("))

	// Handle indentation adjustments before writing the line
	if isEnd && !isStart {
		// End a block
		if len(ctx.blockStack) > 0 {
			ctx.blockStack = ctx.blockStack[:len(ctx.blockStack)-1]
		}
		if ctx.indentLevel > 0 {
			ctx.indentLevel--
		}
	}

	// Special formatting for Grimoire and spell declarations
	if isGrimoire {
		// Start of a Grimoire - reset to base level
		ctx.inGrimoire = true
		ctx.indentLevel = 0

		// Add empty line before Grimoire unless it's the first line
		if lineNum > 0 && !ctx.lastLineEmpty {
			ctx.output.WriteString("\n")
		}
	} else if isSpell {
		// Add empty line between spells in a Grimoire
		if ctx.inSpell && !ctx.lastLineEmpty {
			ctx.output.WriteString("\n")
		}

		// Spells inside a Grimoire should be indented one level
		if ctx.inGrimoire {
			ctx.indentLevel = 1
		} else {
			ctx.indentLevel = 0
		}

		ctx.inSpell = true
	} else if isControlFlow && ctx.inSpell {
		// Control flow statements follow the current indent level
		// No adjustment needed as they will be properly indented
	} else if isTopLevelStatement && ctx.inSpell {
		// This is a top-level statement, exit spell context
		ctx.inSpell = false
		if ctx.inGrimoire {
			ctx.indentLevel = 0  // Return to grimoire level
		} else {
			ctx.indentLevel = 0  // Return to global level
		}
	}

	// Apply current indentation level
	indent := strings.Repeat(ctx.indentStr, ctx.indentLevel)
	ctx.output.WriteString(indent)

	// Process the line content
	formattedLine := f.processLine(trimmedLine)
	ctx.output.WriteString(formattedLine)
	ctx.output.WriteString("\n")

	// Update state for next line
	ctx.lastLineEmpty = false
	ctx.lastLineIndented = isStart && !isEnd

	// Handle indentation adjustments after writing the line
	if isStart && !isEnd {
		// Start a new block and track its type
		blockType := getBlockType(trimmedLine)
		ctx.blockStack = append(ctx.blockStack, blockType)
		ctx.indentLevel++
	}

	// Handle spacing between different sections
	trimmedNextLine := strings.TrimSpace(nextLine)
	if isSpell && ctx.inGrimoire && trimmedNextLine != "" &&
		!strings.HasPrefix(trimmedNextLine, "spell ") {
		// Add a blank line after spell definition if the next line isn't empty or another spell
		ctx.output.WriteString("\n")
		ctx.lastLineEmpty = true
	}
}

// getBlockType determines the type of block being started
func getBlockType(line string) string {
	if strings.HasPrefix(line, "grim ") {
		return "grim"
	} else if strings.HasPrefix(line, "spell ") {
		return "spell"
	} else if strings.HasPrefix(line, "if ") {
		return "if"
	} else if strings.HasPrefix(line, "otherwise") {
		return "otherwise"
	} else if strings.HasPrefix(line, "for ") {
		return "for"
	} else if strings.HasPrefix(line, "while ") {
		return "while"
	} else if strings.HasPrefix(line, "attempt:") {
		return "attempt"
	} else if strings.HasPrefix(line, "ensnare") {
		return "ensnare"
	} else if strings.HasPrefix(line, "resolve:") {
		return "resolve"
	} else if strings.HasPrefix(line, "match ") {
		return "match"
	} else if strings.HasPrefix(line, "case ") {
		return "case"
	}
	return "generic"
}

// isControlFlowStart returns true if the line starts a control flow statement
func isControlFlowStart(line string) bool {
	controlFlowStarters := []string{
		"if ", "otherwise", "else:", "for ", "while ", "attempt:",
		"ensnare", "resolve:", "match ", "case ",
	}

	for _, starter := range controlFlowStarters {
		if strings.HasPrefix(line, starter) {
			return true
		}
	}

	return false
}

// processLine applies formatting rules to a line of code
func (f *CarrionFormatter) processLine(line string) string {
	// 1. Remove extra spaces around operators
	line = formatOperators(line)

	// 2. Ensure proper spacing after commas in function calls, arrays, etc.
	line = formatCommas(line)

	// 3. Ensure proper spacing around colons
	line = formatColons(line)

	// 4. Normalize comments
	line = formatComments(line)

	// 5. Handle parentheses spacing
	line = formatParentheses(line)

	return line
}

// isBlockStart returns true if the line starts a new block
func isBlockStart(line string) bool {
	// Check if the line ends with a colon
	if strings.HasSuffix(strings.TrimSpace(line), ":") {
		return true
	}

	blockStarters := []string{
		"if ", "otherwise", "else:", "for ", "while ", "spell ", "grim ", "attempt:",
		"ensnare", "resolve:", "match ", "case ", "init",
	}

	for _, starter := range blockStarters {
		if strings.HasPrefix(line, starter) {
			return true
		}
	}

	return false
}

// isBlockEnd returns true if the line ends a block
func isBlockEnd(line string) bool {
	blockEnders := []string{
		"}", "end", "endif", "endspell", "endGrimoire", "endfor", "endwhile",
		"endattempt", "endensnare", "endresolve", "endmatch", "endcase",
	}

	for _, ender := range blockEnders {
		if line == ender || strings.HasPrefix(line, ender+" ") {
			return true
		}
	}

	return false
}

// formatOperators ensures proper spacing around operators
func formatOperators(line string) string {
	operators := []string{
		"+", "-", "*", "=", "==", "!=", "<", ">", "<=", ">=",
		"and", "or", "not", "%", "**", "+=", "-=", "*=", "/=",
	}

	// Replace operators with properly spaced versions
	result := line
	for _, op := range operators {
		// Skip certain operator combinations to avoid incorrect replacements
		if op == "+" && (strings.Contains(result, "++") || strings.Contains(result, "+=")) {
			continue
		}
		if op == "-" && (strings.Contains(result, "--") || strings.Contains(result, "-=")) {
			continue
		}

		// Don't add spaces around operators in strings or comments
		inString := false
		inSingleQuoteString := false
		inComment := false
		var processedLine strings.Builder

		for i := 0; i < len(result); i++ {
			if result[i] == '"' && (i == 0 || result[i-1] != '\\') {
				inString = !inString
				processedLine.WriteByte(result[i])
			} else if result[i] == '\'' && (i == 0 || result[i-1] != '\\') {
				inSingleQuoteString = !inSingleQuoteString
				processedLine.WriteByte(result[i])
			} else if result[i] == '#' && !inString && !inSingleQuoteString {
				inComment = true
				processedLine.WriteByte(result[i])
			} else if i < len(result)-1 && result[i:i+2] == "//" && !inString && !inSingleQuoteString {
				// Handle double slash comments - don't format as operators
				processedLine.WriteString("//")
				i++ // Skip the next character
			} else if !inString && !inSingleQuoteString && !inComment && i <= len(result)-len(op) && result[i:i+len(op)] == op {
				// Only format standalone operators, not parts of words
				prevChar := byte(0)
				nextChar := byte(0)
				if i > 0 {
					prevChar = result[i-1]
				}
				if i+len(op) < len(result) {
					nextChar = result[i+len(op)]
				}

				// For "or" and "and", ensure they're standalone words
				if op == "or" || op == "and" {
					isPrevWordChar := isIdentifierChar(prevChar)
					isNextWordChar := isIdentifierChar(nextChar)
					
					if isPrevWordChar || isNextWordChar {
						// This is part of a word, don't format as operator
						processedLine.WriteByte(result[i])
						continue
					}
				}

				// Add spaces around the operator, but only if not in a string
				if i > 0 && result[i-1] != ' ' {
					processedLine.WriteByte(' ')
				}
				processedLine.WriteString(op)
				if i+len(op) < len(result) && result[i+len(op)] != ' ' {
					processedLine.WriteByte(' ')
				}
				i += len(op) - 1
			} else {
				processedLine.WriteByte(result[i])
			}
		}

		result = processedLine.String()
	}

	// Handle division operator separately to avoid breaking // comments
	result = formatDivisionOperator(result)

	return result
}

// formatDivisionOperator handles division operator formatting while preserving // comments
func formatDivisionOperator(line string) string {
	inString := false
	inSingleQuoteString := false
	inComment := false
	var result strings.Builder

	for i := 0; i < len(line); i++ {
		if line[i] == '"' && (i == 0 || line[i-1] != '\\') {
			inString = !inString
			result.WriteByte(line[i])
		} else if line[i] == '\'' && (i == 0 || line[i-1] != '\\') {
			inSingleQuoteString = !inSingleQuoteString
			result.WriteByte(line[i])
		} else if line[i] == '#' && !inString && !inSingleQuoteString {
			inComment = true
			result.WriteByte(line[i])
		} else if i < len(line)-1 && line[i:i+2] == "//" && !inString && !inSingleQuoteString {
			// Preserve // comments
			result.WriteString("//")
			i++ // Skip the next character
		} else if !inString && !inSingleQuoteString && !inComment && line[i] == '/' {
			// Check if this is a standalone division operator
			prevChar := byte(0)
			nextChar := byte(0)
			if i > 0 {
				prevChar = line[i-1]
			}
			if i+1 < len(line) {
				nextChar = line[i+1]
			}

			// Make sure it's not part of // comment or /= operator
			if nextChar != '/' && nextChar != '=' {
				// Add spaces around division operator
				if i > 0 && prevChar != ' ' {
					result.WriteByte(' ')
				}
				result.WriteByte('/')
				if i+1 < len(line) && nextChar != ' ' {
					result.WriteByte(' ')
				}
			} else {
				result.WriteByte(line[i])
			}
		} else {
			result.WriteByte(line[i])
		}
	}

	return result.String()
}

// formatColons ensures proper spacing around colons
func formatColons(line string) string {
	// Don't format colons in strings
	inString := false
	inSingleQuoteString := false
	var result strings.Builder

	for i := 0; i < len(line); i++ {
		if line[i] == '"' && (i == 0 || line[i-1] != '\\') {
			inString = !inString
			result.WriteByte(line[i])
		} else if line[i] == '\'' && (i == 0 || line[i-1] != '\\') {
			inSingleQuoteString = !inSingleQuoteString
			result.WriteByte(line[i])
		} else if !inString && !inSingleQuoteString && line[i] == ':' {
			// Special handling for scope definition colons (if, else, etc.)
			isBlockColon := false
			prevWordStart := i - 1

			// Find start of previous word
			for prevWordStart >= 0 && (line[prevWordStart] == ' ' || line[prevWordStart] == '\t') {
				prevWordStart--
			}

			// Find beginning of the previous word
			wordStart := prevWordStart
			for wordStart >= 0 && isIdentifierChar(line[wordStart]) {
				wordStart--
			}
			wordStart++

			if wordStart <= prevWordStart {
				word := line[wordStart : prevWordStart+1]
				blockWords := []string{
					"if", "else", "for", "while", "spell", "grim",
					"attempt", "ensnare", "resolve", "match", "case", "init",
				}

				for _, blockWord := range blockWords {
					if word == blockWord {
						isBlockColon = true
						break
					}
				}
			}

			if isBlockColon {
				// For block definitions, no space before colon
				result.WriteByte(':')
			} else {
				// For other colons (like in dictionary literals), space after but not before
				if i > 0 && line[i-1] == ' ' {
					// Remove the space before
					resultString := result.String()
					resultRunes := []rune(resultString)
					result = strings.Builder{}
					result.WriteString(string(resultRunes[:len(resultRunes)-1]))
					result.WriteByte(':')
				} else {
					result.WriteByte(':')
				}

				// Add space after if not already there and not at end of line
				if i+1 < len(line) && line[i+1] != ' ' && line[i+1] != '\n' {
					result.WriteByte(' ')
				}
			}
		} else {
			result.WriteByte(line[i])
		}
	}

	return result.String()
}

// formatCommas ensures proper spacing after commas
func formatCommas(line string) string {
	// Don't format commas in strings
	inString := false
	inSingleQuoteString := false
	var result strings.Builder

	for i := 0; i < len(line); i++ {
		if line[i] == '"' && (i == 0 || line[i-1] != '\\') {
			inString = !inString
			result.WriteByte(line[i])
		} else if line[i] == '\'' && (i == 0 || line[i-1] != '\\') {
			inSingleQuoteString = !inSingleQuoteString
			result.WriteByte(line[i])
		} else if !inString && !inSingleQuoteString && line[i] == ',' {
			// Add comma and ensure exactly one space follows
			result.WriteByte(',')

			// Check if we need to add a space after the comma
			if i+1 < len(line) && line[i+1] != ' ' {
				result.WriteByte(' ')
			} else if i+1 < len(line) {
				// Ensure only one space follows
				j := i + 1
				for j < len(line) && line[j] == ' ' {
					j++
				}

				if j > i+2 { // More than one space
					result.WriteByte(' ')
					i = j - 1 // Skip extra spaces
				} else {
					result.WriteByte(' ')
				}
			}
		} else {
			result.WriteByte(line[i])
		}
	}

	return result.String()
}

// formatComments normalizes comments
func formatComments(line string) string {
	// Find comment start (#), ensuring it's not inside a string
	inString := false
	inSingleQuoteString := false
	commentStart := -1

	for i := 0; i < len(line); i++ {
		if line[i] == '"' && (i == 0 || line[i-1] != '\\') {
			inString = !inString
		} else if line[i] == '\'' && (i == 0 || line[i-1] != '\\') {
			inSingleQuoteString = !inSingleQuoteString
		} else if !inString && !inSingleQuoteString && line[i] == '#' {
			commentStart = i
			break
		}
	}

	if commentStart == -1 {
		return line // No comment in this line
	}

	// Ensure there's a space after the # if it's not a line-only comment
	if commentStart > 0 {
		code := strings.TrimSpace(line[:commentStart])
		comment := line[commentStart:]

		if len(comment) > 1 && comment[1] != ' ' {
			comment = "# " + comment[1:]
		}

		// Ensure exactly one space between code and comment
		return code + " " + comment
	}

	// Line-only comment
	comment := line[commentStart:]
	if len(comment) > 1 && comment[1] != ' ' {
		comment = "# " + comment[1:]
	}

	return comment
}

// formatParentheses handles spacing around parentheses
func formatParentheses(line string) string {
	// Don't format parentheses in strings
	inString := false
	inSingleQuoteString := false
	var result strings.Builder

	for i := 0; i < len(line); i++ {
		if line[i] == '"' && (i == 0 || line[i-1] != '\\') {
			inString = !inString
			result.WriteByte(line[i])
		} else if line[i] == '\'' && (i == 0 || line[i-1] != '\\') {
			inSingleQuoteString = !inSingleQuoteString
			result.WriteByte(line[i])
		} else if !inString && !inSingleQuoteString {
			if line[i] == '(' {
				// For opening parenthesis, ensure no space after
				result.WriteByte('(')

				// Skip spaces after opening parenthesis
				j := i + 1
				for j < len(line) && line[j] == ' ' {
					j++
				}

				if j > i+1 {
					i = j - 1 // Skip spaces
				}
			} else if line[i] == ')' {
				// For closing parenthesis, ensure no space before
				// If the last character is a space, remove it
				if result.Len() > 0 {
					resultString := result.String()
					if resultString[len(resultString)-1] == ' ' {
						resultRunes := []rune(resultString)
						result = strings.Builder{}
						result.WriteString(string(resultRunes[:len(resultRunes)-1]))
					}
				}

				result.WriteByte(')')
			} else {
				result.WriteByte(line[i])
			}
		} else {
			result.WriteByte(line[i])
		}
	}

	return result.String()
}

// isIdentifierChar returns true if the character is valid in an identifier
func isIdentifierChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}
