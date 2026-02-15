PLUGIN_NAME  := protolint-plugin-no-shared-repeated
PLUGIN_SRC   := ..
PLUGIN_BIN   := ./$(PLUGIN_NAME)
PROTO_DIR    := ./proto

.PHONY: all build lint lint-good lint-bad clean

## all: build the plugin and run lint on all proto files
all: build lint

## build: compile the plugin binary into this directory
build:
	go build -o $(PLUGIN_BIN) $(PLUGIN_SRC)

## lint: run protolint with the plugin against all example protos
lint: build
	@echo "=== Linting all proto files ==="
	-protolint lint -plugin $(PLUGIN_BIN) $(PROTO_DIR)
	@echo ""

## lint-good: lint only the good-pattern proto (expect 0 warnings)
lint-good: build
	@echo "=== Linting order_service_good.proto (expect NO warnings) ==="
	protolint lint -plugin $(PLUGIN_BIN) $(PROTO_DIR)/order_service_good.proto
	@echo ""

## lint-bad: lint only the bad-pattern proto (expect warnings)
lint-bad: build
	@echo "=== Linting order_service_bad.proto (expect warnings) ==="
	-protolint lint -plugin $(PLUGIN_BIN) $(PROTO_DIR)/order_service_bad.proto
	@echo ""

## clean: remove the built plugin binary
clean:
	rm -f $(PLUGIN_BIN)
