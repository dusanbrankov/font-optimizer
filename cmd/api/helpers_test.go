package main

import (
	"strings"
	"testing"
)

func TestHasExt(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		exts     []string
		want     bool
	}{
		{"kebab-case", "pexels-kpaukshtite-3270223.jpg", []string{"jpg"}, true},
		{"empty", "", []string{"jpg"}, false},
		{"short", "a", []string{"jpg"}, false},
		{"dot only", ".", []string{"jpg"}, false},
		{"hidden file", ".jpg", []string{"jpg"}, true},
		{"unicode", "файл.JPG", []string{"jpg"}, true},
		{"unicode long", strings.Repeat("s^M`$/=B1쎨K.t{ROfv3t ", 100) + ".jpg", []string{"jpg"}, true},
		{"long", strings.Repeat("a", 10000) + ".jpg", []string{"jpg"}, true},
	}
	for _, tc := range cases {
		got := hasExt(tc.filename, tc.exts...)
		if got != tc.want {
			t.Errorf("%s: got %t; want %t", tc.name, got, tc.want)
		}
	}
}
