# protolint-plugin-no-shared-repeated

A [protolint](https://github.com/yoheimuta/protolint) custom rule plugin that detects when a message type is used as a `repeated` field in multiple different messages.

## Problem

When a message type is shared as `repeated` across request and response messages, adding a field for one use case (e.g. `quantity` on the request side) inadvertently affects the other:

```protobuf
message ItemInfo { string id = 1; }

// ItemInfo is shared as repeated in both request and response
message CreateOrderRequest { repeated ItemInfo items = 1; }
message GetOrderResponse   { repeated ItemInfo items = 1; }
```

This plugin reports these usages as `NO_SHARED_REPEATED_MESSAGE` violations so you can define separate messages for each context.

## Installation

```bash
go build -o protolint-plugin-no-shared-repeated .
```

## Usage

```bash
protolint lint -plugin ./protolint-plugin-no-shared-repeated your_file.proto
```

Add the rule to your `.protolint.yaml`:

```yaml
lint:
  rules:
    add:
      - NO_SHARED_REPEATED_MESSAGE
```

## Detection Rules

- A message type used as a `repeated` field in **2 or more different messages** is a violation.
- A message type used as `repeated` in only **one** message is OK.
- Non-repeated (singular) message fields are not checked.
- Scalar types (`string`, `int32`, etc.) are not checked.
- Enum types defined in the same file are not checked (sharing an enum does not have the same evolution problem as sharing a message).
- Type references are resolved within the file, following protobuf scoping rules:
  - `ItemInfo`, `example.ItemInfo`, and `.example.ItemInfo` are recognized as the same type when they refer to the same definition.
  - Nested types with the same short name in different scopes (e.g. `A.Item` vs `B.Item`) are treated as distinct types.
- Nested parent messages are reported by their full path (e.g. `Outer.Inner`).
- Failures are reported in file-position order.

## Options

### `-allowed_types`

Some types are intentionally shared (e.g. `google.protobuf.Timestamp` or a common `Money` type). Pass a comma-separated list of type names to exclude them from this rule. The flag is passed to the plugin binary itself inside the `-plugin` value:

```bash
protolint lint -plugin "./protolint-plugin-no-shared-repeated -allowed_types=google.protobuf.Timestamp,example.Money" your_file.proto
```

Names are matched against both the reference as written in the field and the resolved package-relative name, with any leading dot ignored.

## Limitations

- **Per-file analysis only.** protolint applies plugin rules to one file at a time, so a type imported and used as `repeated` in messages across *different* `.proto` files is not detected.
- **Imported types are matched textually.** Types defined in other files (e.g. `google.protobuf.Timestamp`) cannot be fully resolved; different spellings are normalized (leading dot and own-package prefix stripped) but aliasing beyond that is not handled.
- **Imported enum types are reported like messages.** Only enums defined in the same file are auto-excluded; use `-allowed_types` for imported enums.

## Example Output

```
[ng.proto:13:3] "ItemInfo" is used as a repeated field in multiple messages ("CreateOrderRequest", "GetOrderResponse"). Consider defining separate messages for each usage to allow independent evolution.
[ng.proto:19:3] "ItemInfo" is used as a repeated field in multiple messages ("CreateOrderRequest", "GetOrderResponse"). Consider defining separate messages for each usage to allow independent evolution.
```

## Running Tests

```bash
go test ./rules/... -v
```
