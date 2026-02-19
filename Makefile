.PHONY: help build test lint ci clean cdk-synth cdk-deploy cdk-destroy deploy destroy

GO ?= go
BIN_DIR ?= bin
STAGE ?= lab

help:
	@printf "%s\n" \
		"Targets:" \
		"  build         Build local binaries into $(BIN_DIR)/" \
		"  test          Run Go tests" \
		"  lint          Run basic Go lint (go vet)" \
		"  ci            Run test + lint" \
		"  cdk-synth     Synth CDK (STAGE=$(STAGE))" \
		"  cdk-deploy    Deploy CDK (STAGE=$(STAGE), requires AWS_PROFILE)" \
		"  cdk-destroy   Destroy CDK (STAGE=$(STAGE), requires AWS_PROFILE)"

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

build: $(BIN_DIR)
	$(GO) build -trimpath -o $(BIN_DIR)/orchestrator ./cmd/orchestrator
	$(GO) build -trimpath -o $(BIN_DIR)/agent-runner ./cmd/agent-runner

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...

ci: lint test

clean:
	rm -rf $(BIN_DIR)

cdk-synth:
	cd infra/cdk && npm ci && npx cdk synth -c stage=$(STAGE)

cdk-deploy:
	cd infra/cdk && npm ci && npx cdk deploy --all -c stage=$(STAGE)

cdk-destroy:
	cd infra/cdk && npm ci && npx cdk destroy --all -c stage=$(STAGE)

deploy: cdk-deploy

destroy: cdk-destroy

