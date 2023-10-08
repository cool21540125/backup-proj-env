package main

import (
	"archive/zip"
	"errors"
	"flag"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

var ErrWorkingFileNotFound = errors.New("no such file")

func archiveBackup(archiveSrcDir, target string) error {
	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	filepath.Walk(archiveSrcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.Join(filepath.Base(archiveSrcDir), filepath.ToSlash(path[len(archiveSrcDir):]))

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

	// fmt.Println(base)
	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if slices.Contains([]string{"node_modules", "venv", ".git", "bin", "obj", ".vscode", "dist", "build"}, d.Name()) {
				return filepath.SkipDir
			}

			srcEnvFile := path + "/.env"
			if _, err := os.Stat(srcEnvFile); err == nil {
				subDir := strings.TrimPrefix(path, srcDir+"/")
				source, err := os.Open(srcEnvFile)
				if err != nil {
					return err
				}
				defer source.Close()

				destEnvFile := destDir + "/" + subDir + "/.env"

				// fmt.Println(srcEnvFile + " -> " + destEnvFile)
				makeBackup(srcEnvFile, destEnvFile)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func makeBackup(srcEnvFile, destEnvFile string) error {
	if err := os.MkdirAll(filepath.Dir(destEnvFile), 0755); err != nil {
		return err
	}

	backupSrcEnv, err := os.Open(srcEnvFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrWorkingFileNotFound
		}
		return err
	}
	defer backupSrcEnv.Close()

	backupDestEnv, err := os.OpenFile(destEnvFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer backupDestEnv.Close()

	content, err := io.ReadAll(backupSrcEnv)
	if err != nil {
		return err
	}

	backupDestEnv.WriteString(string(content))
	if err != nil {
		return err
	}
	return nil
}

func processDirPath(srcDir string) string {
	home := os.Getenv("HOME")
	return strings.ReplaceAll(srcDir, "$HOME", home)
}

func main() {
	bakDir := flag.String("src", ".", "source directory")
	outDir := flag.String("out", ".", "output directory")
	flag.Parse()
	// fmt.Println("備份路徑: " + *bakDir)
	// fmt.Println("產出路徑: " + *outDir)

	srcDir := *bakDir
	srcDir = processDirPath(srcDir)
	destDir := *outDir
	destDir = processDirPath(destDir + "/" + filepath.Base(srcDir))

	err := backupEnvFiles(srcDir, destDir)
	if err != nil {
		log.Fatal(err)
	}

	dt := time.Now()
	err = archiveBackup(destDir, destDir+".env."+dt.Format("20060102")+".zip")
	if err != nil {
		log.Fatal(err)
	}
	os.RemoveAll(destDir)
}
