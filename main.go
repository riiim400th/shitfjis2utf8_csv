package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func main() {
	// 1. コマンドライン引数から対象ディレクトリを取得
	if len(os.Args) < 2 {
		fmt.Println("エラー: 変換対象のフォルダパスを指定してください。")
		fmt.Printf("使い方: go run %s <フォルダのパス>\n", os.Args[0])
		return
	}
	targetDir := os.Args[1]

	// 2. ディレクトリ内のファイル一覧を取得
	files, err := os.ReadDir(targetDir)
	if err != nil {
		log.Fatalf("FATAL: フォルダの読み込みに失敗しました: %v", err)
	}

	fmt.Printf("📁 処理対象フォルダ: %s\n", targetDir)
	fmt.Println("---------------------------------")

	convertedCount := 0

	// 3. ファイルを一つずつ処理
	for _, file := range files {
		// ディレクトリは無視し、拡張子が.csvのファイルのみを対象とする
		if file.IsDir() || filepath.Ext(file.Name()) != ".csv" {
			continue
		}

		filePath := filepath.Join(targetDir, file.Name())

		// 4. Shift-JISからUTF-8への変換処理
		err := convertFileToUTF8(filePath)
		if err != nil {
			log.Printf("❌ 変換失敗: %s - %v\n", file.Name(), err)
			continue // エラーが発生しても次のファイルの処理を続ける
		}

		fmt.Printf("✅ 変換成功: %s\n", file.Name())
		convertedCount++
	}

	fmt.Println("---------------------------------")
	fmt.Printf("✨ 処理完了！ %d個のCSVファイルをUTF-8に変換しました。\n", convertedCount)
}

// convertFileToUTF8 は指定されたファイルの文字コードをShift-JISからUTF-8に変換して上書き保存します。
func convertFileToUTF8(filePath string) error {
	// Shift-JISとしてファイルを開く
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("ファイルを開けませんでした: %w", err)
	}
	defer file.Close()

	// Shift-JISからUTF-8に変換するリーダーを作成
	reader := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())

	// 変換後のUTF-8データをすべて読み込む
	utf8Bytes, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("文字コードの変換に失敗しました: %w", err)
	}

	// 元のファイルにUTF-8として上書き保存する
	// ファイルパーミッションは元のファイルを読み取るため0644（所有者に読み書き、その他に読み取り）を指定
	err = os.WriteFile(filePath, utf8Bytes, 0644)
	if err != nil {
		return fmt.Errorf("ファイルの書き込みに失敗しました: %w", err)
	}

	return nil
}