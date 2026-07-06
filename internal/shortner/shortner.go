package shortner

import (
	"github.com/jaevor/go-nanoid"
)

func GenerateCode() (string, error) {
	generator, err := nanoid.Standard(8)
	if err != nil {
		return "error generating shortCode", err
	}
	return generator(), nil
}
