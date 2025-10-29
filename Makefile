# Output directory for binaries
BIN_DIR := bin

# Executables
PARSER := $(BIN_DIR)/parser
COGEN  := $(BIN_DIR)/cogen
EVALUATOR := $(BIN_DIR)/evaluator

GOFILES := $(shell find . -type f -name '*.go')

# Default target: build both
all: $(PARSER) $(COGEN) $(EVALUATOR)

# Build parser
$(PARSER): $(GOFILES)
	@mkdir -p $(BIN_DIR)
	go build -o $@ ./cmd/parser

# Build cogen
$(COGEN): $(GOFILES)
	@mkdir -p $(BIN_DIR)
	go build -o $@ ./cmd/cogen

$(EVALUATOR): $(GOFILES)
	@mkdir -p $(BIN_DIR)
	go build -o $@ ./cmd/evaluator/

# Build repl
$(REPL): $(GOFILES)
	@mkdir -p $(BIN_DIR)
	go build -o $@ ./cmd/repl

# Clean up build artifacts
clean:
	rm -rf $(BIN_DIR)

.PHONY: all clean
