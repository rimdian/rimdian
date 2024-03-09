package common

import "github.com/rotisserie/eris"

var TagsColors = []string{
	"default",
	"magenta",
	"red",
	"volcano",
	"orange",
	"gold",
	"lime",
	"green",
	"cyan",
	"blue",
	"geekblue",
	"purple",
}

var (
	ErrTagColorInvalid = eris.New("color is not valid")
)

func IsTagColorValid(color string) bool {
	for _, c := range TagsColors {
		if c == color {
			return true
		}
	}
	return false
}

func ValidateColor(color string) error {
	if !IsTagColorValid(color) {
		return ErrTagColorInvalid
	}
	return nil
}
