# tree-sitter-carrion

A Tree-sitter grammar for the Carrion programming language.

## Features

- Full syntax highlighting for Carrion language constructs
- Support for grimoires (classes), spells (methods), and all control flow statements
- Proper handling of f-strings, comments, and literals
- Code folding and local variable scoping
- Integration with editors that support Tree-sitter

## Installation

```bash
npm install tree-sitter-carrion
```

## Usage

### With Node.js

```javascript
const Parser = require('tree-sitter');
const Carrion = require('tree-sitter-carrion');

const parser = new Parser();
parser.setLanguage(Carrion);

const sourceCode = `
grim Calculator:
    spell add(x, y):
        return x + y

calc = Calculator()
print(calc.add(5, 3))
`;

const tree = parser.parse(sourceCode);
console.log(tree.rootNode.toString());
```

### With Editors

This grammar can be used with any editor that supports Tree-sitter, including:

- Neovim
- Emacs
- Helix
- Zed
- And many others

## Language Features Supported

- **Grimoires** (`grim`): Class definitions with inheritance
- **Spells** (`spell`): Method and function definitions  
- **Control Flow**: `if`/`otherwise`/`else`, `for`, `while`, `match`/`case`
- **Error Handling**: `attempt`/`ensnare`/`resolve`
- **Literals**: Strings, f-strings, numbers, lists, dictionaries, tuples
- **Comments**: Single-line (`#`) and block (` ``` `) comments
- **Expressions**: Binary operations, function calls, attribute access, subscripting

## Example

```carrion
import "math"

grim MagicalCreature:
    ```
    Base class for magical creatures
    ```
    
    init(name, power = 100):
        self.name = name
        self.power = power
    
    spell cast_spell(target, spell_name = "fireball"):
        if self.power < 10:
            raise "Not enough power"
        
        self.power -= 10
        return f"{self.name} casts {spell_name} on {target}!"

# Create and use a creature
creature = MagicalCreature("Wizard")
print(creature.cast_spell("Dragon"))
```

## Development

To build the parser:

```bash
npm install
npx tree-sitter generate
```

To test:

```bash
npx tree-sitter parse examples/test.crl
```

## License

MIT