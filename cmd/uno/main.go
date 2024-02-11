package main

import (
	"fmt"
	"github.com/AvikantSrivastava/uno/models"
	"github.com/AvikantSrivastava/uno/models/constants/rank"
)

func main() {
	fmt.Println("welcome to the UNO game")
	card := models.Card{
		Rank: rank.WILD,
		//Color: color.BLUE,
	}
	fmt.Println(card.Type())
}
