## **目次**

- 目的 (Japanese)
- 使い方 (Japanese)
- .ragignore について (Japanese)
- コンパイル方法 (Japanese)
- Purpose (English)
- Usage (English)
- About .ragignore (English)
- How to Compile (English)

## **目的 (Japanese)**

Ragweaver は、ローカルにクローン済みの Git リポジトリやディレクトリ内のファイルをテキストとして書き出し、RAG の埋め込みをサポートするツールです。

- 無視(除外)したいファイルやフォルダを定義できます。(.ragignoreに定義)
- コメント(\#)などを含むBash 互換のパターンマッチ(\*\*)をサポートしています。

## **使い方 (Japanese)**

1. ローカルに対象の Git リポジトリをクローンする。 (リモートの URL だけを指定しても動作しません。すでにクローンしたローカルディレクトリに対してのみ実行できます。)
2. preamble.txtにRAGに埋め込みたい文字列を定義することができます。 指定がない場合、デフォルトでRagweaverは下記のプレアンブルを出力します。  
   `This text file represents the contents of a Git repository. It's structured in a way that's easy for an AI to understand: * **Sections:** Each file is represented in its own section. * **Separators:** Each section starts with a line of four hyphens (`----`). * **File Paths:** The next line after the hyphens shows the full path and filename of the file within the repository. * **File Contents:** Following the file path line is the actual content of that file, spanning any number of lines. **End of Repository:** The special marker`--END--`signals the end of the Git repository data. **Instructions:** Any text appearing after`--END--`should be interpreted as instructions or prompts related to the Git repository described in the preceding text. **Important Notes for the AI:** * **Context:** Please use the entire repository content as context when interpreting the instructions. * **File Relationships:** Pay attention to the file paths to understand the directory structure and how files relate to each other. * **Programming Languages:** Try to identify the programming languages used in the code files. This will help you understand the code's purpose and behavior. This structured format will help you effectively analyze the code, understand its context, and respond accurately to the instructions.`

3. .ragignore ファイルを用意する (任意) 例:  
    `# コメント行`  
    `node_modules/**`  
    `**/*.png`

   コメント行(\#)や空行は無視されます。

4. コマンドを実行する  
   基本的な使い方:  
   `ragweaver -r /path/to/local-repo [-p /path/to/preamble.txt] [-o /path/to/output_file.txt] [-i /path/to/.ragignore]`

   オプション

   - \-r /path/to/local-repo : 処理対象の Git リポジトリのパス (必須)
   - \-p /path/to/preamble.txt : 前置きメッセージ(プレアンブル)を出力ファイルの先頭に挿入します。
   - \-o /path/to/output_file.txt : 出力先ファイルを指定できます。指定がなければ output.txt に書き出されます。
   - \-i /path/to/.ragignore : 使用する .ragignore ファイルのパスを指定します。

## **.ragignore について (Japanese)**

Ragweaver は、指定されたリポジトリ内のファイルを処理する際に、.ragignore ファイルに記述されたルールに従ってファイルを無視します。 .ragignore ファイルは、以下のルールに従って処理されます。

- 各行は1つのパターンを表します。
- \#で始まる行はコメントとして扱われ、無視されます。
- 空行は無視されます。
- \*\* は、任意の数のディレクトリにマッチするワイルドカードとして使用できます。
- \* は、任意の数の文字にマッチするワイルドカードとして使用できます。
- / は、ディレクトリの区切り文字として使用します。

.ragignore ファイルの検索順序は以下の通りです。

1. \-i オプションで指定されたパス
2. リポジトリ直下の .ragignore
3. ホームディレクトリの .ragignore

\-i オプションで .ragignore ファイルのパスが指定された場合、リポジトリ内やホームディレクトリの .ragignore ファイルよりも優先されます。

## **コンパイル方法 (Japanese)**

1. Go 言語 (バージョン 1.18+ など) がインストールされていることを確認します。
2. 本リポジトリをクローンまたはダウンロードし、ディレクトリに移動します。
3. 依存ライブラリを取得します:  
   `go mod tidy`

4. ビルドします:  
   `go build -o ragweaver main.go`

   実行ファイル ragweaver が作成されます。

## **Makefileについて**

Macのインストールは

Bash

`make`  
`make install`

## **Purpose (English)**

Ragweaver is a command-line tool designed to prepare your local Git repositories or directories for use with Retrieval Augmented Generation (RAG) systems. It achieves this by traversing the file structure and outputting the contents as a single text file, optimized for embedding in a vector database.  
Key features include:

- **.ragignore Support:** Define files and folders to be excluded from processing, similar to .gitignore. Supports comments (\#) and Bash-compatible glob patterns (including \*\*).
- **Customizable Preamble:** Add a preamble to the output file to provide context for the RAG system. A default preamble is provided if none is specified.

## **Usage (English)**

1. Clone the target Git repository locally. (Specifying only the remote URL will not work. You can only run it against an already cloned local directory.)  
   You can define the string you want to embed in RAG in preamble.txt. If not specified, Ragweaver outputs the following preamble by default.
   `This text file represents the contents of a Git repository. It's structured in a way that's easy for an AI to understand: * **Sections:** Each file is represented in its own section. * **Separators:** Each section starts with a line of four hyphens (`----`). * **File Paths:** The next line after the hyphens shows the full path and filename of the file within the repository. * **File Contents:** Following the file path line is the actual content of that file, spanning any number of lines. **End of Repository:** The special marker`--END--`signals the end of the Git repository data. **Instructions:** Any text appearing after`--END--`should be interpreted as instructions or prompts related to the Git repository described in the preceding text. **Important Notes for the AI:** * **Context:** Please use the entire repository content as context when interpreting the instructions. * **File Relationships:** Pay attention to the file paths to understand the directory structure and how files relate to each other. * **Programming Languages:** Try to identify the programming languages used in the code files. This will help you understand the code's purpose and behavior. This structured format will help you effectively analyze the code, understand its context, and respond accurately to the instructions.`

2. **(Optional) Create a .ragignore File:** Specify files or folders to exclude. For example:  
   `# comment line`  
   `node_modules/**`  
   `**/*.png`

3. **Run the Command:**  
   `ragweaver -r /path/to/local-repo [-p /path/to/preamble.txt] [-o /path/to/output_file.txt] [-i /path/to/.ragignore]`

   **Options:**

   - \-r /path/to/local-repo: Path to the local Git repository (required).
   - \-p /path/to/preamble.txt: Path to a file containing a preamble to be inserted at the beginning of the output.
   - \-o /path/to/output_file.txt: Path to the output file. Defaults to output.txt.
   - \-i /path/to/.ragignore: Path to a specific .ragignore file to use.

## **About .ragignore (English)**

Ragweaver uses .ragignore files to determine which files and directories should be ignored during processing. Here's how it works:

- **Pattern Matching:** Each line in the .ragignore file represents a single pattern.
- **Comments:** Lines starting with \# are treated as comments and ignored.
- **Wildcards:**
  - \*\* matches any number of directories.
  - \* matches any number of characters.
- **Directory Separator:** / is used as the directory separator.

**.ragignore Search Order:**

1. Path specified by the \-i option.
2. .ragignore in the repository's root directory.
3. .ragignore in the user's home directory.

If the \-i option is used, it overrides .ragignore files found in the repository or home directory.

## **How to Compile (English)**

1. **Install Go:** Ensure you have Go 1.18 or later installed.
2. **Clone the Repository:** Clone this repository to your local machine.
3. **Get Dependencies:**  
   `go mod tidy`

4. **Build:**  
   `go build -o ragweaver main.go`

This will create an executable file named ragweaver.

### **Makefile (English)**

To install on Mac, use:

`make`  
`make install`
