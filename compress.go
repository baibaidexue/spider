package main

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
)

func Zip(srcFile string, destZip string) error {
	zipfile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.Base(path)
		if info.IsDir() {
			return nil
			// header.Name += "/"
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
		}
		return err
	})

	return err
}

func exportMangaData(srcFile, dstFile string) {
	log.Println("input:", srcFile)
	log.Println("output:", dstFile)
	if ok, _ := PathExists(srcFile); ok {
		fileCopy(srcFile, dstFile)
		return
	}
	mangaDir := filepath.Dir(srcFile)
	imageDir := filepath.Join(mangaDir, DIRIMAGES)
	err := Zip(imageDir, dstFile)
	if err != nil {
		log.Println(err)
	}
}
