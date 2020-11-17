package lang

import (
	"errors"
	"fmt"
	"image/color"

	"github.com/mbStavola/slydes/pkg/types"
)

type ScopeType int

const (
	InvalidScope ScopeType = iota

	FileScope
	SlideScope
	BlockScope
)

func (s ScopeType) String() string {
	return []string{
		"InvalidScope",

		"FileScope",
		"SlideScope",
		"BlockScope",
	}[s]
}

type variableValue struct {
	isMutable bool
	value     interface{}
}

type scope struct {
	Type      ScopeType
	parent    *scope
	slides    map[string]types.Slide
	blocks    map[string]types.Block
	variables map[string]variableValue
	macros    map[string][]Statement
}

func newTopLevelScope() *scope {
	scope := new(scope)
	scope.Type = FileScope
	scope.slides = make(map[string]types.Slide)
	scope.blocks = make(map[string]types.Block)
	scope.variables = make(map[string]variableValue)
	scope.macros = make(map[string][]Statement)

	return scope
}

func (s *scope) declareVariable(token Token, isMutable bool, name string, value interface{}) error {
	if _, ok := s.variables[name]; !ok {
		s.variables[name] = variableValue{
			isMutable: isMutable,
			value:     value,
		}
		return nil
	}

	return tokenErrorInfo(token, compilation, "variable already declared in this scope")
}

func (s *scope) setVariable(token Token, name string, value interface{}) error {
	variable, ok := s.variables[name]
	if !ok && s.parent != nil {
		return s.parent.setVariable(token, name, value)
	} else if !ok {
		return tokenErrorInfo(token, compilation, "cannot assign to undeclared variable")
	}

	if !variable.isMutable {
		return tokenErrorInfo(token, compilation, "cannot assign to an immutable binding")
	}

	variable.value = value

	return nil
}

func (s *scope) getVariable(token Token, name string) (interface{}, error) {
	variable, ok := s.variables[name]
	if !ok && s.parent != nil {
		return s.parent.getVariable(token, name)
	} else if !ok {
		return nil, tokenErrorInfo(token, compilation, "variable must be initialized before dereference")
	}

	return variable.value, nil
}

func (s *scope) declareMacro(token Token, name string, statements []Statement) error {
	if _, ok := s.macros[name]; !ok {
		s.macros[name] = statements
		return nil
	}

	return tokenErrorInfo(token, compilation, "macro already declared in this scope")
}

func (s *scope) getMacro(token Token, name string) ([]Statement, error) {
	statements, ok := s.macros[name]
	if !ok && s.parent != nil {
		return s.parent.getMacro(token, name)
	} else if !ok {
		return nil, tokenErrorInfo(token, compilation, "macro must be defined before use")
	}

	return statements, nil
}

type Compiler interface {
	Compile(statements []Statement) (types.Show, error)
}

type DefaultCompiler struct{}

func NewDefaultCompiler() DefaultCompiler {
	return DefaultCompiler{}
}

func (comp DefaultCompiler) Compile(statements []Statement) (types.Show, error) {
	state := newCompilationState()
	errBundle := newErrorInfoBundle()

	for _, statement := range statements {
		if err := state.processStatement(statement); err != nil && errors.As(err, &ErrorInfo{}) {
			errBundle.Add(err.(ErrorInfo))
		} else if err != nil {
			return types.Show{}, err
		}
	}

	if errBundle.HasErrors() {
		return types.Show{}, errBundle
	}

	return state.finalizeCompilation(), nil
}

type compilationState struct {
	show  types.Show
	slide *types.Slide
	block *types.Block
	scope *scope
}

func newCompilationState() compilationState {
	show := types.NewShow()

	return compilationState{
		show:  show,
		slide: nil,
		block: nil,
		scope: newTopLevelScope(),
	}
}

func (cs *compilationState) openScope(ty ScopeType) {
	scope := new(scope)
	scope.Type = ty
	scope.parent = cs.scope
	scope.slides = make(map[string]types.Slide)
	scope.blocks = make(map[string]types.Block)
	scope.variables = make(map[string]variableValue)
	scope.macros = make(map[string][]Statement)

	cs.scope = scope
}

func (cs *compilationState) closeScope() {
	cs.scope = cs.scope.parent
}

func (cs *compilationState) processStatement(statement Statement) error {
	switch statement.Type {
	case SlideDecl:
		decl := statement.data.(SlideDeclaration)

		if cs.scope.Type != FileScope {
			return tokenErrorInfo(statement.token, compilation, "A slide may only be defined at the top level")
		}

		slide := types.NewSlide()
		cs.slide = &slide

		// If the slide has a parent, copy the parent's attributes
		if decl.parent != "" {
			parent, ok := cs.scope.slides[decl.parent]
			if !ok {
				return tokenErrorInfo(statement.token, compilation, "Cannot inherit from an undefined slide")
			}

			slide.Background = parent.Background
		}

		cs.openScope(SlideScope)
		for _, statement := range decl.statements {
			if err := cs.processStatement(statement); err != nil {
				return err
			}
		}
		cs.closeScope()

		cs.show.Slides = append(cs.show.Slides, slide)
		cs.scope.slides[decl.name] = slide
	case BlockDecl:
		decl := statement.data.(BlockDeclaration)

		if cs.scope.Type != SlideScope {
			return tokenErrorInfo(statement.token, compilation, "A block may only be defined within a slide")
		}

		block := types.NewBlock()
		cs.block = &block

		// If the block has a parent, copy the parent's attributes
		if decl.parent != "" {
			parent, ok := cs.scope.blocks[decl.parent]
			if !ok {
				return tokenErrorInfo(statement.token, compilation, "Cannot inherit from an undefined block")
			}

			block.Style = parent.Style
		}

		cs.openScope(BlockScope)
		for _, statement := range decl.statements {
			if err := cs.processStatement(statement); err != nil {
				return err
			}
		}
		cs.closeScope()

		cs.slide.Blocks = append(cs.slide.Blocks, block)
		cs.scope.blocks[decl.name] = block
	case WordBlock:
		if cs.scope.Type != BlockScope {
			return tokenErrorInfo(statement.token, compilation, "Text may only be defined within a block")
		}

		cs.block.Words = statement.data.(string)
	case VariableDeclaration:
		variable := statement.data.(VariableDeclStatement)

		var value interface{}
		switch data := variable.value.(type) {
		case uint8, string, ColorLiteral:
			value = data
		case VariableReference:
			dereferenced, err := cs.scope.getVariable(statement.token, data.reference)
			if err != nil {
				return err
			}

			value = dereferenced
		}

		if err := cs.scope.declareVariable(statement.token, variable.isMutable, variable.name, value); err != nil {
			return err
		}
	case VariableAssignment:
		variable := statement.data.(VariableStatement)

		var value interface{}
		switch data := variable.value.(type) {
		case uint8, string, ColorLiteral:
			value = data
		case VariableReference:
			dereferenced, err := cs.scope.getVariable(statement.token, data.reference)
			if err != nil {
				return err
			}

			value = dereferenced
		}

		if err := cs.scope.setVariable(statement.token, variable.name, value); err != nil {
			return err
		}
	case AttributeAssignment:
		attribute := statement.data.(AttributeStatement)

		switch attribute.name {
		case "backgroundColor":
			if cs.scope.Type != SlideScope {
				return tokenErrorInfo(statement.token, compilation, "backgroundColor attribute is only available for slides")
			}

			switch value := attribute.value.(type) {
			case VariableReference:
				val, err := cs.scope.getVariable(statement.token, value.reference)
				if err != nil {
					return err
				}

				c, err := colorFromLiteral(statement.token, val)
				if err != nil {
					return err
				}

				cs.slide.Background = c
			default:
				c, err := colorFromLiteral(statement.token, value)
				if err != nil {
					return err
				}

				cs.slide.Background = c
			}
		case "justify":
			if cs.scope.Type != BlockScope {
				return tokenErrorInfo(statement.token, compilation, "justify attribute is only available for blocks")
			}

			switch value := attribute.value.(type) {
			case VariableReference:
				val, err := cs.scope.getVariable(statement.token, value.reference)
				if err != nil {
					return err
				}

				justification, err := justificationFromLiteral(statement.token, val)
				if err != nil {
					return err
				}

				cs.block.Style.Justification = justification
			case string:
				justification, err := justificationFromLiteral(statement.token, value)
				if err != nil {
					return err
				}

				cs.block.Style.Justification = justification
			}
		case "font":
			if cs.scope.Type != BlockScope {
				return tokenErrorInfo(statement.token, compilation, "font attribute is only available for blocks")
			}

			switch value := attribute.value.(type) {
			case VariableReference:
				val, err := cs.scope.getVariable(statement.token, value.reference)
				if err != nil {
					return err
				}

				switch val := val.(type) {
				case string:
					cs.block.Style.Font = val
				default:
					return tokenErrorInfo(statement.token, compilation, "Font attribute must be a string")
				}
			case string:
				cs.block.Style.Font = value
			}
		case "fontColor":
			if cs.scope.Type != BlockScope {
				return tokenErrorInfo(statement.token, compilation, "fontColor attribute is only available for blocks")
			}

			switch value := attribute.value.(type) {
			case VariableReference:
				val, err := cs.scope.getVariable(statement.token, value.reference)
				if err != nil {
					return err
				}

				c, err := colorFromLiteral(statement.token, val)
				if err != nil {
					return err
				}

				cs.block.Style.Color = c
			default:
				c, err := colorFromLiteral(statement.token, value)
				if err != nil {
					return err
				}

				cs.block.Style.Color = c
			}
		case "fontSize":
			if cs.scope.Type != BlockScope {
				return tokenErrorInfo(statement.token, compilation, "fontSize attribute is only available for blocks")
			}

			switch value := attribute.value.(type) {
			case VariableReference:
				val, err := cs.scope.getVariable(statement.token, value.reference)
				if err != nil {
					return err
				}

				size, ok := val.(uint8)
				if !ok {
					return tokenErrorInfo(statement.token, compilation, "Font size attribute must be an integer")
				}

				cs.block.Style.Size = size
			case uint8:
				cs.block.Style.Size = value
			default:
				return tokenErrorInfo(statement.token, compilation, "Font size attribute must be an integer")
			}
		default:
			return tokenErrorInfo(statement.token, compilation, "Unrecognized attribute")
		}
	case MacroDecl:
		macroDef := statement.data.(MacroDeclaration)

		if err := cs.scope.declareMacro(statement.token, macroDef.name, macroDef.statements); err != nil {
			return err
		}
	case MacroCall:
		macroCall := statement.data.(MacroInvocation)

		statements, err := cs.scope.getMacro(statement.token, macroCall.reference)
		if err != nil {
			return err
		}

		for _, statement := range statements {
			if err := cs.processStatement(statement); err != nil {
				return err
			}
		}
	}

	return nil
}

func (cs *compilationState) finalizeCompilation() types.Show {
	return cs.show
}

func justificationFromLiteral(token Token, value interface{}) (types.Justification, error) {
	switch value := value.(type) {
	case string:
		switch value {
		case "left":
			return types.Left, nil
		case "right":
			return types.Right, nil
		case "center":
			return types.Center, nil
		}
	}

	message := "Justification attribute must be either 'left', 'right', or 'center'"
	return types.Left, tokenErrorInfo(token, compilation, message)
}

func colorFromLiteral(token Token, value interface{}) (color.Color, error) {
	switch value := value.(type) {
	case string:
		switch value {
		case "white":
			return color.White, nil
		case "black":
			return color.Black, nil
		case "red":
			return color.RGBA{
				R: 255,
				G: 0,
				B: 0,
				A: 255,
			}, nil
		case "blue":
			return color.RGBA{
				R: 0,
				G: 0,
				B: 255,
				A: 255,
			}, nil
		case "green":
			return color.RGBA{
				R: 0,
				G: 255,
				B: 0,
				A: 255,
			}, nil
		default:
			message := fmt.Sprintf("Unsupported color '%s'", value)
			return nil, tokenErrorInfo(token, compilation, message)
		}
	case ColorLiteral:
		return color.RGBA{
			R: value.r,
			G: value.g,
			B: value.b,
			A: value.a,
		}, nil
	default:
		fmt.Printf("ff %v\n", value)
		return nil, tokenErrorInfo(token, compilation, "Color attribute must be either a tuple or string")
	}
}
