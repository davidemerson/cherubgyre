package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// replacePackage replaces the old package name with the new package name in the given file.
func replacePackage(filePath, oldPkg, newPkg string) error {
	input, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	output := strings.Replace(string(input), oldPkg, newPkg, -1)

	err = os.WriteFile(filePath, []byte(output), 0644)
	if err != nil {
		return err
	}

	return nil
}

// findGoFiles finds all .go files and go.mod file in the current directory and subdirectories.
func findGoFiles(rootDir string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "go.mod")) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// getCurrentPackageName reads the package name from go.mod file.
func getCurrentPackageName(goModPath string) (string, error) {
	file, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("package name not found in go.mod")
}

// replacePackagesInProject replaces the old package name with the new one in all .go files and go.mod file.
func replacePackagesInProject(newPkg string) error {
	files, err := findGoFiles(".")
	if err != nil {
		return err
	}

	var goModPath string
	for _, file := range files {
		if strings.HasSuffix(file, "go.mod") {
			goModPath = file
			break
		}
	}

	if goModPath == "" {
		return fmt.Errorf("go.mod file not found")
	}

	oldPkg, err := getCurrentPackageName(goModPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		err := replacePackage(file, oldPkg, newPkg)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <new-package-name>")
		return
	}

	newPkg := os.Args[1]

	err := replacePackagesInProject(newPkg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Package name replacement completed successfully.")
}
