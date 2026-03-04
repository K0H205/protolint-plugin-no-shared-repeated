# エラーハンドリングと defer

## Go のエラーハンドリング

Go には `try-catch` がありません。代わりに **戻り値でエラーを返す** のが基本方針です。

### パターン: 戻り値の最後に `error`

```go
// rules/no_shared_repeated_message_rule.go:42
func (r NoSharedRepeatedMessageRule) Apply(proto *parser.Proto) ([]report.Failure, error) {
```

`([]report.Failure, error)` — 正常な結果とエラーの2つを返します。

### パターン: `if err != nil`

```go
// rules/no_shared_repeated_message_rule_test.go:14-17
f, err := os.Open("../testdata/ok.proto")
if err != nil {
    t.Fatalf("failed to open ok.proto: %v", err)
}
```

1. 関数を呼び、結果とエラーを受け取る
2. `err != nil` ならエラーが発生している
3. エラーを処理する（ここではテストを即座に失敗させている）

> **`nil`** は Go の「値がない」ことを表す特別な値です。ポインタ、インターフェース、スライス、マップ、チャネル、関数のゼロ値が `nil` です。

このパターンはプロジェクト内で何度も登場します。

```go
// rules/no_shared_repeated_message_rule_test.go:20-23
proto, err := protoparser.Parse(f)
if err != nil {
    t.Fatalf("failed to parse ok.proto: %v", err)
}

// rules/no_shared_repeated_message_rule_test.go:25-28
failures, err := r.Apply(proto)
if err != nil {
    t.Fatalf("unexpected error: %v", err)
}
```

### エラーを返す側

```go
// rules/no_shared_repeated_message_rule.go:84
func (v *noSharedRepeatedVisitor) Finally() error {
    // ... 処理 ...
    return nil  // エラーなし
}
```

エラーがなければ `nil` を返します。

## フォーマット動詞 `%v`

エラーメッセージ中の `%v` は Go の書式指定子です。

```go
t.Fatalf("failed to open ok.proto: %v", err)
```

| 動詞 | 意味 | 例 |
|---|---|---|
| `%v` | デフォルトの書式で表示 | `open file: no such file` |
| `%q` | ダブルクォート付き文字列 | `"ItemInfo"` |
| `%s` | 文字列そのまま | `ItemInfo` |
| `%d` | 整数 | `42` |

```go
// rules/no_shared_repeated_message_rule.go:95-96
v.AddFailuref(
    info.pos,
    `%q is used as a repeated field in multiple messages (%s).`,
    typeName,    // → %q に対応
    parentList,  // → %s に対応
)
```

## defer — 遅延実行

`defer` は **関数の終了時に実行される処理** を予約します。

```go
// rules/no_shared_repeated_message_rule_test.go:18
defer f.Close()
```

### 動作の流れ

```go
f, err := os.Open("../testdata/ok.proto")  // 1. ファイルを開く
if err != nil {
    t.Fatalf(...)                            // エラーなら即終了
}
defer f.Close()                              // 2. 関数終了時に Close を予約

proto, err := protoparser.Parse(f)           // 3. ファイルを使う
// ...
// 関数が終わると f.Close() が自動実行される   // 4. クリーンアップ
```

### defer のメリット

1. **リソースの閉じ忘れを防ぐ** — 開いた直後に `defer` で閉じる予約をする
2. **関数のどこで return しても実行される** — 途中で `return` やエラーが起きても確実にクリーンアップされる
3. **panic が起きても実行される** — 異常終了時でも `defer` は呼ばれる

### defer が使われる典型的な場面

| 場面 | コード |
|---|---|
| ファイルを閉じる | `defer f.Close()` |
| ミューテックスのアンロック | `defer mu.Unlock()` |
| データベース接続を閉じる | `defer db.Close()` |

> **注意**: `defer` は **関数の終了時** に実行されます。`for` ループ内で `defer` を使うと、ループではなく関数が終わるまで実行されないので注意が必要です。
