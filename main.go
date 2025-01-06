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

// getIgnoreList は .ragignore 内の各行(パターン)
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
		// 空行やコメント行(#)をスキップ
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

// shouldIgnore は、filePath が ignoreList のいずれかのパターンにマッチしたら true を返します。
// doublestar を使うことで "**/..." など階層をまたぐパターンに対応。
func shouldIgnore(filePath string, ignoreList []string) bool {
	for _, pattern := range ignoreList {
		matched, err := doublestar.PathMatch(pattern, filePath)
		if err != nil {
			// パターンが文法的に不正なら警告ログを出して無視
			log.Printf("[WARN] invalid pattern %q => %v\n", pattern, err)
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

// リポジトリ内のファイルを再帰的に探索し、 無視リストに該当しないものだけを outputFile に書き出します。
func processRepository(repoPath string, ignoreList []string, outputFile *os.File) error {
	return filepath.Walk(repoPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// リポジトリからの相対パスを求める
		relativePath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}

		// "/" 区切りに統一
		relativePath = filepath.ToSlash(relativePath)

		// (1) .ragignore を無視（出力しない）
		if relativePath == ".ragignore" {
			if info.IsDir() {
				// .ragignore がディレクトリ名になっている可能性があるので念の為スキップ
				return filepath.SkipDir
			}
			// 通常はファイル扱いなので何もしないでスキップ
			return nil
		}

		// ディレクトリで、かつ ignoreList にマッチしたら配下をスキップ
		if info.IsDir() {
			if shouldIgnore(relativePath, ignoreList) {
				return filepath.SkipDir
			}
			return nil
		}

		// ファイルで、かつ ignoreList にマッチしたらスキップ
		if shouldIgnore(relativePath, ignoreList) {
			return nil
		}

		// ファイル内容を読み込んで出力
		contents, err := os.ReadFile(path)
		if err != nil {
			log.Printf("[WARN] cannot read file %q: %v\n", path, err)
			return nil
		}

		// ----\n
		// (相対パス)\n
		// (ファイル内容)\n
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
	// 引数処理に flag パッケージを使う
	repoPath := flag.String("r", "", "path to git repository")
	preambleFile := flag.String("p", "", "path to preamble.txt")
	outputFilePath := flag.String("o", "output.txt", "path to output_file.txt")
	ragignorePath := flag.String("i", "", "path to .ragignore") // .ragignore のパスを指定するオプション

	flag.Parse()

	// リポジトリパスが指定されていない場合はエラー
	if *repoPath == "" {
		fmt.Println("Usage: ragweaver -r /path/to/git/repository [-p /path/to/preamble.txt] [-o /path/to/output_file.txt] [-i /path/to/.ragignore]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// .ragignore のパス解決
	if *ragignorePath == "" {
		// オプションで指定されなかった場合

		// リポジトリ直下の .ragignore を試す
		*ragignorePath = filepath.Join(*repoPath, ".ragignore")
		if runtime.GOOS == "windows" {
			// Windows ならパス区切りを "\" に
			*ragignorePath = strings.ReplaceAll(*ragignorePath, "/", "\\")
		}

		if _, err := os.Stat(*ragignorePath); os.IsNotExist(err) {
			// リポジトリになければホームディレクトリの .ragignore を試す
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatal(err)
			}
			*ragignorePath = filepath.Join(homeDir, ".ragignore")
		}
	}

	// ignoreList を取得
	var ignoreList []string
	if _, err := os.Stat(*ragignorePath); err == nil {
		list, err := getIgnoreList(*ragignorePath)
		if err != nil {
			log.Printf("[WARN] Failed to read .ragignore file: %v\n", err)
		} else {
			ignoreList = list
		}
	} else {
		// .ragignore が見つからないなら空
		ignoreList = []string{}
	}

	// 出力ファイルを作成or上書き
	outputFile, err := os.Create(*outputFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	// preambleFile が指定されていれば、存在すれば読み込んで出力
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
		// 未指定ならデフォルトの文言を出力
		const defaultPreamble = "This text file represents the contents of a Git repository. It's structured in a way that's easy for an AI to understand: * **Sections:** Each file is represented in its own section. * **Separators:** Each section starts with a line of four hyphens (`----`). * **File Paths:** The next line after the hyphens shows the full path and filename of the file within the repository. * **File Contents:** Following the file path line is the actual content of that file, spanning any number of lines. **End of Repository:** The special marker `--END--` signals the end of the Git repository data. **Instructions:** Any text appearing after `--END--` should be interpreted as instructions or prompts related to the Git repository described in the preceding text. **Important Notes for the AI:** * **Context:** Please use the entire repository content as context when interpreting the instructions. * **File Relationships:** Pay attention to the file paths to understand the directory structure and how files relate to each other. * **Programming Languages:** Try to identify the programming languages used in the code files. This will help you understand the code's purpose and behavior. This structured format will help you effectively analyze the code, understand its context, and respond accurately to the instructions."

		if _, err := outputFile.WriteString(defaultPreamble + "\n"); err != nil {
			log.Fatal(err)
		}
	}

	// リポジトリを走査して出力
	if err := processRepository(*repoPath, ignoreList, outputFile); err != nil {
		log.Fatal(err)
	}

	// 最後に --END-- を追記
	if _, err := outputFile.WriteString("--END--"); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Repository contents written to %s.\n", *outputFilePath)
}
