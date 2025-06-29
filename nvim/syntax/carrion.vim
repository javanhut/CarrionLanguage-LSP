" Vim syntax file for Carrion language
" Language: Carrion
" Maintainer: Carrion Language Team
" Latest Revision: 2024

if exists("b:current_syntax")
  finish
endif

" Keywords - Carrion-specific magical terminology
syntax keyword carrionKeyword spell grim init self if else otherwise
syntax keyword carrionKeyword for in while stop skip ignore return import
syntax keyword carrionKeyword match case attempt resolve ensnare raise as
syntax keyword carrionKeyword arcane arcanespell super check and or not

" Built-in types and constants
syntax keyword carrionBoolean True False None
syntax keyword carrionBuiltin print len input int float str type enumerate
syntax keyword carrionBuiltin help version modules

" Comments - all three styles supported by Carrion
syntax match carrionComment "#.*$"
syntax region carrionBlockComment start="/\*" end="\*/" contains=carrionTodo
syntax region carrionTripleBacktickComment start="```" end="```" contains=carrionTodo

" TODOs and FIXMEs in comments
syntax keyword carrionTodo TODO FIXME XXX NOTE contained

" Strings - single, double, and f-strings
syntax region carrionString start=+"+ skip=+\\\\\|\\"+ end=+"+ contains=carrionEscape,carrionInterpolation
syntax region carrionString start=+'+ skip=+\\\\\|\\'+ end=+'+ contains=carrionEscape
syntax region carrionFString start=+f"+ skip=+\\\\\|\\"+ end=+"+ contains=carrionEscape,carrionInterpolation
syntax region carrionFString start=+f'+ skip=+\\\\\|\\'+ end=+'+ contains=carrionEscape,carrionInterpolation

" String interpolation in f-strings
syntax region carrionInterpolation matchgroup=carrionDelimiter start=+{+ end=+}+ contained contains=@carrionAll

" Docstrings - triple quoted strings
syntax region carrionDocstring start=+"""+ end=+"""+ 
syntax region carrionDocstring start=+'''+ end=+'''+

" String escape sequences
syntax match carrionEscape +\\[abfnrtv'"\\]+ contained
syntax match carrionEscape "\\[0-7]\{1,3}" contained
syntax match carrionEscape "\\x[0-9a-fA-F]\{2}" contained
syntax match carrionEscape "\\u[0-9a-fA-F]\{4}" contained
syntax match carrionEscape "\\U[0-9a-fA-F]\{8}" contained

" Numbers
syntax match carrionNumber "\<\d\+\>"
syntax match carrionNumber "\<0[xX][0-9a-fA-F]\+\>"
syntax match carrionNumber "\<0[bB][01]\+\>"
syntax match carrionNumber "\<0[oO]\o\+\>"
syntax match carrionFloat "\<\d\+\.\d*\([eE][+-]\?\d\+\)\?\>"
syntax match carrionFloat "\<\.\d\+\([eE][+-]\?\d\+\)\?\>"
syntax match carrionFloat "\<\d\+[eE][+-]\?\d\+\>"

" Operators
syntax match carrionOperator "+"
syntax match carrionOperator "-"
syntax match carrionOperator "\*"
syntax match carrionOperator "/"
syntax match carrionOperator "//"
syntax match carrionOperator "%"
syntax match carrionOperator "\*\*"
syntax match carrionOperator "="
syntax match carrionOperator "+="
syntax match carrionOperator "-="
syntax match carrionOperator "\*="
syntax match carrionOperator "/="
syntax match carrionOperator "=="
syntax match carrionOperator "!="
syntax match carrionOperator "<"
syntax match carrionOperator ">"
syntax match carrionOperator "<="
syntax match carrionOperator ">="

" Delimiters
syntax match carrionDelimiter "("
syntax match carrionDelimiter ")"
syntax match carrionDelimiter "{"
syntax match carrionDelimiter "}"
syntax match carrionDelimiter "\["
syntax match carrionDelimiter "\]"
syntax match carrionDelimiter ","
syntax match carrionDelimiter ":"
syntax match carrionDelimiter "\."

" Function and class definitions
syntax match carrionFunction "\<spell\s\+\w\+\ze\s*(" contains=carrionKeyword
syntax match carrionClass "\<grim\s\+\w\+\ze\s*[:(]" contains=carrionKeyword

" Highlight groups
highlight default link carrionKeyword Keyword
highlight default link carrionBoolean Boolean
highlight default link carrionBuiltin Function
highlight default link carrionComment Comment
highlight default link carrionBlockComment Comment
highlight default link carrionTripleBacktickComment Comment
highlight default link carrionTodo Todo
highlight default link carrionString String
highlight default link carrionFString String
highlight default link carrionDocstring String
highlight default link carrionEscape Special
highlight default link carrionInterpolation Special
highlight default link carrionNumber Number
highlight default link carrionFloat Float
highlight default link carrionOperator Operator
highlight default link carrionDelimiter Delimiter
highlight default link carrionFunction Function
highlight default link carrionClass Type

" Define clusters for interpolation
syntax cluster carrionAll contains=carrionKeyword,carrionBoolean,carrionBuiltin,carrionString,carrionNumber,carrionFloat,carrionOperator,carrionDelimiter

let b:current_syntax = "carrion"