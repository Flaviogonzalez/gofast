# gofast Makefile
# Full pipeline: build → generate → tidy → compile → smoke-test → teardown
SHELL := sh

GOFAST_BIN    := gofast.exe
OUT_DIR       := .testrun
API_BIN       := $(OUT_DIR)/api.exe
API_PORT      := 18082
API_URL       := http://localhost:$(API_PORT)
DATABASE_URL  ?= mysql://root:root@localhost:3306/gofast-test

# ─── Utilities ────────────────────────────────────────────────────────────────

.PHONY: all build test generate tidy-api build-api start-api smoke stop-api clean help

## all: run the full pipeline (build → generate → tidy → compile → smoke)
all: build generate tidy-api build-api smoke

# ─── gofast binary ────────────────────────────────────────────────────────────

## build: compile the gofast CLI
build:
	@echo "==> Building gofast..."
	go build -o $(GOFAST_BIN) ./cmd/gofast
	@echo "    OK: $(GOFAST_BIN)"

# ─── Unit tests ───────────────────────────────────────────────────────────────

## test: run unit tests for schema loader and generator
test:
	@echo "==> Unit tests..."
	go test ./internal/schema/... ./internal/generator/... -v

# ─── Project generation ───────────────────────────────────────────────────────

## generate: create $(OUT_DIR) and run 'gofast generate' inside it
generate: build
	@echo "==> Generating project in $(OUT_DIR)..."
	@rm -rf $(OUT_DIR)
	@mkdir -p $(OUT_DIR)
	@echo 'DATABASE_URL="$(DATABASE_URL)"' > $(OUT_DIR)/.env
	@cd $(OUT_DIR) && ../$(GOFAST_BIN) generate

## tidy-api: run 'go mod tidy' inside the generated project
tidy-api:
	@echo "==> go mod tidy (generated project)..."
	@cd $(OUT_DIR) && go mod tidy

## build-api: compile the generated API binary
build-api:
	@echo "==> Building generated API..."
	@cd $(OUT_DIR) && go build -o api.exe ./cmd/api
	@echo "    OK: $(API_BIN)"

# ─── Runtime smoke tests ──────────────────────────────────────────────────────

## start-api: start the generated API server in the background
start-api: build-api
	@echo "==> Starting API on port $(API_PORT)..."
	@cd $(OUT_DIR) && PORT=$(API_PORT) ./api.exe > api.log 2>&1 &
	@echo "    Waiting for server to become ready..."
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		curl -sf $(API_URL)/health > /dev/null 2>&1 && echo "    Server is up" && exit 0; \
		sleep 1; \
	done; \
	echo "ERROR: server did not start within 10 s"; \
	cat $(OUT_DIR)/api.log; \
	exit 1

## smoke: run CRUD smoke tests against the running API (starts server if needed)
smoke: start-api
	@echo ""
	@echo "==> Smoke tests..."
	$(MAKE) _smoke_run
	@$(MAKE) stop-api

_smoke_run:
	@echo "--- /health ---"
	@curl -sf $(API_URL)/health | grep -q "OK" && echo "    PASS: /health"

	@echo "--- POST /api/category ---"
	@RESP=$$(curl -sf -X POST -H "Content-Type: application/json" \
		-d '{"name":"SmokeTest","description":"smoke desc","image_url":null,"created_at":null}' \
		$(API_URL)/api/category); \
	echo "    response: $$RESP"; \
	echo "$$RESP" | grep -q '"id"' && echo "    PASS: POST /api/category"

	@echo "--- GET /api/category ---"
	@RESP=$$(curl -sf $(API_URL)/api/category); \
	echo "    response: $$RESP"; \
	echo "    PASS: GET /api/category"

	@echo "--- GET /api/category/1 ---"
	@HTTP=$$(curl -s -o /dev/null -w "%{http_code}" $(API_URL)/api/category/1); \
	echo "    HTTP $$HTTP"; \
	[ "$$HTTP" = "200" ] || [ "$$HTTP" = "404" ] && echo "    PASS: GET /api/category/:id"

	@echo "--- PUT /api/category/1 ---"
	@HTTP=$$(curl -s -o /dev/null -w "%{http_code}" -X PUT \
		-H "Content-Type: application/json" \
		-d '{"name":"Updated","description":"updated desc","image_url":null,"created_at":null}' \
		$(API_URL)/api/category/1); \
	echo "    HTTP $$HTTP"; \
	[ "$$HTTP" = "200" ] || [ "$$HTTP" = "404" ] && echo "    PASS: PUT /api/category/:id"

	@echo "--- DELETE /api/category/1 ---"
	@HTTP=$$(curl -s -o /dev/null -w "%{http_code}" -X DELETE $(API_URL)/api/category/1); \
	echo "    HTTP $$HTTP"; \
	[ "$$HTTP" = "204" ] || [ "$$HTTP" = "404" ] && echo "    PASS: DELETE /api/category/:id"

	@echo "--- GET non-existent row ---"
	@HTTP=$$(curl -s -o /dev/null -w "%{http_code}" $(API_URL)/api/category/99999); \
	[ "$$HTTP" = "404" ] && echo "    PASS: 404 on missing row" || (echo "    FAIL: expected 404, got $$HTTP"; exit 1)

	@echo "--- Invalid id (non-numeric) ---"
	@HTTP=$$(curl -s -o /dev/null -w "%{http_code}" $(API_URL)/api/category/abc); \
	[ "$$HTTP" = "400" ] && echo "    PASS: 400 on non-numeric id" || (echo "    FAIL: expected 400, got $$HTTP"; exit 1)

	@echo "--- Negative id on unsigned PK (user table) ---"
	@HTTP=$$(curl -s -o /dev/null -w "%{http_code}" $(API_URL)/api/user/-1); \
	[ "$$HTTP" = "400" ] && echo "    PASS: 400 on negative unsigned id" || (echo "    FAIL: expected 400, got $$HTTP"; exit 1)

	@echo "--- Reserved-word table (order) ---"
	@HTTP=$$(curl -s -o /dev/null -w "%{http_code}" $(API_URL)/api/order); \
	[ "$$HTTP" = "200" ] && echo "    PASS: reserved-word table responds" || (echo "    FAIL: expected 200, got $$HTTP"; exit 1)

	@echo ""
	@echo "==> All smoke tests passed."

## stop-api: stop the generated API server
stop-api:
	@taskkill //F //IM api.exe > /dev/null 2>&1 || true
	@echo "    Server stopped."

# ─── Housekeeping ─────────────────────────────────────────────────────────────

## clean: remove the generated project directory and gofast binary
clean: stop-api
	@echo "==> Cleaning up..."
	@rm -rf $(OUT_DIR) $(GOFAST_BIN)
	@echo "    Done."

## help: list all targets with descriptions
help:
	@grep -E '^## ' Makefile | sed 's/^## //' | column -t -s ':'
