package esbuild

import (
	"fmt"
	"testing"
)

// Test if BuildApp() works
func TestBuildApp(t *testing.T) {
	if err := BuildApp(); err != nil {
		fmt.Println("error:", err.Error())
	}
}
