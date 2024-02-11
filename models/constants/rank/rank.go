package rank

type Rank string

const (
	ZERO    Rank = "0"
	ONE     Rank = "1"
	TWO     Rank = "2"
	THREE   Rank = "3"
	FOUR    Rank = "4"
	FIVE    Rank = "5"
	SIX     Rank = "6"
	SEVEN   Rank = "7"
	EIGHT   Rank = "8"
	NINE    Rank = "9"
	WILD    Rank = "wild"
	DRAW_4  Rank = "draw_4"
	DRAW_2  Rank = "draw_2"
	REVERSE Rank = "reverse"
	SKIP    Rank = "skip"
)

var NumberCards = []Rank{ZERO, ONE, TWO, THREE, FOUR, FIVE, SIX, SEVEN, EIGHT, NINE}

var ActionCards = []Rank{DRAW_2, REVERSE, SKIP}

var ActionCardsNoColor = []Rank{WILD, DRAW_4}
