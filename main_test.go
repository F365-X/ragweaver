package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ... (既存のコード: getIgnoreList, shouldIgnore, processRepository, main は省略)

// TestShouldIgnore_CommentLines は、
// 実際に .ragignore ファイルを生成してコメント行や空行が含まれた場合でも
// getIgnoreList → shouldIgnore が正しく動作するかをテストします。
func TestShouldIgnore_CommentLines(t *testing.T) {
	// 1. テスト用の一時ディレクトリを作成 (Go 1.15以降なら t.TempDir が使えます)
	tmpDir := t.TempDir()

	// 2. テスト用の .ragignore ファイルを作成し、コメント行(#)や空行などを含む内容を書き込み
	ragPath := filepath.Join(tmpDir, ".ragignore")
	ragContent := `# This is a comment, should be ignored
node_modules/**
# Another comment
image.png
`
	// ragignore があれば削除 (エラーは無視して進む)
	if _, err := os.Stat(ragPath); err == nil {
		// 存在するなら削除
		_ = os.Remove(ragPath)
	}
	// ファイルに書き出し
	if err := os.WriteFile(ragPath, []byte(ragContent), 0644); err != nil {
		t.Fatalf("Failed to write .ragignore: %v", err)
	}

	// 3. getIgnoreList を呼び出し、コメント行(#)・空行がスキップされた ignoreList を取得
	ignoreList, err := getIgnoreList(ragPath)
	if err != nil {
		t.Fatalf("getIgnoreList failed: %v", err)
	}

	// 4. テストケース定義
	// 「node_modules/**」「image.png」だけが有効パターンになっている想定。
	// (コメント行と空行は skip される)
	type testCase struct {
		path string
		want bool
	}
	testCases := []testCase{
		{path: "node_modules/foo.js", want: true}, // node_modules/** により無視
		{path: "image.png", want: true},           // image.png は無視
		{path: "some.png", want: false},           // some.png は違う → 無視されない
		{path: "some.txt", want: false},           // パターン外 → 無視されない
	}

	// 5. テスト実行: shouldIgnore(path, ignoreList) の結果が期待どおりかチェック
	for _, tc := range testCases {
		got := shouldIgnore(tc.path, ignoreList)
		if got != tc.want {
			t.Errorf("shouldIgnore(%q) = %v; want %v", tc.path, got, tc.want)
		}
	}
}

// TestProcessRepository は、processRepository の動作をテストします
func TestProcessRepository(t *testing.T) {
	// 一時ディレクトリ
	tempDir := t.TempDir()

	// ファイル: "tempDir/image.png" (無視される想定)
	err := os.WriteFile(filepath.Join(tempDir, "image.png"), []byte("PNG DATA"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// ファイル: "tempDir/hello.txt" (無視されない想定)
	err = os.WriteFile(filepath.Join(tempDir, "hello.txt"), []byte("Hello"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// ディレクトリ: "tempDir/subdir"
	subdirPath := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subdirPath, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// ファイル: "tempDir/subdir/hello.txt" (無視される想定)
	err = os.WriteFile(filepath.Join(subdirPath, "hello.txt"), []byte("Hello from subdir"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// (1) テスト用の .ragignore を作成
	ragPath := filepath.Join(tempDir, ".ragignore")
	ragContent := `
image.png
subdir/**
`
	if err := os.WriteFile(ragPath, []byte(ragContent), 0644); err != nil {
		t.Fatalf("Failed to write .ragignore: %v", err)
	}

	// (2) getIgnoreList で読み込む
	ignoreList, err := getIgnoreList(ragPath)
	if err != nil {
		t.Fatalf("Failed to getIgnoreList: %v", err)
	}

	// (3) processRepository 実行
	outPath := filepath.Join(tempDir, "output.txt")
	outFile, err := os.Create(outPath)
	if err != nil {
		t.Fatal(err)
	}
	defer outFile.Close()

	err = processRepository(tempDir, ignoreList, outFile)
	if err != nil {
		t.Fatalf("processRepository failed: %v", err)
	}

	// (4) 出力チェック
	outData, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	outStr := string(outData)

	// image.png は無視されるはず
	if strings.Contains(outStr, "image.png") {
		t.Errorf("image.png should be ignored, but found in output!")
	}
	// hello.txt は残るはず
	if !strings.Contains(outStr, "hello.txt") {
		t.Errorf("hello.txt should be in output, but not found!")
	}
	// subdir/hello.txt は無視されるはず
	if strings.Contains(outStr, "subdir/hello.txt") {
		t.Errorf("subdir/hello.txt should be ignored, but found in output!")
	}
}
