# パッケージとインポート

Go のすべてのファイルは必ず **パッケージ** に属します。パッケージはコードを整理し、名前空間を分ける仕組みです。

## package 宣言

ファイルの先頭に必ず `package` を書きます。

```go
// main.go
package main
```

```go
// rules/no_shared_repeated_message_rule.go
package rules
```

| パッケージ名 | 役割 |
|---|---|
| `main` | プログラムのエントリポイント。`func main()` を含む |
| それ以外（`rules` など） | ライブラリとして他のパッケージから利用される |

> **ポイント**: ディレクトリ名とパッケージ名は一致させるのが慣例です。`rules/` ディレクトリ内のファイルは `package rules` にします。

## import 文

外部パッケージや標準ライブラリを使うには `import` を書きます。

```go
// main.go
import (
    "github.com/K0H205/protolint-plugin-no-shared-repeated/rules"
    "github.com/yoheimuta/protolint/plugin"
)
```

```go
// rules/no_shared_repeated_message_rule.go
import (
    "sort"
    "strings"
    "unicode"

    "github.com/yoheimuta/go-protoparser/v4/parser"
    "github.com/yoheimuta/go-protoparser/v4/parser/meta"
    "github.com/yoheimuta/protolint/linter/report"
    "github.com/yoheimuta/protolint/linter/rule"
    "github.com/yoheimuta/protolint/linter/visitor"
)
```

### import のルール

1. **丸括弧でグループ化** — 複数のインポートは `()` でまとめます
2. **空行でグループ分け** — 標準ライブラリ（`sort`, `strings` など）と外部パッケージの間に空行を入れるのが慣例です
3. **パス = モジュールパス + 内部パス** — 例: `github.com/yoheimuta/protolint/linter/rule`
4. **使わないインポートはコンパイルエラー** — Go では未使用のインポートを残すとビルドが通りません

## go.mod — モジュール定義

プロジェクトのルートにある `go.mod` がモジュール（パッケージの集合体）を定義します。

```
module github.com/K0H205/protolint-plugin-no-shared-repeated

go 1.24

require (
    github.com/yoheimuta/go-protoparser/v4 v4.14.2
    github.com/yoheimuta/protolint v0.54.0
)
```

| 項目 | 意味 |
|---|---|
| `module` | このプロジェクトのモジュールパス（import 時に使う名前） |
| `go 1.24` | 使用する Go のバージョン |
| `require` | 直接依存するパッケージとそのバージョン |

## パッケージの使い方

インポートしたパッケージは **パッケージ名.識別子** でアクセスします。

```go
// plugin パッケージの RegisterCustomRules 関数を呼ぶ
plugin.RegisterCustomRules(
    rules.NewNoSharedRepeatedMessageRule(),
)
```

> **ポイント**: Go では大文字で始まる名前（`RegisterCustomRules`, `NewNoSharedRepeatedMessageRule`）だけが外部パッケージから参照できます。これを **エクスポート** と呼びます。小文字で始まる名前はパッケージ内でのみ使えます。
