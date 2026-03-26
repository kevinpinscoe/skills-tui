BINARY := $(HOME)/skills/skill
SRC    := ./

.PHONY: build run test tidy clean

build: tidy
	go build -o $(BINARY) $(SRC)

tidy:
	go mod tidy

run: build
	$(BINARY)

test:
	go test ./...

clean:
	rm -f $(BINARY)
