module.exports = grammar({
  name: 'carrion',

  extras: $ => [
    /\s/,
    $.comment,
    $.block_comment
  ],

  word: $ => $.identifier,

  conflicts: $ => [
    [$.expression_statement, $.call_expression],
    [$.expression_statement, $.subscript],
    [$.expression_statement, $.attribute],
    [$.attempt_statement],
    [$.grimoire_definition],
    [$.spell_definition],
    [$.if_statement],
    [$.match_statement],
    [$.for_statement],
    [$.while_statement],
    [$.ensnare_clause],
    [$.resolve_clause],
    [$.elif_clause],
    [$.else_clause],
    [$.case_clause],
  ],

  rules: {
    source_file: $ => repeat($._statement),

    _statement: $ => choice(
      $.import_statement,
      $.grimoire_definition,
      $.spell_definition,
      $.expression_statement,
      $.assignment_statement,
      $.if_statement,
      $.match_statement,
      $.for_statement,
      $.while_statement,
      $.attempt_statement,
      $.return_statement,
      $.raise_statement,
      $.ignore_statement
    ),

    // Import statements
    import_statement: $ => seq(
      'import',
      $.string_literal
    ),

    // Grimoire (class) definitions
    grimoire_definition: $ => seq(
      'grim',
      $.identifier,
      optional(seq('(', commaSep($.identifier), ')')),
      ':',
      optional('\n'),
      repeat($._statement)
    ),

    // Spell (method/function) definitions
    spell_definition: $ => seq(
      'spell',
      $.identifier,
      $.parameter_list,
      ':',
      optional('\n'),
      repeat($._statement)
    ),

    // Parameter list
    parameter_list: $ => seq(
      '(',
      optional(commaSep($.parameter)),
      ')'
    ),

    parameter: $ => seq(
      $.identifier,
      optional(seq('=', $._expression))
    ),

    // Statements
    assignment_statement: $ => seq(
      $._assignable,
      choice('=', '+=', '-=', '*=', '/='),
      $._expression
    ),

    _assignable: $ => choice(
      $.identifier,
      $.attribute,
      $.subscript
    ),

    expression_statement: $ => $._expression,

    if_statement: $ => seq(
      'if',
      $._expression,
      ':',
      optional('\n'),
      repeat($._statement),
      repeat($.elif_clause),
      optional($.else_clause)
    ),

    elif_clause: $ => seq(
      'otherwise',
      $._expression,
      ':',
      optional('\n'),
      repeat($._statement)
    ),

    else_clause: $ => seq(
      'else',
      ':',
      optional('\n'),
      repeat($._statement)
    ),

    match_statement: $ => seq(
      'match',
      $._expression,
      ':',
      optional('\n'),
      repeat($.case_clause)
    ),

    case_clause: $ => seq(
      choice('case', '_'),
      optional($._expression),
      ':',
      optional('\n'),
      repeat($._statement)
    ),

    for_statement: $ => seq(
      'for',
      $.identifier,
      'in',
      $._expression,
      ':',
      optional('\n'),
      repeat($._statement)
    ),

    while_statement: $ => seq(
      'while',
      $._expression,
      ':',
      optional('\n'),
      repeat($._statement)
    ),

    attempt_statement: $ => seq(
      'attempt',
      ':',
      optional('\n'),
      repeat($._statement),
      repeat($.ensnare_clause),
      optional($.resolve_clause)
    ),

    ensnare_clause: $ => seq(
      'ensnare',
      $.identifier,
      ':',
      optional('\n'),
      repeat($._statement)
    ),

    resolve_clause: $ => seq(
      'resolve',
      ':',
      optional('\n'),
      repeat($._statement)
    ),

    return_statement: $ => prec.right(seq(
      'return',
      optional($._expression)
    )),

    raise_statement: $ => seq(
      'raise',
      $._expression
    ),

    ignore_statement: $ => 'ignore',

    // Expressions
    _expression: $ => choice(
      $.binary_expression,
      $.unary_expression,
      $.call_expression,
      $.attribute,
      $.subscript,
      $.list_literal,
      $.dictionary_literal,
      $.tuple_literal,
      $.string_literal,
      $.f_string,
      $.integer_literal,
      $.float_literal,
      $.boolean_literal,
      $.none_literal,
      $.identifier,
      $.parenthesized_expression
    ),

    binary_expression: $ => choice(
      prec.left(10, seq($._expression, choice('+', '-'), $._expression)),
      prec.left(11, seq($._expression, choice('*', '/', '//', '%', '**'), $._expression)),
      prec.left(6, seq($._expression, choice('==', '!=', '<', '>', '<=', '>='), $._expression)),
      prec.left(5, seq($._expression, 'and', $._expression)),
      prec.left(4, seq($._expression, 'or', $._expression))
    ),

    unary_expression: $ => prec(12, seq(
      choice('not', '-', '+'),
      $._expression
    )),

    call_expression: $ => prec(13, seq(
      $._expression,
      $.argument_list
    )),

    argument_list: $ => seq(
      '(',
      optional(commaSep($._expression)),
      ')'
    ),

    attribute: $ => prec(13, seq(
      $._expression,
      '.',
      $.identifier
    )),

    subscript: $ => prec(13, seq(
      $._expression,
      '[',
      $._expression,
      ']'
    )),

    parenthesized_expression: $ => seq(
      '(',
      $._expression,
      ')'
    ),

    // Literals
    list_literal: $ => seq(
      '[',
      optional(commaSep($._expression)),
      ']'
    ),

    dictionary_literal: $ => seq(
      '{',
      optional(commaSep($.pair)),
      '}'
    ),

    pair: $ => seq(
      $._expression,
      ':',
      $._expression
    ),

    tuple_literal: $ => seq(
      '(',
      $._expression,
      ',',
      optional(commaSep($._expression)),
      ')'
    ),

    string_literal: $ => choice(
      seq('"', repeat(choice(/[^"\\]/, $.escape_sequence)), '"'),
      seq("'", repeat(choice(/[^'\\]/, $.escape_sequence)), "'")
    ),

    f_string: $ => seq(
      'f',
      choice(
        seq('"', repeat(choice(/[^"\\{]/, $.escape_sequence, $.f_string_expression)), '"'),
        seq("'", repeat(choice(/[^'\\{]/, $.escape_sequence, $.f_string_expression)), "'")
      )
    ),

    f_string_expression: $ => seq(
      '{',
      $._expression,
      '}'
    ),

    escape_sequence: $ => token(seq(
      '\\',
      choice(
        /[abfnrtv\\'"]/,
        /\d{1,3}/,
        /x[0-9a-fA-F]{2}/,
        /u[0-9a-fA-F]{4}/,
        /U[0-9a-fA-F]{8}/
      )
    )),

    integer_literal: $ => /\d+/,

    float_literal: $ => /\d+\.\d+/,

    boolean_literal: $ => choice('True', 'False'),

    none_literal: $ => 'None',

    identifier: $ => /[a-zA-Z_][a-zA-Z0-9_]*/,

    // Comments
    comment: $ => token(seq('#', /.*/)),

    block_comment: $ => token(seq(
      '```',
      repeat(choice(/[^`]/, /`[^`]/, /``[^`]/)),
      '```'
    )),
  }
});

function commaSep(rule) {
  return optional(commaSep1(rule));
}

function commaSep1(rule) {
  return seq(rule, repeat(seq(',', rule)));
}