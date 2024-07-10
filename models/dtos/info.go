package dtos

type InfoDTO struct {
	Message string
}

func (dto InfoDTO) Serialize() []byte {
	return Serialize(
		dto, "info")
}
