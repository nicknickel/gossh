package menus

import (
	"testing"
)

func TestStyleTitle(t *testing.T) {
	text := "some text"
	expect := " some text "

	result := StyleTitle(text)
	if result != expect {
		t.Errorf("TestStyleTitle() want %q, got %q", expect, result)
	}
}

func TestGlobalStyle(t *testing.T) {
	text := "some text"
	expect := "             \n  some text  \n             "

	result := globalStyle(text)
	if result != expect {
		t.Errorf("TestGlobalStyle() want %q, got %q", expect, result)
	}
}
