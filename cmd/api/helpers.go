package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
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

func getContentType(file multipart.File) (string, error) {
	sniff := make([]byte, 512)
	n, err := file.Read(sniff)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	return http.DetectContentType(sniff[:n]), nil
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

const (
	_ = iota
	isRegular
	isDir
	isSymbolic
)

func fileType(name string) (int, error) {
	fi, err := os.Lstat(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, ErrNotExist
		}
		return 0, err
	}

	switch mode := fi.Mode(); {
	case mode.IsRegular():
		return isRegular, nil
	case mode.IsDir():
		return isDir, nil
	case mode&fs.ModeSymlink != 0:
		return isSymbolic, nil
	}

	return 0, ErrUnknownFileType
}
