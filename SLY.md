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

Variables are global and are not scoped to slides or blocks.

Variables may be shadowed by redeclaring them.

## Attribute Declaration

Sly allows you to control specific characteristics of the current slide using an attribute declaration.

```
@attributeName = "value";
```

As you can see, the syntax is very similar to variable declaration, but preceded with an @ symbol.

There is also a fixed set of attributes which you can define. They are as follows:

- backgroundColor
    - the background color of the slide. Can be either the name of a color (ex: "black") or a color literal.
- font
    - the font of a text block. Must be a string value.
- fontColor
    - the font color of a text block. Can be either the name of a color (ex: "black") or a color literal.
- fontSize
    - the font size of a text block. Must be an integer value.
- justify
    - the justification for a text block. Accepted values are "left", "center", or "right".
    
Slide level attributes will be inherited by following slides. Block level attributes will be inherited by later blocks, but are reset between slides.

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

## Text

The Text declaration is what you use to actually write the content for your slides.

```
---This is my first line of text---

---
This is my
Second line of text
---
```

This can be multi-line or single-line, up to you!

## Slide Scopes

These signify the start of a new slide. A title slide is assumed, so you only need to use this from your second slide onwards.

```
[Name of my slide]

---Hello!---
```

The text between the square brackets is not currently used by Slydes; scope title text is purely for organizational purposes.

## Block Scopes

This represents a unique, styled section of text.

```
[Starting Slide]

[[My Text]]
---
Boo!
---
```

Like slide scopes, the title text not used in processing.

## Macros

To cut down on repetition, Sly provides macro functionality.

```
green = (0, 255, 0);

$themeMacro = {
    foo = "hello";
    @font = "Fira Code";

    # Can refer to outside variables
    @fontColor = green;

    # Variables only need to be instantiated
    # by the time the macro is called, not when
    # it is defined
    @fontSize = defaultFontSize;
};

[Example Section]
defaultFontSize = 14;

# Will expand to the statement block originally
# defined in the macro
$themeMacro()
```

Macros are very basic in Sly. When called they simply expand to the statement block originally provided.

There is no concept of scope within a macro as each macro expansion is inlined.