package log

import (
	"fmt"

	"github.com/logrusorgru/aurora"
)

const (
	Debug = 3
	Info  = 2
	Warn  = 1
	Error = 0
)

var Level = Debug

func SetLogLevel(logLevelName string) {
	switch logLevelName {
	case "debug":
		Level = Debug
		break
	case "info":
		Level = Info
		break
	case "warn":
		Level = Warn
		break
	case "error":
		Level = Error
		break
	default:
		Level = Debug
	}
}

// Colorize change the logger to support colors printing.
func Colorize() {
	au = aurora.NewAurora(true)
}

// internal colorized
var au aurora.Aurora

// Au Aurora instance used for colors
func Au() aurora.Aurora {
	if au == nil {
		au = aurora.NewAurora(false)
	}
	return au
}

// Printf print a message with formatting
func Printf(part string, parts ...interface{}) {
	if Level >= Debug {
		managementPrint()
		fmt.Println(fmt.Sprintf(part, parts...))
	}
}

// Errorf print a error with formatting (red)
func Errorf(part string, parts ...interface{}) {
	if Level >= Error {
		managementPrint()
		fmt.Println(Au().Colorize(fmt.Sprintf(fmt.Sprintf("%v", part), parts...), aurora.RedFg).String())
	}
}

// Warnf print a warning with formatting (yellow)
func Warnf(part string, parts ...interface{}) {
	if Level >= Warn {
		managementPrint()
		fmt.Println(Au().Colorize(fmt.Sprintf(fmt.Sprintf("%v", part), parts...), aurora.YellowFg).String())
	}
}

// Infof print a information with formatting (green)
func Infof(part string, parts ...interface{}) {
	if Level >= Info {
		managementPrint()
		fmt.Println(Au().Colorize(fmt.Sprintf(fmt.Sprintf("%v", part), parts...), aurora.GreenFg).String())
	}
}

// Infof print a information with formatting (green)
func Debugf(part string, parts ...interface{}) {
	if Level >= Debug {
		managementPrint()
		fmt.Println(Au().Colorize(fmt.Sprintf(fmt.Sprintf("%v", part), parts...), aurora.GreenFg).String())
	}
}

func managementPrint() {
	fmt.Print(Au().Bold(Au().Cyan("management: ")).String())
}
