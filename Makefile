.PHONY: help registry clean install

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

registry: ## Regenerate the agent registry from all agent files
	@echo "Generating agent registry..."
	@python3 scripts/generate-registry.py
	@echo "Registry updated at .claude-plugin/agent-registry.json"

clean: ## Clean generated files
	@echo "Cleaning generated files..."
	@rm -f .claude-plugin/agent-registry.json

install: ## Install dependencies (none required for basic usage)
	@echo "No dependencies to install"
	@echo "Python 3.x is required for registry generation"

validate: ## Validate all agent files have proper frontmatter
	@echo "Validating agent registry..."
	@python3 scripts/validate-registry.py
