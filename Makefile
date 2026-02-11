.PHONY: help build test lint fmt vet clean demo demos demo-quick demo-tui demo-dashboard demo-full demo-clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build drift binary
	go build ./cmd/drift/

test: ## Run tests with race detector
	go test -race ./...

lint: vet fmt ## Run all linters

vet: ## Run go vet
	go vet ./...

fmt: ## Check formatting
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

cover: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

demo: build ## Record demo GIF (requires VHS)
	vhs demo-quick.tape

demos: build ## Generate all demo GIFs
	@echo "🎬 Generating all demos..."
	@./scripts/generate-demos.sh

demo-quick: build ## Generate quick demo (30s)
	vhs demo-quick.tape

demo-tui: build ## Generate TUI demo (50s)
	vhs demo-tui.tape

demo-dashboard: build ## Generate dashboard demo (40s)
	vhs demo-dashboard.tape

demo-full: build ## Generate full demo (60s)
	vhs demo.tape

demo-clean: ## Remove all demo files
	rm -f demo*.gif demo*.mp4 demo*.webm

clean: ## Remove build artifacts
	go clean
	rm -f drift coverage.out
	rm -f demo*.gif demo*.mp4 demo*.webm
