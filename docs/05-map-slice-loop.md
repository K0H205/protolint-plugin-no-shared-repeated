# マップ・スライス・ループ

## スライス（Slice）

スライスは **可変長の配列** です。Go で最もよく使うデータ構造の一つです。

### 宣言

```go
// rules/no_shared_repeated_message_rule.go:120
var result []string  // string のスライス。初期値は nil
```

`[]string` が「string のスライス」型です。

### append で要素追加

```go
// rules/no_shared_repeated_message_rule.go:124
result = append(result, info.parentMessage)
```

`append` は組み込み関数で、スライスに要素を追加した **新しいスライス** を返します。元のスライスは変更されないため、必ず戻り値を受け取ります。

### 複数の戻り値とスライス

```go
func (r NoSharedRepeatedMessageRule) Apply(proto *parser.Proto) ([]report.Failure, error) {
```

`[]report.Failure` は `report.Failure` のスライスです。

## マップ（Map）

マップは **キーと値のペア** を保持するデータ構造です（他の言語の辞書/ハッシュマップ）。

### make で初期化

```go
// rules/no_shared_repeated_message_rule.go:45
usages: make(map[string][]usageInfo),
```

`make(map[キーの型]値の型)` で空のマップを作ります。

| 式 | 意味 |
|---|---|
| `map[string][]usageInfo` | キーが `string`、値が `[]usageInfo`（スライス）のマップ |
| `map[string]struct{}` | キーが `string`、値が空構造体のマップ（セットとして使う） |

### マップへのアクセスと追加

```go
// rules/no_shared_repeated_message_rule.go:73
v.usages[field.Type] = append(v.usages[field.Type], usageInfo{ ... })
```

- `v.usages[field.Type]` — キーで値を取得
- 存在しないキーにアクセスすると値の型のゼロ値（スライスなら `nil`）が返る
- `append(nil, element)` は正常に動作するので、キーの存在チェックが不要

### マップの存在チェック（カンマ ok イディオム）

```go
// rules/no_shared_repeated_message_rule.go:122
if _, ok := seen[info.parentMessage]; !ok {
    seen[info.parentMessage] = struct{}{}
    result = append(result, info.parentMessage)
}
```

`値, ok := マップ[キー]` の形式で、キーが存在するかを `ok`（`bool`）で判定できます。

- `ok == true` → キーが存在する
- `ok == false` → キーが存在しない
- `_` で値を捨てている（存在チェックだけが目的）

### 空構造体 `struct{}` をマップの値に使う

```go
seen := make(map[string]struct{})
seen[info.parentMessage] = struct{}{}
```

`struct{}` はメモリを消費しない空の構造体です。「セット（集合）」を表現する際に、マップの値として使う Go の定番パターンです。値に意味がないことを明示できます。

## ループ（for range）

Go のループは `for` だけです（`while` はありません）。

### スライスのループ

```go
// rules/no_shared_repeated_message_rule.go:67
for _, body := range msg.MessageBody {
    // body は各要素
}
```

`range` はインデックスと要素を返します。インデックスが不要なら `_` で捨てます。

```go
// rules/no_shared_repeated_message_rule.go:92
for _, info := range infos {
    v.AddFailuref(info.pos, ...)
}
```

### マップのループ

```go
// rules/no_shared_repeated_message_rule.go:85
for typeName, infos := range v.usages {
    // typeName はキー、infos は値
}
```

マップの `range` はキーと値を返します。

### range の戻り値まとめ

| 対象 | 第1戻り値 | 第2戻り値 |
|---|---|---|
| スライス `[]T` | インデックス（`int`） | 要素（`T`） |
| マップ `map[K]V` | キー（`K`） | 値（`V`） |
| 文字列 `string` | インデックス（`int`） | 文字（`rune`） |

## len — 長さの取得

```go
// rules/no_shared_repeated_message_rule.go:87
if len(parents) < 2 {
    continue
}
```

`len()` は組み込み関数で、スライス・マップ・文字列などの長さを返します。

## sort — ソート

```go
// rules/no_shared_repeated_message_rule.go:90
sort.Strings(parents)
```

標準ライブラリ `sort` パッケージの `Strings` 関数で、文字列スライスを昇順にソートします（破壊的操作 — 元のスライスが変更される）。

## strings.Join — 文字列結合

```go
// rules/no_shared_repeated_message_rule.go:91
parentList := `"` + strings.Join(parents, `", "`) + `"`
```

スライスの要素を指定した区切り文字で結合します。バッククォート `` ` `` はエスケープ不要な **raw 文字列リテラル** です。
