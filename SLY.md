# The Sly Markup Language

Slydes processes Sly markup, which describes a presentation.

## Comments

Comments start with `#` and ignore the rest of the line

## Variable Declaration

You can define variables with Sly, the syntax is pretty simple:

```
let variableName = 1;
mut otherVarName = "f";
```

A variable declaration is composed of three parts: mutability binding, identifier, and an initial value.

A variable can either be bound as immutable (`let`) or mutable (`mut`). Immutable variables cannot be reassigned after initialization whereas mutables ones can.

The identifier for the variable must start with a letter but can contain any alphanumeric character subsequently.

The possible value types for a variable will be outlined in the Data Types section.

All variable declarations must end with a semicolon.

You can use a variable in an attribute declaration or possibly in another variable declaration.

```
let variable1 = "hello";
let variable2 = variable1;
```

Variables are scoped to the slides or blocks that defined them and you may refer to variables from the enclosing scope(s).

Variables may be shadowed by redeclaring them.

## Data Types

Sly supports some very (very) basic data types.

- string
    - Any set of character between quotes.
    - Ex: "Hello World!"
- integer
    - An unsigned, eight-bit integer (intentionally limited, subject to change).
    - Ex: 42
- color literal
    - A compound type representing an RGB or RGBA color value.
    - Trailing comma optional 
    - Ex: (255, 0, 0)
    - Ex: (255, 255, 0, 255,)
    
In the future we may support these types as well:

- larger integral values
- floating point values
- booleans
- multiline strings

## Slide Scopes

These signify the start of a new slide.

```
slide exampleSlide {
    let foo = "red";
    self.backgroundColor = foo;

    block title {
        ---Hello!---
    }
}
```

You may define variables and blocks.

Sly allows you to control specific characteristics of the current slide using an attribute declaration. Currently, slides support the following attributes:

- backgroundColor
    - the background color of the slide. Can be either the name of a color (ex: "black") or a color literal.

Sly also supports a limited form of inheritance for slides, where the child slide will copy all the attributes defined on the parent slide.

```
slide foo {
    self.backgroundColor = "red";
}

slide bar : foo {
    # The background color of this slide will start off as red
}
```

A slide scope must be defined at the top level.

## Block Scopes

This represents a unique, styled section of text.

```
block foo {
    ---Boo!---
}
```

The block must end with *exactly* one text declaration.

The text declaration can be multi-line or single-line, up to you!

```
block foo {
    ---
    This is my
        Second line of text
    ---
}
```

Blocks also have attributes. The following attributes are currently supported:

- font
    - the font of a text block. Must be a string value.
- fontColor
    - the font color of a text block. Can be either the name of a color (ex: "black") or a color literal.
- fontSize
    - the font size of a text block. Must be an integer value.
- justify
    - the justification for a text block. Accepted values are "left", "center", or "right".

Like slide scopes, you can use inheritance to copy styles between blocks in the same scope.

```
block foo {
    self.font = "Fira Code";
    self.fontColor = "red";

    ---Hello!---
}

block bar : foo {
    self.fontColor = "blue";

    ---Boo!---
}
```

A block scope must be defined within a slide scope.

## Macros

To cut down on repetition, Sly provides macro functionality.

```
let green = (0, 255, 0);

macro themeMacro {
    let foo = "hello";
    self.font = "Fira Code";

    # Can refer to outside variables
    self.fontColor = green;

    # Variables only need to be instantiated
    # by the time the macro is called, not when
    # it is defined
    self.fontSize = defaultFontSize;
}

slide example {
    let defaultFontSize = 14;

    # Will expand to the statement block originally
    # defined in the macro
    $themeMacro();
}
```

Macros are very basic in Sly. When called they simply expand to the statement block originally provided.

There is no concept of scope within a macro as each macro expansion is inlined.