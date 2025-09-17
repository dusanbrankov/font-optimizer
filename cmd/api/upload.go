package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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

	fonts := r.MultipartForm.File["font"]
	if len(fonts) > maxFiles {
		clientError(w, http.StatusRequestEntityTooLarge)
		return
	}

	selectedSubsets := r.PostForm["subsets"]
	subsettedFonts := make([]string, 0, len(fonts)*len(selectedSubsets))

	for _, mf := range fonts {
		file, err := mf.Open()
		if err != nil {
			serverError(w, err)
			return
		}

		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			serverError(w, err)
			return
		}

		mediaType, err := getMediaType(data)
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

		font, err := sfnt.Parse(data)
		if err != nil {
			serverError(w, err)
			return
		}

		var buf sfnt.Buffer

		family, err := font.Name(&buf, sfnt.NameIDFamily)
		if err != nil {
			if errors.Is(err, sfnt.ErrNotFound) {
				family = "Unknown"
			} else {
				serverError(w, err)
				return
			}
		}

		subfamily, err := font.Name(&buf, sfnt.NameIDSubfamily)
		if err != nil {
			if errors.Is(err, sfnt.ErrNotFound) {
				http.Error(w, "cannot read subfamily", http.StatusUnprocessableEntity)
			} else {
				serverError(w, err)
			}
			return
		}

		if ok := filenameRX.Match([]byte(family)); !ok {
			clientError(w, http.StatusUnprocessableEntity)
			return
		}
		if ok := filenameRX.Match([]byte(subfamily)); !ok {
			clientError(w, http.StatusUnprocessableEntity)
			return
		}

		parent := strings.ReplaceAll(family, " ", "-")
		// TODO: validate destDir against directory traversal (../)
		destDir := filepath.Join(uploadDir, parent)

		if err := os.MkdirAll(destDir, 0755); err != nil {
			serverError(w, err)
			return
		}

		var savePath, filename string

		for _, v := range selectedSubsets {
			urange, ok := subsets[v]
			if !ok {
				clientError(w, http.StatusBadRequest)
				return
			}

			base := fmt.Sprintf("%s-%s", parent, subfamily)

			filename = fmt.Sprintf("%s.%s.%s", base, v, "woff2")
			savePath = filepath.Join(destDir, filename)

			if !exists(savePath) {
				if err := subsetFont(data, savePath, urange); err != nil {
					os.Remove(savePath)
					for _, f := range subsettedFonts {
						os.Remove(f)
					}
					serverError(w, err)
					return
				}
			}

			subsettedFonts = append(subsettedFonts, savePath)
		}

		w.Header().Set("Content-Type", "font/woff2")
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		http.ServeFile(w, r, savePath)
	}
}

func subsetFont(data []byte, dest string, unicodes []int) error {
	from := fmt.Sprintf("%04X", unicodes[0])
	to := fmt.Sprintf("%04X", unicodes[1])

	tmp, err := os.CreateTemp("", "font-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		return err
	}
	tmp.Close()

	cmd := exec.Command(
		"pyftsubset",
		tmp.Name(),
		fmt.Sprintf("--unicodes=U+%s-%s", from, to),
		"--flavor=woff2",
		"--output-file="+dest,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pyftsubset: %w\n%s", err, string(out))
	}

	return nil
}
