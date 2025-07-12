package veracity

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/datatrails/go-datatrails-merklelog/massifs/watcher"
)

func filePathToLogMassifs(filePath string) ([]watcher.LogMassif, error) {
	filePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return scannerToLogMassifs(bufio.NewScanner(f))
}

func stdinToDecodedLogMassifs() ([]watcher.LogMassif, error) {
	return scannerToLogMassifs(bufio.NewScanner(os.Stdin))
}

func scannerToLogMassifs(scanner *bufio.Scanner) ([]watcher.LogMassif, error) {
	var data []byte
	for scanner.Scan() {
		data = append(data, scanner.Bytes()...)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return watcher.LogMassifsFromData(data)
}
