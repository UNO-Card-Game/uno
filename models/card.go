package models

import (
	"github.com/AvikantSrivastava/uno/models/constants/color"
	"github.com/AvikantSrivastava/uno/models/constants/rank"
)

type Card struct {
	Rank  rank.Rank   `validate:"required"`
	Color color.Color `validate:"omitempty"`
}

func (c Card) Type() string {
	if c.Rank >= "0" && c.Rank <= "9" {
		return "number-card"
	} else if c.Rank == rank.WILD || c.Rank == rank.DRAW_4 {
		return "action-card-no-constants"
	}
	return "action-card"
}

func (c Card) logCard() string {
	return string(c.Rank) + " " + string(c.Color)
}
