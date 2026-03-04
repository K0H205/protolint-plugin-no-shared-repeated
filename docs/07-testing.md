# テスト

## Go のテストの基本

Go には標準でテストの仕組みが組み込まれています。外部フレームワークは不要です。

### テストファイルの命名規則

```
rules/
├── no_shared_repeated_message_rule.go       ← 本体
└── no_shared_repeated_message_rule_test.go  ← テスト
```

- ファイル名が `_test.go` で終わるファイルがテストファイル
- ビルド時には含まれない（テスト実行時のみ使われる）

### テスト用パッケージ

```go
// rules/no_shared_repeated_message_rule_test.go:1
package rules_test
```

`rules_test` のように `_test` サフィックスを付けると、**外部テストパッケージ** になります。つまり `rules` パッケージのエクスポートされた（大文字で始まる）識別子しか使えません。これにより、公開 API としての正しさをテストできます。

## テスト関数

テスト関数は `Test` で始まり、`*testing.T` を引数に取ります。

```go
// rules/no_shared_repeated_message_rule_test.go:11
func TestNoSharedRepeatedMessageRule_OK(t *testing.T) {
    // テストのロジック
}
```

| 規則 | 説明 |
|---|---|
| 関数名は `Test` で始まる | `Test` の後は大文字で始める |
| 引数は `t *testing.T` | テストのヘルパーメソッドを提供 |
| 戻り値なし | 成功/失敗は `t` のメソッドで報告 |

## テストヘルパーメソッド

### t.Fatalf — 即座にテストを中止

```go
// rules/no_shared_repeated_message_rule_test.go:16
t.Fatalf("failed to open ok.proto: %v", err)
```

エラーメッセージを出力し、**そのテスト関数をただちに停止** します。後続の処理に致命的な問題がある場合に使います。

### t.Errorf — エラーを記録して続行

```go
// rules/no_shared_repeated_message_rule_test.go:30
t.Errorf("expected no failures for ok.proto, got %d:", len(failures))
```

エラーメッセージを記録しますが、**テストの実行は続行** します。

### t.Logf — ログ出力

```go
// rules/no_shared_repeated_message_rule_test.go:64
t.Logf("failure: %s", f)
```

テスト実行時に情報を出力します。`-v` フラグ付きで実行した場合のみ表示されます。

### t.Error vs t.Fatal の使い分け

| メソッド | テスト続行 | 使いどころ |
|---|---|---|
| `t.Errorf` / `t.Error` | 続行する | 期待値との不一致を報告 |
| `t.Fatalf` / `t.Fatal` | 即座に停止 | 後続のテストが実行不可能なとき |

## テストのパターン

このプロジェクトのテストは以下の流れで構成されています。

```go
func TestNoSharedRepeatedMessageRule_NG(t *testing.T) {
    // 1. ルールを生成
    r := rules.NewNoSharedRepeatedMessageRule()

    // 2. テスト用のprotoファイルを開く
    f, err := os.Open("../testdata/ng.proto")
    if err != nil {
        t.Fatalf("failed to open ng.proto: %v", err)
    }
    defer f.Close()

    // 3. protoファイルをパースする
    proto, err := protoparser.Parse(f)
    if err != nil {
        t.Fatalf("failed to parse ng.proto: %v", err)
    }

    // 4. ルールを適用する
    failures, err := r.Apply(proto)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // 5. 結果を検証する
    if len(failures) != 2 {
        t.Fatalf("expected 2 failures for ng.proto, got %d", len(failures))
    }
}
```

## テストの実行

```bash
# すべてのテストを実行
go test ./...

# 特定のパッケージのテストを実行
go test ./rules/

# 特定のテスト関数を実行
go test ./rules/ -run TestNoSharedRepeatedMessageRule_OK

# 詳細出力（t.Logf の出力も表示）
go test -v ./rules/
```

## 型アサーション（テスト内で使われている構文）

テストコードにはありませんが、ルール本体で使われている重要な構文です。

```go
// rules/no_shared_repeated_message_rule.go:68
field, ok := body.(*parser.Field)
if !ok {
    continue
}
```

**型アサーション** は、インターフェース型の値を具体的な型に変換します。

- `body` はインターフェース型（`parser.Message` の body 要素）
- `body.(*parser.Field)` で「body は `*parser.Field` か？」を確認
- `ok == true` なら `field` に変換後の値が入る
- `ok == false` なら `field` はゼロ値、`continue` でスキップ

> **注意**: `ok` を受け取らずに `field := body.(*parser.Field)` と書くと、型が合わない場合に **panic**（実行時エラー）が発生します。安全に使うには必ず `ok` を受け取りましょう。
