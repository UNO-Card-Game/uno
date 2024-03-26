package color

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
