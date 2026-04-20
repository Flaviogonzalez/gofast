# gofast Makefile
# Full pipeline: build -> generate -> tidy -> compile -> smoke-test -> teardown

# --- Platform detection -------------------------------------------------------
ifeq ($(OS),Windows_NT)
    EXT := .exe
else
    SHELL := sh
    EXT   :=
endif

GOFAST_BIN  := gofast$(EXT)
OUT_DIR     := .testrun
API_BIN     := $(OUT_DIR)/api$(EXT)
API_PORT    := 18082
API_URL     := http://localhost:$(API_PORT)
DATABASE_URL ?= mysql://root:root@localhost:3306/gofast-test

# --- Utilities ----------------------------------------------------------------

.PHONY: all build test generate tidy-api build-api start-api smoke stop-api clean help

## all: run the full pipeline (build -> generate -> tidy -> compile -> smoke)
all: build generate tidy-api build-api smoke

## build: compile the gofast CLI
build:
	@echo ==> Building gofast...
	go build -o $(GOFAST_BIN) ./cmd/gofast
	@echo     OK: $(GOFAST_BIN)

## test: run unit tests
test:
	@echo ==> Unit tests...
	go test ./internal/schema/... ./internal/generator/... -v

ifeq ($(OS),Windows_NT)

## generate: create OUT_DIR and run gofast generate
generate: build
	@echo ==> Generating project in $(OUT_DIR)...
	@if exist "$(OUT_DIR)" rmdir /s /q "$(OUT_DIR)"
	@mkdir "$(OUT_DIR)"
	@echo DATABASE_URL="$(DATABASE_URL)"> "$(OUT_DIR)\.env"
	@cd "$(OUT_DIR)" && "..\$(GOFAST_BIN)" generate

tidy-api:
	@echo ==> go mod tidy (generated project)...
	@cd "$(OUT_DIR)" && go mod tidy

build-api:
	@echo ==> Building generated API...
	@cd "$(OUT_DIR)" && go build -o "api$(EXT)" "./cmd/api"
	@echo     OK: $(API_BIN)

start-api: build-api
	@echo ==> Starting API on port $(API_PORT)...
	@cd "$(OUT_DIR)" && start /b cmd /c "set PORT=$(API_PORT) && api$(EXT) > api.log 2>&1"
	@powershell -NoProfile -NonInteractive -Command "$$up=$$false; for($$i=0;$$i-lt 10;$$i++){try{$$null=Invoke-WebRequest '$(API_URL)/health' -UseBasicParsing -EA Stop;$$up=$$true;break}catch{Start-Sleep 1}}; if(-not $$up){Get-Content '$(OUT_DIR)\api.log'; exit 1}; Write-Host '    Server is up'"

smoke: start-api
	@echo.
	@echo ==> Smoke tests...
	$(MAKE) _smoke_run
	$(MAKE) stop-api

_smoke_run:
	@powershell -NoProfile -NonInteractive -Command "$$f=0; $$c=curl.exe -s -o NUL -w '%{http_code}' $(API_URL)/health; if($$c -eq '200'){Write-Host '    PASS: /health'}else{$$f=1}; $$c=curl.exe -s -o NUL -w '%{http_code}' $(API_URL)/api/category/99999; if($$c -eq '404'){Write-Host '    PASS: 404 on missing row'}else{$$f=1}; if($$f -ne 0){exit 1}; Write-Host '==> All smoke tests passed.'"

stop-api:
	@powershell -NoProfile -NonInteractive -Command "Get-Process -Name 'api' -EA SilentlyContinue | Where-Object { $$_.Path -like '*.testrun*' } | ForEach-Object { $$p = $$_; Stop-Process -Id $$p.Id -Force }; Write-Host '    Server stopped.'"

clean: stop-api
	@echo ==> Cleaning up...
	@if exist "$(OUT_DIR)" rmdir /s /q "$(OUT_DIR)"
	@if exist "$(GOFAST_BIN)" del /f /q "$(GOFAST_BIN)"
	@echo     Done.

else

generate: build
	@echo "==> Generating project in $(OUT_DIR)..."
	@rm -rf $(OUT_DIR)
	@mkdir -p $(OUT_DIR)
	@echo 'DATABASE_URL="$(DATABASE_URL)"' > $(OUT_DIR)/.env
	@cd $(OUT_DIR) && ../$(GOFAST_BIN) generate

tidy-api:
	@echo "==> go mod tidy (generated project)..."
	@cd $(OUT_DIR) && go mod tidy

build-api:
	@echo "==> Building generated API..."
	@cd $(OUT_DIR) && go build -o api$(EXT) ./cmd/api
	@echo "    OK: $(API_BIN)"

start-api: build-api
	@echo "==> Starting API on port $(API_PORT)..."
	@cd $(OUT_DIR) && PORT=$(API_PORT) ./api$(EXT) > api.log 2>&1 &
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		curl -sf $(API_URL)/health > /dev/null 2>&1 && echo "    Server is up" && exit 0; \
		sleep 1; \
	done; \
	echo "ERROR: server did not start within 10 s"; \
	cat $(OUT_DIR)/api.log; \
	exit 1

smoke: start-api
	@echo ""
	@echo "==> Smoke tests..."
	$(MAKE) _smoke_run
	@$(MAKE) stop-api

_smoke_run:
	@echo "--- /health ---"
	@curl -sf $(API_URL)/health | grep -q "OK" && echo "    PASS: /health"
	@HTTP=$$(curl -s -o /dev/null -w "%{http_code}" $(API_URL)/api/category/99999); \
	[ "$$HTTP" = "404" ] && echo "    PASS: 404 on missing row" || (echo "    FAIL"; exit 1)
	@HTTP=$$(curl -s -o /dev/null -w "%{http_code}" $(API_URL)/api/category/abc); \
	[ "$$HTTP" = "400" ] && echo "    PASS: 400 on non-numeric id" || (echo "    FAIL"; exit 1)
	@echo "==> All smoke tests passed."

stop-api:
	@pkill -f "$(OUT_DIR)/api" > /dev/null 2>&1 || true
	@echo "    Server stopped."

clean: stop-api
	@echo "==> Cleaning up..."
	@rm -rf $(OUT_DIR) $(GOFAST_BIN)
	@echo "    Done."

endif

## help: list all targets with descriptions
help:
	@grep -E '^## ' Makefile | sed 's/^## //' | column -t -s ':'
