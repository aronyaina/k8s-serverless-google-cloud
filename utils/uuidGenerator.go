package utils

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func CreateUniqueString(name string) (string, error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s", name, strings.Split(uuid.String(), "-")[0]), nil
}
