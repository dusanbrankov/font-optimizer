package main

import (
	"archive/zip"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidMediaType = errors.New("malformed media type")
	ErrUnknownFileType  = errors.New("cannot determine file type")
	ErrNotExist         = errors.New("file doesn't exist")
)

func saveToDisc(file multipart.File, base string, ext string, destDir string) (string, error) {
	dst, err := os.Create(filepath.Join(destDir, fmt.Sprintf("%s.%s", base, ext)))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(dst.Name())
		return "", err
	}

	return dst.Name(), nil
}

func getMediaType(data []byte) (string, error) {
	return http.DetectContentType(data), nil
}

// hasExt reports whether filename ends with one of the provided extensions
// (case-insensitive). This is a fast, user-friendly pre-check and should not
// be relied on for security.
func hasExt(filename string, exts ...string) bool {
	maxExtLen := 0
	for _, ext := range exts {
		if len(ext) > maxExtLen {
			maxExtLen = len(ext)
		}
	}
	n := maxExtLen + 1 // add 1 to include dot
	if utf8.RuneCountInString(filename) < n {
		return false
	}
	suffix := filename[len(filename)-n:]
	suffix = strings.ToLower(suffix)
	for _, ext := range exts {
		fmt.Printf("%s == .%s\n", suffix, ext)
		if suffix == "."+ext {
			return true
		}
	}
	return false
}

func fileExt(mediaType string) (string, error) {
	parts := strings.Split(mediaType, "/")

	if len(parts) != 2 {
		return "", ErrInvalidMediaType
	}

	return parts[1], nil
}

func clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func serverError(w http.ResponseWriter, err error) {
	log.Print(err)
	status := http.StatusInternalServerError
	http.Error(w, http.StatusText(status), status)
}

func decodePostForm(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	return r.ParseMultipartForm(maxMemory)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func createZip(dest string, files ...string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		w, err := zw.Create(filepath.Join("font-optimizer", filepath.Base(file)))
		if err != nil {
			return err
		}

		if _, err := io.Copy(w, f); err != nil {
			return err
		}
	}

	return nil
}

func randStr(l int) string {
	bytes := make([]byte, l)
	rand.Read(bytes)
	// return without padding and lowercase
	return strings.ToLower(base32.StdEncoding.EncodeToString(bytes)[:l])
}
