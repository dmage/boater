package printer

import (
	"fmt"

	"github.com/fatih/color"
)

func Key(key string) {
	color.New(color.FgHiBlue).Print(key)
}

func Value(value interface{}) {
	color.New(color.FgHiGreen).Print(value)
}

func Referencef(format string, a ...interface{}) {
	color.New(color.FgHiYellow).Printf(format, a...)
}

func Delim(delim string) {
	color.New(color.FgHiWhite).Print(delim)
}

func Keyln(prefix string, key string) {
	Delim(prefix)
	Key(key)
	Delim(":")
	fmt.Print("\n")
}

func Valueln(prefix string, value interface{}) {
	Delim(prefix)
	Value(value)
	fmt.Print("\n")
}

func KeyValueln(prefix string, key string, value interface{}) {
	Delim(prefix)
	Key(key)
	Delim(": ")
	Value(value)
	fmt.Print("\n")
}
