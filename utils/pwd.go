package utils

import (
	"os"
)

func GetCurDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}
