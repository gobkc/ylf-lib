package ylf

import "github.com/satori/go.uuid"

func Uuid() string {
	return uuid.NewV4().String()
}