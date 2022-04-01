.PHONY: prebuild fmt

build:
	mkdir -p bin
	go build -o bin/walrss github.com/codemicro/walrss/walrss

run: build
	mkdir -p run
	cd run && ../bin/walrss

fmt:
	go fmt github.com/codemicro/walrss/...
