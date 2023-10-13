BIN := bluecare

GOPATH := $(shell go env GOPATH)
GO_FILES := $(shell find . -name "*.go")
GO_DEPS := $(shell find . -name go.mod -o -name go.sum)

$(BIN): $(GO_FILES) $(GO_DEPS)
	$(MAKE) pretty
	go build -o $(BIN) main.go

test: $(BIN)
	./$(BIN)
.PHONY: test

pretty: $(GO_FILES)
	gofumpt -w $^
.PHONY: pretty

install: $(GOPATH)/bin/$(BIN)
.PHONY: install

$(GOPATH)/bin/$(BIN): $(BIN)
	mv $(BIN) $(GOPATH)/bin/$(BIN)

clean:
	rm -f $(BIN)
.PHONY: clean
