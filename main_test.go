package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TimofeyChernyshev/MIREA-Configuration-management-1/vfs"
)

func TestParser(t *testing.T) {
	tests := []struct {
		name         string
		cmd          string
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name:         "command without args",
			cmd:          "ls",
			expectedCmd:  "ls",
			expectedArgs: []string{},
		},
		{
			name:         "command with 1 arg",
			cmd:          "ls arg",
			expectedCmd:  "ls",
			expectedArgs: []string{"arg"},
		},
		{
			name:         "command with 1 arg in \"",
			cmd:          "ls \"arg1 arg2\"",
			expectedCmd:  "ls",
			expectedArgs: []string{"arg1 arg2"},
		},
		{
			name:         "empty command",
			cmd:          "",
			expectedCmd:  "",
			expectedArgs: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := parser(tt.cmd)
			if cmd != tt.expectedCmd {
				t.Errorf("expected command: %v, got %v", tt.expectedCmd, cmd)
			}
			if len(args) != len(tt.expectedArgs) {
				t.Errorf("expected args: %v, got %v", tt.expectedArgs, args)
			}
		})
	}
}

func TestLsCommand(t *testing.T) {
	shell := NewShell()
	dirNode := &vfs.VFSNode{
		Name:    "testdir",
		IsDir:   true,
		ModTime: time.Now(),
		Children: []*vfs.VFSNode{
			{Name: "file1.txt", IsDir: false},
			{Name: "file2.txt", IsDir: false},
		},
	}
	shell.vfs.Root.Children = append(shell.vfs.Root.Children, dirNode)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	shell.lsCommand([]string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "testdir") {
		t.Errorf("Expected 'testdir' in output, got: %s", output)
	}

	oldStdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w

	shell.lsCommand([]string{"file1.txt"})

	w.Close()
	os.Stdout = oldStdout

	buf.Reset()
	io.Copy(&buf, r)
	output = buf.String()

	if !strings.Contains(output, "Error") {
		t.Errorf("Expected error message, got: %s", output)
	}

	oldStdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w

	shell.lsCommand([]string{"/nonexistent"})

	w.Close()
	os.Stdout = oldStdout

	buf.Reset()
	io.Copy(&buf, r)
	output = buf.String()

	if !strings.Contains(output, "Error") {
		t.Errorf("Expected error message, got: %s", output)
	}
}

func TestCdCommand(t *testing.T) {
	shell := NewShell()

	// Create test directory
	dirNode := &vfs.VFSNode{
		Name:    "test",
		IsDir:   true,
		ModTime: time.Now(),
	}
	shell.vfs.Root.Children = append(shell.vfs.Root.Children, dirNode)

	// Test cd to existing directory
	shell.cdCommand([]string{"/test"})
	if shell.currentPath != "/test" {
		t.Errorf("Expected path '/test', got '%s'", shell.currentPath)
	}

	// Test cd to non-existent directory
	originalPath := shell.currentPath
	shell.cdCommand([]string{"/nonexistent"})
	if shell.currentPath != originalPath {
		t.Error("Path should not change when cd to non-existent directory")
	}

	// Test cd to file (not directory)
	fileNode := &vfs.VFSNode{
		Name:  "file.txt",
		IsDir: false,
	}
	shell.vfs.Root.Children = append(shell.vfs.Root.Children, fileNode)
	shell.cdCommand([]string{"file.txt"})
	if shell.currentPath != originalPath {
		t.Error("Path should not change when cd to file")
	}

	// Test cd ..
	shell.cdCommand([]string{".."})
	if shell.currentPath != "/" {
		t.Errorf("Expected path '/', got '%s'", shell.currentPath)
	}
}

func TestUniqCommand(t *testing.T) {
	shell := NewShell()

	// Create test file with duplicate lines
	fileNode := &vfs.VFSNode{
		Name:    "test.txt",
		IsDir:   false,
		Content: "line1\nline2\nline2\nline3",
		ModTime: time.Now(),
	}
	shell.vfs.Root.Children = append(shell.vfs.Root.Children, fileNode)

	// Should output unique lines: line1, line2, line3, line2
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	shell.uniqCommand([]string{"/test.txt"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "line1\nline2\nline3") {
		t.Errorf("Expected 'line1\nline2\nline3\nline2' in output, got: %s", output)
	}
}

func TestTailCommand(t *testing.T) {
	shell := NewShell()

	// Create test file with multiple lines
	content := ""
	for i := 1; i <= 15; i++ {
		content += fmt.Sprintf("%d\n", i)
	}

	fileNode := &vfs.VFSNode{
		Name:    "test.txt",
		IsDir:   false,
		Content: content,
		ModTime: time.Now(),
	}
	shell.vfs.Root.Children = append(shell.vfs.Root.Children, fileNode)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	shell.tailCommand([]string{"/test.txt"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "7\n8\n9\n10\n11\n12\n13\n14\n15\n") {
		t.Errorf("Expected '7\n8\n9\n10\n11\n12\n13\n14\n15\n' in output, got: %s", output)
	}

	oldStdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w

	shell.tailCommand([]string{"-n", "5", "/test.txt"})

	w.Close()
	os.Stdout = oldStdout

	io.Copy(&buf, r)
	output = buf.String()

	if !strings.Contains(output, "11\n12\n13\n14\n15\n") {
		t.Errorf("Expected '11\n12\n13\n14\n15\n' in output, got: %s", output)
	}

}

func TestMvCommand(t *testing.T) {
	shell := NewShell()

	// Create source file
	sourceFile := &vfs.VFSNode{
		Name:    "source.txt",
		IsDir:   false,
		Content: "test content",
		ModTime: time.Now(),
	}
	shell.vfs.Root.Children = append(shell.vfs.Root.Children, sourceFile)

	// Create target directory
	targetDir := &vfs.VFSNode{
		Name:    "target",
		IsDir:   true,
		ModTime: time.Now(),
	}
	shell.vfs.Root.Children = append(shell.vfs.Root.Children, targetDir)

	// Test move file
	shell.mvCommand([]string{"/source.txt", "/target"})

	// Check if file was moved
	_, err := shell.vfs.FindNode("/source.txt")
	if err == nil {
		t.Error("Source file should not exist after move")
	}

	_, err = shell.vfs.FindNode("/target/source.txt")
	if err != nil {
		t.Error("File should exist in target directory after move")
	}
}

func TestChownCommand(t *testing.T) {
	shell := NewShell()

	// Create test file
	fileNode := &vfs.VFSNode{
		Name:    "test.txt",
		IsDir:   false,
		Content: "test content",
		Owner:   "olduser",
		ModTime: time.Now(),
	}
	shell.vfs.Root.Children = append(shell.vfs.Root.Children, fileNode)

	// Test chown command
	shell.chownCommand([]string{"newuser", "/test.txt"})

	// Check if owner was changed
	node, err := shell.vfs.FindNode("/test.txt")
	if err != nil {
		t.Fatalf("File should exist: %v", err)
	}

	if node.Owner != "newuser" {
		t.Errorf("Expected owner 'newuser', got '%s'", node.Owner)
	}
}
