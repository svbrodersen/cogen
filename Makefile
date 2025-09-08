# Output directory for binaries
BIN_DIR := bin

# Executables
PARSER := $(BIN_DIR)/parser
COGEN  := $(BIN_DIR)/cogen

GOFILES := $(shell find . -type f -name '*.go')

# Default target: build both
all: $(PARSER) $(COGEN)

# Build parser
$(PARSER): $(GOFILES)
	@mkdir -p $(BIN_DIR)
	go build -o $@ ./cmd/parser

# Build cogen
$(COGEN): $(GOFILES)
	@mkdir -p $(BIN_DIR)
	go build -o $@ ./cmd/cogen

# Clean up build artifacts
clean:
	rm -rf $(BIN_DIR)

.PHONY: all clean
