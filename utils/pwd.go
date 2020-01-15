package utils

import (
	"go12306/helpers/logger"
	"os"
)

func GetCurDir() string {
	dir, err := os.Getwd()
	if err != nil {
		logger.Error("Getwd err", err)
		return ""
	}
	return dir
}
