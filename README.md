
csvファイルを一括でshiftjisからutf8に変換

go run . <変換csvのあるディレクトリパス>

find <ディレクトリ郡のあるパス> -type d -name "*_csv" -print0 | xargs -0 -I {} go run . {}