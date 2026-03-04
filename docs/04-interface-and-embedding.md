# インターフェースと構造体の埋め込み

## インターフェースとは

インターフェースは **メソッドの集合** を定義する型です。そのメソッドをすべて実装している構造体は、自動的にそのインターフェースを満たします（**暗黙的な実装**）。

Java のように `implements` キーワードを書く必要はありません。

## このプロジェクトでの例

`NoSharedRepeatedMessageRule` は `rule.Rule` インターフェースを満たしています。

```
rule.Rule インターフェース（protolint が定義）
├── ID() string
├── Purpose() string
├── IsOfficial() bool
├── Severity() rule.Severity
└── Apply(*parser.Proto) ([]report.Failure, error)
```

これらのメソッドがすべて定義されているので、`NoSharedRepeatedMessageRule` は自動的に `rule.Rule` として扱えます。

```go
func (r NoSharedRepeatedMessageRule) ID() string              { ... }
func (r NoSharedRepeatedMessageRule) Purpose() string          { ... }
func (r NoSharedRepeatedMessageRule) IsOfficial() bool         { ... }
func (r NoSharedRepeatedMessageRule) Severity() rule.Severity  { ... }
func (r NoSharedRepeatedMessageRule) Apply(...) (...)           { ... }
```

## コンパイル時のインターフェース検証

```go
// rules/no_shared_repeated_message_rule.go:131
var _ rule.Rule = NoSharedRepeatedMessageRule{}
```

これは Go の定番イディオムです。

- `NoSharedRepeatedMessageRule{}` を `rule.Rule` 型の変数に代入する
- もしインターフェースのメソッドが足りなければ **コンパイルエラー** になる
- `_`（ブランク識別子）に代入するので、実行時には何もしない
- つまり「コンパイル時にインターフェースの実装漏れを検出する安全策」

## 構造体の埋め込み（Embedding）

Go にはクラスの継承がありません。代わりに **埋め込み（Embedding）** で振る舞いを再利用します。

```go
// rules/no_shared_repeated_message_rule.go:58-62
type noSharedRepeatedVisitor struct {
    *visitor.BaseAddVisitor           // ← 埋め込み
    usages map[string][]usageInfo
}
```

`*visitor.BaseAddVisitor` がフィールド名なしで記述されています。これが埋め込みです。

### 埋め込みの効果

`BaseAddVisitor` が持つメソッド（例: `AddFailuref`）が `noSharedRepeatedVisitor` のメソッドであるかのように使えます。

```go
// rules/no_shared_repeated_message_rule.go:93-98
v.AddFailuref(
    info.pos,
    `%q is used as a repeated field in multiple messages (%s).`,
    typeName,
    parentList,
)
```

`AddFailuref` は `BaseAddVisitor` のメソッドですが、`v.AddFailuref(...)` のように直接呼べます。

### 継承との違い

| 特徴 | 継承（Java/Python） | 埋め込み（Go） |
|---|---|---|
| 関係 | is-a（AはBである） | has-a（AはBを持つ） |
| メソッド | 自動で引き継がれる | 自動で委譲される（見た目は同じ） |
| 多重継承 | 言語による制限あり | 複数の構造体を埋め込める |
| オーバーライド | 可能 | 同名メソッドを定義すれば上書きされる |

## 埋め込みフィールドの初期化

```go
// rules/no_shared_repeated_message_rule.go:43-46
v := &noSharedRepeatedVisitor{
    BaseAddVisitor: visitor.NewBaseAddVisitor(r.ID(), string(r.Severity())),
    usages:         make(map[string][]usageInfo),
}
```

- `BaseAddVisitor:` — 埋め込みフィールドも型名で初期化します
- `&noSharedRepeatedVisitor{...}` — `&` でポインタを取得しています（後述のポインタレシーバで必要）

## ポインタと `&`

```go
v := &noSharedRepeatedVisitor{ ... }
```

`&` は **アドレス演算子** で、値のポインタ（メモリ上のアドレス）を取得します。

- `noSharedRepeatedVisitor{}` → 値そのもの
- `&noSharedRepeatedVisitor{}` → その値へのポインタ（`*noSharedRepeatedVisitor` 型）

ポインタレシーバ `(v *noSharedRepeatedVisitor)` を持つメソッドを呼ぶには、ポインタが必要です。
