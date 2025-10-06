package veracity

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/forestrie/go-merklelog/massifs/watcher"
)

func FilePathToLogMassifs(filePath string) ([]watcher.LogMassif, error) {
	filePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return ScannerToLogMassifs(bufio.NewScanner(f))
}

func StdinToDecodedLogMassifs() ([]watcher.LogMassif, error) {
	return ScannerToLogMassifs(bufio.NewScanner(os.Stdin))
}

func ScannerToLogMassifs(scanner *bufio.Scanner) ([]watcher.LogMassif, error) {
	var data []byte
	for scanner.Scan() {
		data = append(data, scanner.Bytes()...)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return LogMassifsFromData(data)
}

func LogMassifsFromData(data []byte) ([]watcher.LogMassif, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var doc []watcher.LogMassif
	err := decoder.Decode(&doc)
	if err == nil {
		return doc, nil
	}
	return nil, err
}
