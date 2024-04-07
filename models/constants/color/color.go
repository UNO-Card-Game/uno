package color

import (
	"fmt"
	"strings"
)

type Color string

const (
	RED    Color = "red"
	BLUE   Color = "blue"
	GREEN  Color = "green"
	YELLOW Color = "yellow"
)

var ALLColors = []Color{
	RED, BLUE, GREEN, YELLOW,
}

func ParseColor(colorStr string) (Color, error) {
	colorStr = strings.ToLower(colorStr)
	switch colorStr {
	case "red":
		return RED, nil
	case "blue":
		return BLUE, nil
	case "green":
		return GREEN, nil
	case "yellow":
		return YELLOW, nil
	default:
		return "", fmt.Errorf("invalid color: %s", colorStr)
	}
}
