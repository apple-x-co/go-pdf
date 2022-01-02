BIN := 'go-pdf'

VERSION := '0.0.5'
REVISION := '$(shell git rev-parse --short HEAD)'

BUILD_TAGS_PRODUCTION := 'production'
BUILD_TAGS_DEVELOPMENT := 'development unittest'

build-all: clean dev-mac linux

.PHONY: version
version:
	echo $(VERSION).$(REVISION)

.PHONY: build
build:
	go build -o $(BIN_NAME) -tags '$(BUILD_TAGS) netgo' -installsuffix netgo -ldflags '-s -w -X main.version=$(VERSION) -X main.revision=$(REVISION) -extldflags "-static"' main.go

.PHONY: run
run:
	go run ./main.go --in samples/sample-delivery-note/layout.json --out samples/sample-delivery-note/output.pdf --ttf fonts/TakaoPGothic.ttf
	#go run ./main.go --in samples/sample-report1/layout.json --out samples/sample-report1/output.pdf --ttf fonts/TakaoPGothic.ttf
	#go run ./main.go --in samples/sample-report2/layout.json --out samples/sample-report2/output.pdf --ttf fonts/TakaoPGothic.ttf
	#go run ./main.go --in samples/text-wrap2/layout.json --out samples/text-wrap2/output.pdf --ttf fonts/TakaoPGothic.ttf

.PHONY: dev-mac
dev-mac:
	go mod tidy
	go fmt
	if [ ! -d bin ]; then mkdir bin; fi
	$(MAKE) build BUILD_TAGS=$(BUILD_TAGS_DEVELOPMENT) BIN_NAME=bin/$(BIN)-dev-mac

.PHONY: linux
linux:
	if [ ! -d bin ]; then mkdir bin; fi
	$(MAKE) build BUILD_TAGS=$(BUILD_TAGS_PRODUCTION) CGO_ENABLED=0 GOOS=linux GOARCH=amd64 BIN_NAME=bin/$(BIN)-linux64

.PHONY: clean
clean:
	rm -rf bin/*
	go clean

.PHONY: ci-test
ci-test:
	if [ ! -d work ]; then mkdir work; fi
	./bin/$(BIN)-dev-mac --in layout.json --out work/output.pdf --ttf fonts/TakaoPGothic.ttf

.PHONY: exec-samples
exec-samples:
	./bin/$(BIN)-dev-mac --in samples/commpress-level/layout.json --out samples/commpress-level/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/header-footer/layout.json --out samples/header-footer/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/header-footer-layoutconstant/layout.json --out samples/header-footer-layoutconstant/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/image/layout.json --out samples/image/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/image-border/layout.json --out samples/image-border/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/image-margin/layout.json --out samples/image-margin/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/image-origin/layout.json --out samples/image-origin/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/image-resize/layout.json --out samples/image-resize/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/image-size/layout.json --out samples/image-size/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/image-template/layout.json --out samples/image-template/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/layout-layoutconstant/layout.json --out samples/layout-layoutconstant/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/layout-orientation/layout.json --out samples/layout-orientation/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/page-header-footer/layout.json --out samples/page-header-footer/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/password-protect/layout.json --out samples/password-protect/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/sample-delivery-note/layout.json --out samples/sample-delivery-note/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/sample-report1/layout.json --out samples/sample-report1/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/sample-report2/layout.json --out samples/sample-report2/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/text/layout.json --out samples/text/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/text-align/layout.json --out samples/text-align/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/text-backgroundcolor/layout.json --out samples/text-backgroundcolor/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/text-border/layout.json --out samples/text-border/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/text-color/layout.json --out samples/text-color/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/text-linebreak/layout.json --out samples/text-linebreak/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/text-margin/layout.json --out samples/text-margin/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/text-origin/layout.json --out samples/text-origin/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/text-size/layout.json --out samples/text-size/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/text-template/layout.json --out samples/text-template/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/text-textsize/layout.json --out samples/text-textsize/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/text-wrap/layout.json --out samples/text-wrap/output.pdf --ttf fonts/TakaoPGothic.ttf
	./bin/$(BIN)-dev-mac --in samples/text-wrap2/layout.json --out samples/text-wrap2/output.pdf --ttf fonts/TakaoPGothic.ttf