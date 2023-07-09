package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrWorkingFileNotFound = errors.New("no such file")

func archiveBackup(source, target string) error {
	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.Join(filepath.Base(source), filepath.ToSlash(path[len(source):]))

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return nil
}

func backupEnvFiles(srcDir, destDir string) error {

	err := filepath.Walk(srcDir, func(src string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == ".env" {
			envFile := strings.TrimPrefix(src, srcDir+"/")
			source, err := os.Open(srcDir + "/" + envFile)
			if err != nil {
				return err
			}
			defer source.Close()

			dest := destDir + "/" + envFile
			fmt.Println("Backing up " + src + " -> " + dest)
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

func processSrcDir(srcDir string) string {
	home := os.Getenv("HOME")
	return strings.ReplaceAll(srcDir, "$HOME", home)
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("*** Please provide the directory path as an argument. ***")
		os.Exit(1)
	}

	srcDir := processSrcDir(args[0])
	srcBaseDir := filepath.Base(srcDir)

	HOME, _ := os.UserHomeDir()
	destDir := filepath.Join(HOME, "Desktop", srcBaseDir+"_bak")

	err := backupEnvFiles(srcDir, destDir)
	if err != nil {
		fmt.Println(err)
		fmt.Println("*** Backup failed. ***")
	}

	dt := time.Now().Format("20060709")
	fmt.Println(dt)
	err = archiveBackup(destDir, destDir+"_"+dt+".zip")
	if err != nil {
		fmt.Println(err)
	}
}
