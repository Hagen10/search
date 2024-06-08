# Define the executable name
EXEC = search

# Define the Go source file
SRC = main.go

# Define the output directory
OUT_DIR = bin

# Default target: build the executable
all: build

# Build the executable
build:
	@echo "Building the Go executable..."
	@mkdir -p $(OUT_DIR)
	@go build -o $(OUT_DIR)/$(EXEC) $(SRC)
	@echo "Build complete."

# Check if the directory is already in PATH and add it if not
add-to-path:
	@DIR=$(shell pwd)/$(OUT_DIR); \
	if cat ~/.zshrc | grep -q 'export PATH=$$PATH:'"$$DIR"; then \
		echo "'$$DIR' is already in PATH"; \
	else \
		echo "Adding '$$DIR' to PATH..."; \
		echo 'export PATH=$$PATH:'"$$DIR" >> ~/.zshrc; \
		echo "Run 'source ~/.zshrc' to update your current shell session."; \
	fi

# Clean up the output directory
clean:
	@echo "Cleaning up..."
	@rm -rf $(OUT_DIR)
	@echo "Cleanup complete."

# Phony targets
.PHONY: all build add-to-path clean
