package main

import (
	"errors"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"

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

	dest, err := saveToDisc(file, ext)
	if err != nil {
		serverError(w, err)
		return
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		log.Fatal(err)
	}

	font, err := sfnt.Parse(data)
	if err != nil {
		serverError(w, err)
		return
	}

	var buf sfnt.Buffer

	nameID := func(name sfnt.NameID) string {
		n, _ := font.Name(&buf, name)
		return n
	}

	fmt.Println("NameIDCopyright", nameID(sfnt.NameIDCopyright))
	fmt.Println("NameIDFamily", nameID(sfnt.NameIDFamily))
	fmt.Println("NameIDSubfamily", nameID(sfnt.NameIDSubfamily))
	fmt.Println("NameIDUniqueIdentifier", nameID(sfnt.NameIDUniqueIdentifier))
	fmt.Println("NameIDFull", nameID(sfnt.NameIDFull))
	fmt.Println("NameIDVersion", nameID(sfnt.NameIDVersion))
	fmt.Println("NameIDPostScript", nameID(sfnt.NameIDPostScript))
	fmt.Println("NameIDTrademark", nameID(sfnt.NameIDTrademark))
	fmt.Println("NameIDManufacturer", nameID(sfnt.NameIDManufacturer))
	fmt.Println("NameIDDesigner", nameID(sfnt.NameIDDesigner))
	fmt.Println("NameIDDescription", nameID(sfnt.NameIDDescription))
	fmt.Println("NameIDVendorURL", nameID(sfnt.NameIDVendorURL))
	fmt.Println("NameIDDesignerURL", nameID(sfnt.NameIDDesignerURL))
	fmt.Println("NameIDLicense", nameID(sfnt.NameIDLicense))
	fmt.Println("NameIDLicenseURL", nameID(sfnt.NameIDLicenseURL))
	fmt.Println("NameIDTypographicFamily", nameID(sfnt.NameIDTypographicFamily))
	fmt.Println("NameIDTypographicSubfamily", nameID(sfnt.NameIDTypographicSubfamily))
	fmt.Println("NameIDCompatibleFull", nameID(sfnt.NameIDCompatibleFull))
	fmt.Println("NameIDSampleText", nameID(sfnt.NameIDSampleText))
	fmt.Println("NameIDPostScriptCID", nameID(sfnt.NameIDPostScriptCID))
	fmt.Println("NameIDWWSFamily", nameID(sfnt.NameIDWWSFamily))
	fmt.Println("NameIDWWSSubfamily", nameID(sfnt.NameIDWWSSubfamily))
	fmt.Println("NameIDLightBackgroundPalette", nameID(sfnt.NameIDLightBackgroundPalette))
	fmt.Println("NameIDDarkBackgroundPalette", nameID(sfnt.NameIDDarkBackgroundPalette))
	fmt.Println("NameIDVariationsPostScriptPrefix", nameID(sfnt.NameIDVariationsPostScriptPrefix))

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
