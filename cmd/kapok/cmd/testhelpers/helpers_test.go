package testhelpers

import (
	"testing"
)

func TestCreateTempDir(t *testing.T) {
	dir := CreateTempDir(t)
	
	if dir == "" {
		t.Fatal("CreateTempDir returned empty string")
	}
	
	if !DirExists(dir) {
		t.Fatalf("Temp directory does not exist: %s", dir)
	}
}

func TestWriteTestFile(t *testing.T) {
	dir := CreateTempDir(t)
	content := "test content"
	
	path := WriteTestFile(t, dir, "test.txt", content)
	
	if !FileExists(path) {
		t.Fatalf("Test file was not created: %s", path)
	}
}

func TestFileExists(t *testing.T) {
	dir := CreateTempDir(t)
	
	// File doesn't exist
	if FileExists(dir + "/nonexistent.txt") {
		t.Error("FileExists returned true for non-existent file")
	}
	
	// File exists
	path := WriteTestFile(t, dir, "exists.txt", "content")
	if !FileExists(path) {
		t.Error("FileExists returned false for existing file")
	}
}

func TestDirExists(t *testing.T) {
	dir := CreateTempDir(t)
	
	// Directory exists
	if !DirExists(dir) {
		t.Error("DirExists returned false for existing directory")
	}
	
	// Directory doesn't exist
	if DirExists(dir + "/nonexistent") {
		t.Error("DirExists returned true for non-existent directory")
	}
}
