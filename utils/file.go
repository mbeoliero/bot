package utils

import (
	"errors"
	"os"

	"github.com/sirupsen/logrus"
)

func DelFile(path string) {
	err := os.Remove(path)
	if err != nil {
		logrus.Warnf("failed to delete file, err %s", err)
		return
	}
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || errors.Is(err, os.ErrExist)
}

func ReadFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		logrus.Errorf("failed to read file, err %s", err)
		return ""
	}
	return string(b)
}

func WriteFile(filepath string, data []byte) {
	err := os.WriteFile(filepath, data, 0o644)
	if err != nil {
		logrus.Errorf("failed to write file, err %s", err)
	}
}
