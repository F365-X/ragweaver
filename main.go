package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// getIgnoreList reads each line (pattern) from .ragignore.
func getIgnoreList(ragFilePath string) ([]string, error) {
	f, err := os.Open(ragFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ignoreList []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comment lines (#).
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		ignoreList = append(ignoreList, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return ignoreList, nil
}

// shouldIgnore checks if a file should be ignored based on the ignoreList.
// Uses doublestar to support patterns that span directories, like "**/...".
func shouldIgnore(filePath string, ignoreList []string) bool {
	for _, pattern := range ignoreList {
		matched, err := doublestar.PathMatch(pattern, filePath)
		if err != nil {
			// If the pattern is syntactically invalid, log a warning and ignore it.
			log.Printf("[WARN] invalid pattern %q => %v\n", pattern, err)
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

// processRepository recursively explores the files in the repository
// and writes only those that do not match .ragignore to the outputFile.
func processRepository(repoPath string, ignoreList []string, outputFile *os.File) error {
	return filepath.Walk(repoPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Determine the relative path from the repository.
		relativePath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}

		// Standardize to "/" separators.
		relativePath = filepath.ToSlash(relativePath)

		// (1) Ignore (do not output) .ragignore.
		if relativePath == ".ragignore" {
			if info.IsDir() {
				// Skip if .ragignore is a directory name, just in case.
				return filepath.SkipDir
			}
			// Usually treated as a file, so skip without doing anything.
			return nil
		}

		// If it is a directory and matches the ignoreList, skip its contents.
		if info.IsDir() {
			if shouldIgnore(relativePath, ignoreList) {
				return filepath.SkipDir
			}
			return nil
		}

		// If it is a file and matches the ignoreList, skip it.
		if shouldIgnore(relativePath, ignoreList) {
			return nil
		}

		// Read and output the file contents.
		contents, err := os.ReadFile(path)
		if err != nil {
			log.Printf("[WARN] cannot read file %q: %v\n", path, err)
			return nil
		}

		// ----\n
		// (relative path)\n
		// (file contents)\n
		if _, err := outputFile.WriteString("----\n"); err != nil {
			return err
		}
		if _, err := outputFile.WriteString(relativePath + "\n"); err != nil {
			return err
		}
		if _, err := outputFile.WriteString(string(contents) + "\n"); err != nil {
			return err
		}

		return nil
	})
}

func main() {
	// Use the flag package for argument processing.
	repoPath := flag.String("r", "", "path to git repository")
	preambleFile := flag.String("p", "", "path to preamble.txt")
	outputFilePath := flag.String("o", "output.txt", "path to output_file.txt")
	ragignorePath := flag.String("i", "", "path to .ragignore") // Option to specify the path to .ragignore.

	flag.Parse()

	// Error if repository path is not specified.
	if *repoPath == "" {
		fmt.Println("Usage: ragweaver -r /path/to/git/repository [-p /path/to/preamble.txt] [-o /path/to/output_file.txt] [-i /path/to/.ragignore]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Resolve the path to .ragignore.
	if *ragignorePath == "" {
		// If not specified in the options

		// Try .ragignore directly under the repository.
		*ragignorePath = filepath.Join(*repoPath, ".ragignore")
		if runtime.GOOS == "windows" {
			// If Windows, change the path separator to "\".
			*ragignorePath = strings.ReplaceAll(*ragignorePath, "/", "\\")
		}

		if _, err := os.Stat(*ragignorePath); os.IsNotExist(err) {
			// If not in the repository, try .ragignore in the home directory.
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatal(err)
			}
			*ragignorePath = filepath.Join(homeDir, ".ragignore")
		}
	}

	// Get ignoreList.
	var ignoreList []string
	if _, err := os.Stat(*ragignorePath); err == nil {
		list, err := getIgnoreList(*ragignorePath)
		if err != nil {
			log.Printf("[WARN] Failed to read .ragignore file: %v\n", err)
		} else {
			ignoreList = list
		}
	} else {
		// If .ragignore is not found, it is empty.
		ignoreList = []string{}
	}

	// Create or overwrite the output file.
	outputFile, err := os.Create(*outputFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	// If preambleFile is specified, read and output it if it exists.
	if *preambleFile != "" {
		if _, err := os.Stat(*preambleFile); os.IsNotExist(err) {
			log.Printf("[INFO] preamble file %q not found; skipping.\n", *preambleFile)
		} else {
			data, err := os.ReadFile(*preambleFile)
			if err != nil {
				log.Fatalf("Failed to read preamble file %s: %v", *preambleFile, err)
			}
			if _, err := outputFile.WriteString(string(data) + "\n"); err != nil {
				log.Fatal(err)
			}
		}
	} else {
		// If not specified, output the default text.
		const defaultPreamble = "This text file represents the contents of a Git repository. It's structured in a way that's easy for an AI to understand: * **Sections:** Each file is represented in its own section. * **Separators:** Each section starts with a line of four hyphens (`----`). * **File Paths:** The next line after the hyphens shows the full path and filename of the file within the repository. * **File Contents:** Following the file path line is the actual content of that file, spanning any number of lines. **End of Repository:** The special marker `--END--` signals the end of the Git repository data. **Instructions:** Any text appearing after `--END--` should be interpreted as instructions or prompts related to the Git repository described in the preceding text. **Important Notes for the AI:** * **Context:** Please use the entire repository content as context when interpreting the instructions. * **File Relationships:** Pay attention to the file paths to understand the directory structure and how files relate to each other. * **Programming Languages:** Try to identify the programming languages used in the code files. This will help you understand the code's purpose and behavior. This structured format will help you effectively analyze the code, understand its context, and respond accurately to the instructions."

		if _, err := outputFile.WriteString(defaultPreamble + "\n"); err != nil {
			log.Fatal(err)
		}
	}

	// Scan the repository and output.
	if err := processRepository(*repoPath, ignoreList, outputFile); err != nil {
		log.Fatal(err)
	}

	// Finally append --END--.
	if _, err := outputFile.WriteString("--END--"); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Repository contents written to %s.\n", *outputFilePath)
}
