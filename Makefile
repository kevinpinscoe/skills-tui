BINARY      := ./skill
INSTALL_DIR := $(HOME)/.local/bin
SRC         := ./
VERSION     := $(shell git describe --tags --dirty --always 2>/dev/null || echo dev)
LDFLAGS     := -X main.version=$(VERSION)

.PHONY: build install run test tidy clean

build: tidy
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) $(SRC)

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/skill


tidy:
	go mod tidy

run: build
	$(BINARY)

test:
	go test ./...

clean:
	rm -f $(BINARY)
