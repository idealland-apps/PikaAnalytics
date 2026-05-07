package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetDataPath() string {
	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		dataPath = "./data"
	}
	return dataPath
}

func GetMainDbPath() string {
	return filepath.Join(GetDataPath(), "main.db")
}

func GetVisitsDir() string {
	return filepath.Join(GetDataPath(), "visits")
}

func GetVisitsDbPath(year int, month int) string {
	return filepath.Join(GetVisitsDir(), fmt.Sprintf("visits_%04d_%02d.db", year, month))
}
