.PHONY: build test

build::
	mkdir -p .build
	rm -rf .build/*
	go build -o .build/ ./cmd/go-enum-codegen/

test::
	go test ./...