package main

import (
    "flag"
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"
    "unicode/utf8"

    "golang.org/x/text/encoding/japanese"
    "golang.org/x/text/transform"
)

func main() {
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, `使い方: %s [-r <depth>] <フォルダのパス>

オプション:
  -r <depth>    再帰的に探索する深さ（0なら再帰しない）
  -h, --help    このヘルプを表示

`, os.Args[0])
    }
    var recursiveDepth int
    flag.IntVar(&recursiveDepth, "r", 0, "再帰的に探索する深さ（0なら再帰しない）")
    flag.Parse()

    if flag.NArg() < 1 {
        flag.Usage()
        return
    }
    targetDir := flag.Arg(0)

    fmt.Printf("📁 処理対象フォルダ: %s\n", targetDir)
    fmt.Println("---------------------------------")

    convertedCount := 0

    // ファイル探索
    var walkFn filepath.WalkFunc = func(path string, info os.FileInfo, err error) error {
        if err != nil {
            log.Printf("❌ ファイルアクセス失敗: %s - %v\n", path, err)
            return nil
        }
        if info.IsDir() {
            // 深さ制限
            if recursiveDepth > 0 {
                rel, _ := filepath.Rel(targetDir, path)
                if rel != "." && len(filepath.SplitList(rel)) > recursiveDepth {
                    return filepath.SkipDir
                }
            }
            return nil
        }
        ext := filepath.Ext(info.Name())
        if ext != ".csv" && ext != ".txt" {
            return nil
        }

        // UTF-8判定
        isUTF8, err := isFileUTF8(path)
        if err != nil {
            log.Printf("❌ 判定失敗: %s - %v\n", info.Name(), err)
            return nil
        }
        if isUTF8 {
            fmt.Printf("🟦 既にUTF-8: %s\n", info.Name())
            return nil
        }

        // Shift-JIS判定（簡易: UTF-8でなければShift-JISとみなす）
        err = convertFileToUTF8(path)
        if err != nil {
            log.Printf("❌ 変換失敗: %s - %v\n", info.Name(), err)
            return nil
        }
        fmt.Printf("✅ 変換成功: %s\n", info.Name())
        convertedCount++
        return nil
    }

    if recursiveDepth > 0 {
        filepath.Walk(targetDir, walkFn)
    } else {
        files, err := os.ReadDir(targetDir)
        if err != nil {
            log.Fatalf("FATAL: フォルダの読み込みに失敗しました: %v", err)
        }
        for _, file := range files {
            if file.IsDir() {
                continue
            }
            ext := filepath.Ext(file.Name())
            if ext != ".csv" && ext != ".txt" {
                continue
            }
            filePath := filepath.Join(targetDir, file.Name())
            isUTF8, err := isFileUTF8(filePath)
            if err != nil {
                log.Printf("❌ 判定失敗: %s - %v\n", file.Name(), err)
                continue
            }
            if isUTF8 {
                fmt.Printf("🟦 既にUTF-8: %s\n", file.Name())
                continue
            }
            err = convertFileToUTF8(filePath)
            if err != nil {
                log.Printf("❌ 変換失敗: %s - %v\n", file.Name(), err)
                continue
            }
            fmt.Printf("✅ 変換成功: %s\n", file.Name())
            convertedCount++
        }
    }

    fmt.Println("---------------------------------")
    fmt.Printf("✨ 処理完了！ %d個のファイルをUTF-8に変換しました。\n", convertedCount)
}

// ファイルがUTF-8か判定
func isFileUTF8(filePath string) (bool, error) {
    f, err := os.Open(filePath)
    if err != nil {
        return false, err
    }
    defer f.Close()
    buf := make([]byte, 4096)
    n, err := f.Read(buf)
    if err != nil && err != io.EOF {
        return false, err
    }
    return utf8.Valid(buf[:n]), nil
}

// Shift-JISからUTF-8に変換
func convertFileToUTF8(filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("ファイルを開けませんでした: %w", err)
    }
    defer file.Close()
    reader := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())
    utf8Bytes, err := io.ReadAll(reader)
    if err != nil {
        return fmt.Errorf("文字コードの変換に失敗しました: %w", err)
    }
    err = os.WriteFile(filePath, utf8Bytes, 0644)
    if err != nil {
        return fmt.Errorf("ファイルの書き込みに失敗しました: %w", err)
    }
    return nil
}