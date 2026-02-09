package linkworktrees

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindLinkFiles(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create test directory structure
	dirs := []string{
		filepath.Join(tmpDir, ".claude"),
		filepath.Join(tmpDir, "subdir"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	// Create test files
	files := []string{
		filepath.Join(tmpDir, "file1.txt"),
		filepath.Join(tmpDir, "file2.txt"),
		filepath.Join(tmpDir, ".hidden"),
		filepath.Join(tmpDir, ".claude", "settings.json"),
		filepath.Join(tmpDir, "subdir", "nested.txt"),
	}
	for _, file := range files {
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", file, err)
		}
	}

	tests := []struct {
		name        string
		linksDir    string
		excludeDirs []string
		wantCount   int
		wantFiles   []string
	}{
		{
			name:        "exclude .claude directory",
			linksDir:    tmpDir,
			excludeDirs: []string{".claude"},
			wantCount:   3, // file1.txt, file2.txt, subdir/nested.txt (excludes .hidden and .claude/*)
			wantFiles: []string{
				filepath.Join(tmpDir, "file1.txt"),
				filepath.Join(tmpDir, "file2.txt"),
				filepath.Join(tmpDir, "subdir", "nested.txt"),
			},
		},
		{
			name:        "no exclusions",
			linksDir:    tmpDir,
			excludeDirs: []string{},
			wantCount:   4, // file1.txt, file2.txt, subdir/nested.txt, .claude/settings.json (excludes .hidden)
		},
		{
			name:        "exclude multiple directories",
			linksDir:    tmpDir,
			excludeDirs: []string{".claude", "subdir"},
			wantCount:   2, // file1.txt, file2.txt
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findLinkFiles(tt.linksDir, tt.excludeDirs)
			if err != nil {
				t.Fatalf("findLinkFiles() error = %v", err)
			}

			if len(got) != tt.wantCount {
				t.Errorf("findLinkFiles() returned %d files, want %d", len(got), tt.wantCount)
				t.Logf("Got files: %v", got)
			}

			if tt.wantFiles != nil {
				gotSet := make(map[string]bool)
				for _, f := range got {
					gotSet[f] = true
				}
				for _, want := range tt.wantFiles {
					if !gotSet[want] {
						t.Errorf("findLinkFiles() missing expected file: %s", want)
					}
				}
			}
		})
	}
}

func TestFindLinkFilesEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	got, err := findLinkFiles(tmpDir, nil)
	if err != nil {
		t.Fatalf("findLinkFiles() error = %v", err)
	}

	if len(got) != 0 {
		t.Errorf("findLinkFiles() returned %d files for empty dir, want 0", len(got))
	}
}

func TestFindLinkFilesNonExistentDir(t *testing.T) {
	_, err := findLinkFiles("/non/existent/path", nil)
	if err == nil {
		t.Error("findLinkFiles() expected error for non-existent directory")
	}
}
