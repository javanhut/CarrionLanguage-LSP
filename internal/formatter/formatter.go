package formatter

import (
	"strings"

	"github.com/carrionlang-lsp/lsp/internal/protocol"
	"github.com/carrionlang-lsp/lsp/internal/util"

	"github.com/javanhut/TheCarrionLanguage/src/lexer"
	"github.com/javanhut/TheCarrionLanguage/src/parser"

	lsp "go.lsp.dev/protocol"
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
	indentLevel int
	indentStr   string
	output      *strings.Builder
}

// formatDocument formats a Carrion document
func (f *CarrionFormatter) formatDocument(text string) string {
	// Create a formatter context
	ctx := &formatterContext{
		indentLevel: 0,
		indentStr:   "    ", // 4 spaces for indentation
		output:      &strings.Builder{},
	}

	// Split the document into lines
	lines := strings.Split(text, "\n")

	// Process each line with correct indentation
	for i, line := range lines {
		f.formatLine(line, i, ctx)
	}

	return ctx.output.String()
}

// formatLine formats a single line of code
func (f *CarrionFormatter) formatLine(line string, lineNum int, ctx *formatterContext) {
	// Skip empty lines
	trimmedLine := strings.TrimSpace(line)
	if trimmedLine == "" {
		ctx.output.WriteString("\n")
		return
	}

	// Handle indentation for block structures
	if isBlockEnd(trimmedLine) {
		if ctx.indentLevel > 0 {
			ctx.indentLevel--
		}
	}

	// Apply current indentation level
	for i := 0; i < ctx.indentLevel; i++ {
		ctx.output.WriteString(ctx.indentStr)
	}

	// Process the line content
	formattedLine := f.processLine(trimmedLine)
	ctx.output.WriteString(formattedLine)
	ctx.output.WriteString("\n")

	// Check if this line starts a new block
	if isBlockStart(trimmedLine) {
		ctx.indentLevel++
	}
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
	blockStarters := []string{
		"if ", "otherwise ", "else:", "for ", "while ", "spell ", "spellbook ", "attempt:",
		"ensnare", "resolve:", "match ", "case ", "init",
	}

	// Check if the line ends with a colon
	if strings.HasSuffix(strings.TrimSpace(line), ":") {
		return true
	}

	for _, starter := range blockStarters {
		if strings.HasPrefix(strings.TrimSpace(line), starter) {
			return true
		}
	}

	return false
}

// isBlockEnd returns true if the line ends a block
func isBlockEnd(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "}")
}

// formatOperators ensures proper spacing around operators
func formatOperators(line string) string {
	operators := []string{
		"+", "-", "*", "/", "=", "==", "!=", "<", ">", "<=", ">=",
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

		// Don't add spaces around operators in strings
		inString := false
		inSingleQuoteString := false
		var processedLine strings.Builder

		for i := 0; i < len(result); i++ {
			if result[i] == '"' && (i == 0 || result[i-1] != '\\') {
				inString = !inString
				processedLine.WriteByte(result[i])
			} else if result[i] == '\'' && (i == 0 || result[i-1] != '\\') {
				inSingleQuoteString = !inSingleQuoteString
				processedLine.WriteByte(result[i])
			} else if !inString && !inSingleQuoteString && i <= len(result)-len(op) && result[i:i+len(op)] == op {
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

	return result
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
					"if", "else", "for", "while", "spell", "spellbook",
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
	// Find comment start (//), ensuring it's not inside a string
	inString := false
	inSingleQuoteString := false
	commentStart := -1

	for i := 0; i < len(line); i++ {
		if line[i] == '"' && (i == 0 || line[i-1] != '\\') {
			inString = !inString
		} else if line[i] == '\'' && (i == 0 || line[i-1] != '\\') {
			inSingleQuoteString = !inSingleQuoteString
		} else if !inString && !inSingleQuoteString && i < len(line)-1 && line[i] == '/' && line[i+1] == '/' {
			commentStart = i
			break
		}
	}

	if commentStart == -1 {
		return line // No comment in this line
	}

	// Ensure there's a space after the // if it's not a line-only comment
	if commentStart > 0 {
		code := strings.TrimSpace(line[:commentStart])
		comment := line[commentStart:]

		if len(comment) > 2 && comment[2] != ' ' {
			comment = "// " + comment[2:]
		}

		// Ensure exactly one space between code and comment
		return code + " " + comment
	}

	// Line-only comment
	comment := line[commentStart:]
	if len(comment) > 2 && comment[2] != ' ' {
		comment = "// " + comment[2:]
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
