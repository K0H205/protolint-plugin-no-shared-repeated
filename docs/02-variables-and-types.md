# 変数と型

## 基本的な型

Go の主な組み込み型です。このプロジェクトでは `string`, `bool`, `int` などを使っています。

| 型 | 説明 | 例 |
|---|---|---|
| `string` | 文字列 | `"NO_SHARED_REPEATED_MESSAGE"` |
| `bool` | 真偽値 | `true`, `false` |
| `int` | 整数 | `0`, `2` |
| `error` | エラーを表すインターフェース | `nil`（エラーなし） |
| `rune` | Unicode 文字（`int32` の別名） | `'A'` |

## 変数宣言

### var 宣言

型を明示して宣言します。

```go
// rules/no_shared_repeated_message_rule.go:120
var result []string
```

この例では `result` を `string` のスライス（配列のようなもの）として宣言しています。初期値は `nil`（空）です。

### 短縮宣言 `:=`

型を省略して、右辺の値から型を推論させます。**関数の中でのみ** 使えます。

```go
// rules/no_shared_repeated_message_rule.go:119
seen := make(map[string]struct{})
```

```go
// rules/no_shared_repeated_message_rule.go:86
parents := distinctParents(infos)
```

> **`var` と `:=` の使い分け**: 初期値を代入する場面では `:=` が簡潔です。ゼロ値で初期化したいだけなら `var` を使います。

## ゼロ値

Go の変数は宣言時に自動で **ゼロ値** が設定されます。

| 型 | ゼロ値 |
|---|---|
| `string` | `""`（空文字列） |
| `bool` | `false` |
| `int` | `0` |
| スライス / マップ / ポインタ | `nil` |

```go
// var result []string → result は nil（ゼロ値）
// append で要素を追加すれば自動的にメモリが確保される
result = append(result, info.parentMessage)
```

## 定数

このプロジェクトでは明示的な定数宣言はありませんが、文字列リテラルが定数的に使われています。

```go
func (r NoSharedRepeatedMessageRule) ID() string {
    return "NO_SHARED_REPEATED_MESSAGE"  // 文字列リテラル
}
```

定数を宣言する場合は `const` を使います。

```go
const RuleID = "NO_SHARED_REPEATED_MESSAGE"  // このプロジェクトでは使っていないが参考
```

## 型変換

Go では暗黙の型変換がなく、必ず明示的に変換します。

```go
// rules/no_shared_repeated_message_rule.go:115
return unicode.IsUpper(rune(typeName[0]))
```

`typeName[0]` は `byte` 型ですが、`unicode.IsUpper` は `rune` 型を受け取るため、`rune(...)` で明示的に変換しています。

## ブランク識別子 `_`

使わない値を捨てるために `_` を使います。

```go
// rules/no_shared_repeated_message_rule.go:131
var _ rule.Rule = NoSharedRepeatedMessageRule{}
```

この行は「`NoSharedRepeatedMessageRule` が `rule.Rule` インターフェースを満たすことをコンパイル時に確認する」というイディオムです。変数自体は使わないので `_` に代入しています。詳しくは [04-interface-and-embedding.md](./04-interface-and-embedding.md) で解説します。
