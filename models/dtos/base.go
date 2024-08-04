package dtos

import "encoding/json"

type DTO struct {
	Type string `json:"type"`
	Obj  ObjDTO `json:"obj"`
}

type ObjDTO interface {
	Serialize() []byte
}

func Serialize(obj ObjDTO, type_ string) []byte {
	dto := DTO{
		Type: type_,
		Obj:  obj,
	}
	data, err := json.Marshal(dto)
	if err != nil {
		return nil
	}
	return data
}
