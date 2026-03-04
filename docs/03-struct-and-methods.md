# 構造体（struct）とメソッド

## 構造体の定義

`struct` は複数のフィールドをまとめたデータ型です。他の言語のクラスに近い概念です。

```go
// rules/no_shared_repeated_message_rule.go:19
type NoSharedRepeatedMessageRule struct{}
```

これはフィールドを持たない空の構造体です。メソッドを持つだけの「振る舞いの入れ物」として使います。

```go
// rules/no_shared_repeated_message_rule.go:51-54
type usageInfo struct {
    parentMessage string
    pos           meta.Position
}
```

こちらは2つのフィールドを持つ構造体です。

| フィールド | 型 | 説明 |
|---|---|---|
| `parentMessage` | `string` | 親メッセージの名前 |
| `pos` | `meta.Position` | ソースファイル内の位置情報 |

## 構造体の生成

### リテラルで生成

```go
// rules/no_shared_repeated_message_rule.go:73-76
usageInfo{
    parentMessage: msg.MessageName,
    pos:           field.Meta.Pos,
}
```

フィールド名を指定して値を設定します。指定しなかったフィールドはゼロ値になります。

### コンストラクタ関数

Go には `new` キーワードによるコンストラクタがありません。代わりに `New〇〇` という命名の関数を作る慣例があります。

```go
// rules/no_shared_repeated_message_rule.go:21-23
func NewNoSharedRepeatedMessageRule() NoSharedRepeatedMessageRule {
    return NoSharedRepeatedMessageRule{}
}
```

## メソッド

関数名の前に **レシーバ** を書くと、その構造体のメソッドになります。

### 値レシーバ

```go
// rules/no_shared_repeated_message_rule.go:25-27
func (r NoSharedRepeatedMessageRule) ID() string {
    return "NO_SHARED_REPEATED_MESSAGE"
}
```

`(r NoSharedRepeatedMessageRule)` がレシーバです。`r` はメソッド内で構造体にアクセスするための変数名です。

他のメソッドも同様です。

```go
func (r NoSharedRepeatedMessageRule) Purpose() string { ... }
func (r NoSharedRepeatedMessageRule) IsOfficial() bool { ... }
func (r NoSharedRepeatedMessageRule) Severity() rule.Severity { ... }
```

### ポインタレシーバ

```go
// rules/no_shared_repeated_message_rule.go:66
func (v *noSharedRepeatedVisitor) VisitMessage(msg *parser.Message) bool {
```

`*noSharedRepeatedVisitor` — `*` が付いているのでポインタレシーバです。

| レシーバの種類 | 書き方 | 特徴 |
|---|---|---|
| 値レシーバ | `(r Type)` | 構造体のコピーを受け取る。元の値は変更されない |
| ポインタレシーバ | `(v *Type)` | 構造体のポインタを受け取る。元の値を変更できる |

> **使い分けの目安**: 構造体の中身を変更する場合はポインタレシーバ、読み取りだけなら値レシーバ。`VisitMessage` は `v.usages` マップに書き込むのでポインタレシーバが必要です。

## メソッドの呼び出し

```go
// main.go:9-11
plugin.RegisterCustomRules(
    rules.NewNoSharedRepeatedMessageRule(),
)
```

```go
// rules/no_shared_repeated_message_rule.go:42
func (r NoSharedRepeatedMessageRule) Apply(proto *parser.Proto) ([]report.Failure, error) {
```

`Apply` は **複数の戻り値** を返しています。Go では `(型1, 型2)` のように複数の値を返せます。

```go
// テストコードでの呼び出し（rules/no_shared_repeated_message_rule_test.go:25）
failures, err := r.Apply(proto)
```

## パッケージレベルの関数

レシーバを持たない関数はパッケージレベルの関数です。

```go
// rules/no_shared_repeated_message_rule.go:108
func isMessageType(typeName string) bool {
```

```go
// rules/no_shared_repeated_message_rule.go:118
func distinctParents(infos []usageInfo) []string {
```

小文字で始まるので **エクスポートされない**（パッケージ外からは呼べない）関数です。
