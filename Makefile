ifndef WASM_EXEC_PATH
WASM_EXEC_PATH="$(shell go env GOROOT)/misc/wasm/wasm_exec.js"
endif

ifndef ITCH_USERNAME
ITCH_USERNAME=eliasdaler
endif

PROJ_NAME=crten
ITCH_PATH=crten

## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## serve: serve WebAssembly version locally
.PHONY: serve
serve:
	@echo "Hosting game on http://localhost:4242"
	wasmserve -http=":4242" -allow-origin='*' -tags main.go

## prod/linux: build prod Linux version
.PHONY: prod/linux
prod/linux:
	@echo "Making prod Linux build..."
	rm -rf prod/linux
	mkdir -p prod/linux/$(PROJ_NAME)
	go build -tags prod -o prod/linux/$(PROJ_NAME)/$(PROJ_NAME) .

## prod/win32: build prod Win32 version
.PHONY: prod/win32
prod/win32:
	@echo "Making prod Win32 build..."
	rm -rf prod/win32
	mkdir -p prod/win32/$(PROJ_NAME)
	GOOS=windows go build -tags prod -o prod/win32/$(PROJ_NAME)/$(PROJ_NAME).exe .

## prod/web: build prod WebAssembly version
.PHONY: prod/web
prod/web:
	@echo "Making prod wasm build..."
	rm -rf prod/web
	mkdir -p prod/web
	GOOS=js GOARCH=wasm go build -tags "prod,js" -o prod/web/game.wasm .
	cp -r html/* prod/web
	cp $(WASM_EXEC_PATH) prod/web

## prod: build all prod versions
.PHONY: prod
prod: prod/linux prod/win32 prod/web

## itch: upload all prod versions on itch.io
.PHONY: itch
itch: clean/prod prod
	butler push --if-changed prod/win32 $(ITCH_USERNAME)/$(ITCH_PATH):windows
	butler push --if-changed prod/linux $(ITCH_USERNAME)/$(ITCH_PATH):linux-amd64
	butler push prod/web $(ITCH_USERNAME)/$(ITCH_PATH):web
	@echo "Project is live on http://$(ITCH_USERNAME).itch.io/$(ITCH_PATH)"

## run: run game (dev)
.PHONY: run
run:
	go run .

## clean/prod: remove all previosly built prod versions
.PHONY: clean/prod
clean/prod:
	rm -rf prod

## clean: clean all build artifacts
.PHONY: clean
clean: clean_prod
	rm -rf game.wasm site.zip

