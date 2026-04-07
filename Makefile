BINARY      := ./skill
INSTALL_DIR := $(HOME)/.local/bin
SRC         := ./

.PHONY: build install run test tidy clean

build: tidy
	go build -o $(BINARY) $(SRC)

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
