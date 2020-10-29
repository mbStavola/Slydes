package lang

import (
	"fmt"
	"github.com/mbStavola/slydes/pkg/types"
	"image/color"
)

type Compiler interface {
	Compile(statements []Statement) (types.Show, error)
}

type DefaultCompiler struct{}

func NewDefaultCompiler() DefaultCompiler {
	return DefaultCompiler{}
}

func (comp DefaultCompiler) Compile(statements []Statement) (types.Show, error) {
	state := newCompilationState()

	for _, statement := range statements {
		if err := state.processStatement(statement); err != nil {
			return types.Show{}, err
		}
	}

	return state.finalizeCompilation(), nil
}

type compilationState struct {
	show      types.Show
	slide     types.Slide
	block     types.Block
	variables map[string]interface{}
	macros    map[string][]Statement
}

func newCompilationState() compilationState {
	show := types.NewShow()
	variables := make(map[string]interface{})
	macros := make(map[string][]Statement)

	var slide = types.NewSlide()
	var block = types.NewBlock()

	return compilationState{
		show:      show,
		slide:     slide,
		block:     block,
		variables: variables,
		macros:    macros,
	}
}

func (cs *compilationState) processStatement(statement Statement) error {
	switch statement.Type {
	case SlideDecl:
		cs.slide.Blocks = append(cs.slide.Blocks, cs.block)
		cs.show.Slides = append(cs.show.Slides, cs.slide)

		cs.block = types.NewBlock()
		cs.slide = types.Slide{
			Background: cs.slide.Background,
			Blocks:     make([]types.Block, 0),
		}

		break
	case ScopeDecl:
		cs.slide.Blocks = append(cs.slide.Blocks, cs.block)

		// Copy Style from previous block to make things
		// less tedious when writing multiple blocks
		cs.block = types.Block{Style: cs.block.Style}

		break
	case WordBlock:
		cs.block.Words = statement.data.(string)

		break
	case VariableAssignment:
		variable := statement.data.(VariableStatement)

		switch value := variable.value.(type) {
		case uint8, string, ColorLiteral:
			cs.variables[variable.name] = value

			break
		case VariableReference:
			cs.variables[variable.name] = cs.variables[value.reference]
		}

		break
	case AttributeAssignment:
		attribute := statement.data.(AttributeStatement)

		switch attribute.name {
		case "backgroundColor":
			switch value := attribute.value.(type) {
			case VariableReference:
				c, err := colorFromLiteral(statement.token, cs.variables[value.reference])
				if err != nil {
					return err
				}

				cs.slide.Background = c

				break
			default:
				c, err := colorFromLiteral(statement.token, value)
				if err != nil {
					return err
				}

				cs.slide.Background = c
			}

			break
		case "justify":
			switch value := attribute.value.(type) {
			case VariableReference:
				justification, err := justificationFromLiteral(statement.token, cs.variables[value.reference])
				if err != nil {
					return err
				}

				cs.block.Style.Justification = justification

				break
			case string:
				justification, err := justificationFromLiteral(statement.token, value)
				if err != nil {
					return err
				}

				cs.block.Style.Justification = justification
			}

			break
		case "font":
			switch value := attribute.value.(type) {
			case VariableReference:
				val := cs.variables[value.reference]
				switch val := val.(type) {
				case string:
					cs.block.Style.Font = val
				default:
					return tokenErrorInfo(statement.token, "Font attribute must be a string")
				}

				break
			case string:
				cs.block.Style.Font = value
			}

			break
		case "fontColor":
			switch value := attribute.value.(type) {
			case VariableReference:
				c, err := colorFromLiteral(statement.token, cs.variables[value.reference])
				if err != nil {
					return err
				}

				cs.block.Style.Color = c

				break
			default:
				c, err := colorFromLiteral(statement.token, value)
				if err != nil {
					return err
				}

				cs.block.Style.Color = c
			}

			break
		case "fontSize":
			switch value := attribute.value.(type) {
			case VariableReference:
				size, ok := cs.variables[value.reference].(uint8)
				if !ok {
					return tokenErrorInfo(statement.token, "Font size attribute must be an integer")
				}

				cs.block.Style.Size = size

				break
			case uint8:
				cs.block.Style.Size = value

				break
			default:
				return tokenErrorInfo(statement.token, "Font size attribute must be an integer")
			}
		default:
			return tokenErrorInfo(statement.token, "Unrecognized attribute")
		}

		break
	case MacroAssignment:
		macroDefinition := statement.data.(MacroStatement)
		cs.macros[macroDefinition.name] = macroDefinition.statements

		break
	case MacroCall:
		macroInvocation := statement.data.(MacroInvocation)
		macro, ok := cs.macros[macroInvocation.reference]
		if !ok {
			return tokenErrorInfo(statement.token, "Macro not defined")
		}

		for _, statement := range macro {
			if err := cs.processStatement(statement); err != nil {
				return err
			}
		}
	}

	return nil
}

func (cs *compilationState) finalizeCompilation() types.Show {
	cs.slide.Blocks = append(cs.slide.Blocks, cs.block)
	cs.show.Slides = append(cs.show.Slides, cs.slide)
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
	return types.Left, tokenErrorInfo(token, message)
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
			return nil, tokenErrorInfo(token, message)
		}
	case ColorLiteral:
		return color.RGBA{
			R: value.r,
			G: value.g,
			B: value.b,
			A: value.a,
		}, nil
	default:
		return nil, tokenErrorInfo(token, "Color attribute must be either a tuple or string")
	}
}
