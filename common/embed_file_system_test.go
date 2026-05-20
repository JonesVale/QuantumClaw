package common

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"testing"
)

//go:embed embed-file-system.go
var testFS embed.FS

func TestEmbedFolder_WithValidFS(t *testing.T) {
	// Using the test file itself as the embed source
	efs := EmbedFolder(testFS, ".")
	if efs == nil {
		t.Fatal("EmbedFolder returned nil for valid embed.FS")
	}
}

func TestEmbedFolder_WithInvalidPath(t *testing.T) {
	// Should return a graceful failFS instead of panicking
	efs := EmbedFolder(testFS, "nonexistent-path")
	if efs == nil {
		t.Fatal("EmbedFolder returned nil, expected graceful fallback")
	}

	// Opening a file in the failFS should return ErrNotExist
	_, err := efs.Open("any-file")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Expected fs.ErrNotExist, got %v", err)
	}
}

func TestEmbedFolder_Exists(t *testing.T) {
	// Validate behavior
	efs := EmbedFolder(testFS, ".")

	// This file exists in the embedded FS
	if !efs.Exists("", "embed-file-system.go") {
		t.Error("Exists() returned false for an existing file")
	}

	// This file doesn't exist
	if efs.Exists("", "nonexistent-file.go") {
		t.Error("Exists() returned true for a nonexistent file")
	}
}

func TestFailFS_Open(t *testing.T) {
	// Direct test of the failFS type
	ffs := failFS{}

	_, err := ffs.Open("anything")
	if err != fs.ErrNotExist {
		t.Errorf("failFS.Open() returned %v, want fs.ErrNotExist", err)
	}
}

func TestEmbedFolder_GracefulFallback(t *testing.T) {
	// Simulate a corrupted/missing dist directory
	efs := EmbedFolder(testFS, "web/default/dist")

	// The directory 'web/default/dist' doesn't exist in our embed
	// The function should return gracefully
	if efs == nil {
		t.Fatal("EmbedFolder should never return nil")
	}

	// Opening should return ErrNotExist (graceful degradation)
	_, err := efs.Open(".")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Expected fs.ErrNotExist for missing dist, got %v", err)
	}
}

func TestEmbedFileSystem_ExistsOnEmpty(t *testing.T) {
	// Empty embedFileSystem on failFS
	efs := embedFileSystem{
		FileSystem: http.FS(failFS{}),
	}

	if efs.Exists("", "anything") {
		t.Error("Exists() should return false for failFS")
	}
}
