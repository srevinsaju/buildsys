package ui

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"os"
)

func NewColor(value ...color.Attribute) *color.Color {
	c := color.New(value...)
	c.EnableColor()
	return c
}

var Green = NewColor(color.FgGreen).SprintFunc()
var Red = NewColor(color.FgRed).SprintFunc()
var Bold = NewColor(color.Bold).SprintFunc()
var Blue = NewColor(color.FgBlue).SprintFunc()
var Grey = NewColor(color.FgHiBlack).SprintFunc()
var Cyan = NewColor(color.FgCyan).SprintFunc()
var White = NewColor(color.FgWhite).SprintFunc()
var Yellow = NewColor(color.FgYellow).SprintFunc()
var Magenta = NewColor(color.FgMagenta).SprintFunc()
var Orange = NewColor(color.FgHiYellow).SprintFunc()
var HiGreen = NewColor(color.FgHiGreen).SprintFunc()
var HiBlue = NewColor(color.FgHiBlue).SprintFunc()
var HiMagenta = NewColor(color.FgHiMagenta).SprintFunc()
var HiBlack = NewColor(color.FgHiBlack).SprintFunc()
var HiCyan = NewColor(color.FgHiCyan).SprintFunc()
var HiWhite = NewColor(color.FgHiWhite).SprintFunc()
var HiYellow = NewColor(color.FgHiYellow).SprintFunc()
var HiRed = NewColor(color.FgHiRed).SprintFunc()
var Italic = NewColor(color.Italic).SprintFunc()
var Plus = NewColor(color.FgHiWhite).SprintFunc()("+")
var SubStage = Grey("==>")
var SubSubStage = Grey("-->")
var Matrix = Blue("matrix")
var Stage = Blue("stage")
var Options = Grey("options")
var FailLazy = Grey("fail-lazy")

var True = Green("true")
var False = Red("false")

func Error(message string) {
	fmt.Println(Red(message))
	os.Exit(1)
}

func Success(message string, args ...interface{}) {
	fmt.Println(Green(fmt.Sprintf(message, args...)))
}

func DeprecationWarning(message string, args ...string) {
	fmt.Println(HiYellow("[deprecated] "), message, args)
}

var AnsiFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "color",
			Type: cty.String,
		},
		{
			Name: "message",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		color := args[0].AsString()
		message := args[1].AsString()
		return cty.StringVal(Color(color, message)), nil
	},
})

func Color(c string, message string) string {
	switch c {
	case "error":
		return Red(Bold(message))
	case "success":
		return Green(Bold(message))
	case "warn":
		return Yellow(Bold(message))
	case "warning":
		return Yellow(Bold(message))
	case "info":
		return Blue(Bold(message))
	case "green":
		return Green(message)
	case "red":
		return Red(message)
	case "blue":
		return Blue(message)
	case "yellow":
		return Yellow(message)
	case "bold":
		return Bold(message)
	case "italic":
		return Italic(message)
	case "cyan":
		return Cyan(message)
	case "grey":
		return Grey(message)
	case "white":
		return White(message)
	case "magenta":
		return Magenta(message)
	case "orange":
		return Orange(message)
	case "hi-green":
		return HiGreen(message)
	case "hi-blue":
		return HiBlue(message)
	case "hi-magenta":
		return HiMagenta(message)
	case "hi-black":
		return HiBlack(message)
	case "hi-white":
		return HiWhite(message)
	case "hi-red":
		return HiRed(message)
	case "hi-cyan":
		return HiCyan(message)

	case "hi-yellow":
		return HiYellow(message)
	default:
		return message
	}
}
