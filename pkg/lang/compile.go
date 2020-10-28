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
	show := types.NewShow()
	variables := make(map[string]interface{})

	var slide = types.NewSlide()
	var block = types.NewBlock()

	for _, statement := range statements {
		switch statement.Type {
		case SlideDecl:
			slide.Blocks = append(slide.Blocks, block)
			show.Slides = append(show.Slides, slide)

			block = types.NewBlock()
			slide = types.Slide{
				Background: slide.Background,
				Blocks:     make([]types.Block, 0),
			}

			break
		case ScopeDecl:
			slide.Blocks = append(slide.Blocks, block)

			// Copy Style from previous block to make things
			// less tedious when writing multiple blocks
			block = types.Block{Style: block.Style}

			break
		case WordBlock:
			block.Words = statement.data.(string)

			break
		case VariableAssignment:
			variable := statement.data.(VariableStatement)

			switch value := variable.value.(type) {
			case uint8, string, ColorLiteral:
				variables[variable.name] = value

				break
			case VariableReference:
				variables[variable.name] = variables[value.reference]
			}

			break
		case AttributeAssignment:
			attribute := statement.data.(AttributeStatement)

			switch attribute.name {
			case "backgroundColor":
				switch value := attribute.value.(type) {
				case VariableReference:
					c, err := colorFromLiteral(statement.token, variables[value.reference])
					if err != nil {
						return show, err
					}

					slide.Background = c

					break
				default:
					c, err := colorFromLiteral(statement.token, value)
					if err != nil {
						return show, err
					}

					slide.Background = c
				}

				break
			case "justify":
				switch value := attribute.value.(type) {
				case VariableReference:
					justification, err := justificationFromLiteral(statement.token, variables[value.reference])
					if err != nil {
						return show, err
					}

					block.Style.Justification = justification

					break
				case string:
					justification, err := justificationFromLiteral(statement.token, value)
					if err != nil {
						return show, err
					}

					block.Style.Justification = justification
				}

				break
			case "font":
				switch value := attribute.value.(type) {
				case VariableReference:
					val := variables[value.reference]
					switch val := val.(type) {
					case string:
						block.Style.Font = val
					default:
						return show, tokenErrorInfo(statement.token, "Font attribute must be a string")
					}

					break
				case string:
					block.Style.Font = value
				}

				break
			case "fontColor":
				switch value := attribute.value.(type) {
				case VariableReference:
					c, err := colorFromLiteral(statement.token, variables[value.reference])
					if err != nil {
						return show, err
					}

					block.Style.Color = c

					break
				default:
					c, err := colorFromLiteral(statement.token, value)
					if err != nil {
						return show, err
					}

					block.Style.Color = c
				}

				break
			case "fontSize":
				switch value := attribute.value.(type) {
				case VariableReference:
					size, ok := variables[value.reference].(uint8)
					if !ok {
						return show, tokenErrorInfo(statement.token, "Font size attribute must be an integer")
					}

					block.Style.Size = size

					break
				case uint8:
					block.Style.Size = value

					break
				default:
					return show, tokenErrorInfo(statement.token, "Font size attribute must be an integer")
				}
			}
		}
	}

	slide.Blocks = append(slide.Blocks, block)
	show.Slides = append(show.Slides, slide)

	return show, nil
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
