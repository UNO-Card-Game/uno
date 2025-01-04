package commands

type PlayCardCommand struct {
	CardIndex int `json:"card_index"`
	NewColor  string `json:"new_color"`
}
