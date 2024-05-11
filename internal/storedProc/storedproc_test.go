package storedProc_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
)

func TestMain(m *testing.M) {
	setup()

	e := m.Run()

	tearDown()

	os.Exit(e)

}

func setup() {
	fmt.Println("setup")
}

func tearDown() {
	fmt.Println("reardown")
}

func TestSlug(t *testing.T) {

	// Arrange
	sp := &storedProc.StoredProc{}

	// Act
	slug := sp.Slug()

	// Assert
	if slug != "" {
		t.Errorf("Got: %s, Want: %s", slug, "*BLANK")
	}

}

func TestSlug2(t *testing.T) {

	// Arrange
	sp := &storedProc.StoredProc{}

	// Act
	slug := sp.Slug()

	// Assert
	if slug != "" {
		t.Errorf("Got: %s, Want: %s", slug, "*BLANK")
	}

}
