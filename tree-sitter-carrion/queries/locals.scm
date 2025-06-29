; Scopes
(grimoire_definition) @scope
(spell_definition) @scope
(if_statement) @scope
(for_statement) @scope
(while_statement) @scope
(attempt_statement) @scope

; Definitions
(grimoire_definition
  (identifier) @definition.type)

(spell_definition
  (identifier) @definition.function)

(parameter
  (identifier) @definition.parameter)

(assignment_statement
  (identifier) @definition.variable)

; References
(identifier) @reference