package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var ErrWorkingFileNotFound = errors.New("查無檔案")

func backupEnvFiles(dirPath string) error {
	homeDir, _ := os.UserHomeDir()
	backupDir := homeDir + "/Desktop/bak_illu"

	err := filepath.Walk(dirPath, func(src string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == ".env" {
			envFile := strings.TrimPrefix(src, dirPath+"/") // api/cloudflare/.env
			source, err := os.Open(dirPath + "/" + envFile) // $HOME/proj/illu/api/cloudflare/.env
			if err != nil {
				return err
			}
			defer source.Close()

			dest := backupDir + "/" + envFile
			fmt.Println("開始備份 " + src + " ;要備份到 " + dest)
			writeBackup(src, dest)

		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func writeBackup(work, backup string) error {
	if err := os.MkdirAll(filepath.Dir(backup), 0755); err != nil {
		return err
	}

	workFile, err := os.Open(work)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrWorkingFileNotFound
		}
		return err
	}
	defer workFile.Close()

	backFile, err := os.OpenFile(backup, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer backFile.Close()

	content, err := io.ReadAll(workFile)
	if err != nil {
		return err
	}

	backFile.WriteString(string(content))
	if err != nil {
		return err
	}
	return nil
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("*** Please provide the directory path as an argument. ***")
		os.Exit(1)
	}

	dirPath := args[0]
	err := backupEnvFiles(dirPath)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Bye 9487")
	}
}
