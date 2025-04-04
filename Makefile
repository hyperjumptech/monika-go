APP_NAME := monika
CMD_DIR := ./cmd/monika
OUTPUT_DIR := bin

.PHONY: all build run clean fmt test

all: build

build:
	@echo "🔨 Building $(APP_NAME)..."
	@mkdir -p $(OUTPUT_DIR)
	@go build -o $(OUTPUT_DIR)/$(APP_NAME) $(CMD_DIR)
	@echo "✅ Built binary at $(OUTPUT_DIR)/$(APP_NAME)"

run:
	@echo "🚀 Running $(APP_NAME)..."
	@go run $(CMD_DIR)

clean:
	@echo "🧹 Cleaning build output..."
	@rm -rf $(OUTPUT_DIR)
	@echo "✅ Clean complete."

fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...

test:
	@echo "🧪 Running tests..."
	@go test ./...