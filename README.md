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

## Example Output

```
[ng.proto:13:3] "ItemInfo" is used as a repeated field in multiple messages ("CreateOrderRequest", "GetOrderResponse"). Consider defining separate messages for each usage to allow independent evolution.
[ng.proto:19:3] "ItemInfo" is used as a repeated field in multiple messages ("CreateOrderRequest", "GetOrderResponse"). Consider defining separate messages for each usage to allow independent evolution.
```

## Running Tests

```bash
go test ./rules/... -v
```
