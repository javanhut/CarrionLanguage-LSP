; Keywords
[
  "grim"
  "spell"
  "import"
  "if"
  "otherwise"
  "else"
  "for"
  "while"
  "match"
  "case"
  "attempt"
  "ensnare"
  "resolve"
  "return"
  "raise"
  "in"
  "and"
  "or"
  "not"
] @keyword

; Special constants
[
  (boolean_literal)
  (none_literal)
] @constant.builtin

(ignore_statement) @constant.builtin

; Function definitions
(spell_definition
  (identifier) @function)

(grimoire_definition
  (identifier) @type)

; Function calls
(call_expression
  (identifier) @function.call)

(call_expression
  (attribute
    (identifier) @function.method.call))

; Operators
[
  "+"
  "-"
  "*"
  "/"
  "//"
  "%"
  "**"
  "="
  "+="
  "-="
  "*="
  "/="
  "=="
  "!="
  "<"
  ">"
  "<="
  ">="
] @operator

; Punctuation
[
  "("
  ")"
  "["
  "]"
  "{"
  "}"
  ","
  ":"
  "."
] @punctuation.delimiter

; Literals
(string_literal) @string
(f_string) @string
(integer_literal) @number
(float_literal) @number

; Comments
(comment) @comment
(block_comment) @comment.block

; Variables
(identifier) @variable

; Parameters
(parameter
  (identifier) @variable.parameter)

; Attributes
(attribute
  (identifier) @property)

; Escape sequences
(escape_sequence) @string.escape