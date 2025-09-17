package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/image/font/sfnt"
)

// A full list of file signatures can be found at
// https://en.wikipedia.org/wiki/List_of_file_signatures
//
// var (
// 	signatureTTF  = []byte{0x00, 0x01, 0x00, 0x00, 0x00}
// 	signatureWOFF = []byte{0x77, 0x4F, 0x46, 0x46}
// 	signatueWOFF2 = []byte{0x77, 0x4F, 0x46, 0x32}
// )

var filenameRX = regexp.MustCompile(`^[0-9A-Za-z_\-\s]+$`)

var subsets = map[string][]int{
	"basic-latin":        {0x20, 0x7f},
	"latin-1-supplement": {0xA0, 0xFF},
}

func fileUploadHandler(w http.ResponseWriter, r *http.Request) {
	if err := decodePostForm(w, r); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			clientError(w, http.StatusRequestEntityTooLarge)
		} else {
			clientError(w, http.StatusBadRequest)
		}
		return
	}

	file, _, err := r.FormFile("font")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			clientError(w, http.StatusBadRequest)
		} else {
			clientError(w, http.StatusBadRequest)
		}
		return
	}
	defer file.Close()

	mediaType, err := getContentType(file)
	if err != nil {
		serverError(w, err)
		return
	}

	switch mediaType {
	// allowed media types
	// ref: https://www.iana.org/assignments/media-types/media-types.xhtml#font
	case "font/ttf", "font/woff", "font/woff2":
	default:
		clientError(w, http.StatusUnsupportedMediaType)
		return
	}

	// Not required since http.DetectContentType already checks for
	// the magic bytes.
	// Can be used for additional validation if needed or to support
	// custom error messages, such as:
	// "file signature invalid" vs. "file type invalid"
	//
	// if !bytes.HasPrefix(sniff, signatureTTF) {
	// 	clientError(w, http.StatusUnsupportedMediaType)
	// 	return
	// }

	ext, err := fileExt(mediaType)
	if err != nil {
		if errors.Is(err, ErrInvalidMediaType) {
			clientError(w, http.StatusUnsupportedMediaType)
		} else {
			serverError(w, err)
		}
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		serverError(w, err)
		return
	}

	font, err := sfnt.Parse(data)
	if err != nil {
		serverError(w, err)
		return
	}

	var buf sfnt.Buffer

	fontFamily, err := font.Name(&buf, sfnt.NameIDTypographicFamily)
	if err != nil {
		if errors.Is(err, sfnt.ErrNotFound) {
			fontFamily = "Unknown"
		} else {
			serverError(w, err)
			return
		}
	}

	fontName, err := font.Name(&buf, sfnt.NameIDPostScript)
	if err != nil {
		serverError(w, err)
		return
	}

	if ok := filenameRX.Match([]byte(fontFamily)); !ok {
		clientError(w, http.StatusUnprocessableEntity)
		return
	}
	if ok := filenameRX.Match([]byte(fontName)); !ok {
		clientError(w, http.StatusUnprocessableEntity)
		return
	}

	fontParentDir := strings.ReplaceAll(fontFamily, " ", "-")
	destDir := filepath.Join(uploadDir, fontParentDir)

	_, err = fileType(destDir)
	if err != nil {
		if errors.Is(err, ErrNotExist) {
			os.Mkdir(destDir, os.ModePerm)
		} else {
			serverError(w, err)
			return
		}
	}

	_, err = saveToDisc(file, fontName, ext, destDir)
	if err != nil {
		serverError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
