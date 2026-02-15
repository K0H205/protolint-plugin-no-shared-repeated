# Example: protolint-plugin-no-shared-repeated

This directory contains an example project demonstrating how to use the `NO_SHARED_REPEATED_MESSAGE` custom protolint plugin.

## Prerequisites

- [Go](https://go.dev/dl/) 1.24+
- [protolint](https://github.com/yoheimuta/protolint) installed and available in `$PATH`

```bash
go install github.com/yoheimuta/protolint/cmd/protolint@v0.54.0
```

## Quick Start

```bash
cd example

# Build the plugin and lint all proto files
make

# Or run each step separately:
make build       # compile the plugin
make lint-good   # lint the good proto (no warnings expected)
make lint-bad    # lint the bad proto (warnings expected)
```

## What's Inside

```
example/
├── Makefile                       # Build & run commands
├── .protolint.yaml                # protolint config enabling the custom rule
├── README.md                      # This file
└── proto/
    ├── order_service_bad.proto    # BAD pattern  — triggers warnings
    └── order_service_good.proto   # GOOD pattern — passes cleanly
```

### Bad Pattern (`order_service_bad.proto`)

A single `ItemInfo` message is used as `repeated` in three different messages
(`CreateOrderRequest`, `GetOrderResponse`, `ListRecommendationsResponse`).
This means any field added for one use case leaks into all the others.

```protobuf
message CreateOrderRequest {
  repeated ItemInfo items = 2;      // NG
}
message GetOrderResponse {
  repeated ItemInfo items = 3;      // NG
}
message ListRecommendationsResponse {
  repeated ItemInfo items = 2;      // NG
}
```

Running the linter produces warnings:

```
[order_service_bad.proto:22:3] "ItemInfo" is used as a repeated field in multiple messages
("CreateOrderRequest", "GetOrderResponse", "ListRecommendationsResponse").
Consider defining separate messages for each usage to allow independent evolution.
```

### Good Pattern (`order_service_good.proto`)

Each context defines its own message type (`CreateOrderItem`, `OrderResponseItem`,
`RecommendationItem`), so they can evolve independently. The linter produces **no warnings**.

## Configuration

`.protolint.yaml` disables all built-in rules and enables only the custom rule:

```yaml
lint:
  rules:
    all_default: false
    add:
      - NO_SHARED_REPEATED_MESSAGE
```

To use this plugin alongside built-in rules, remove the `all_default: false` line.
