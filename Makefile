.PHONY: dockerBuild

build: templates
	mkdir -p bin
	go build -o bin/walrss github.com/codemicro/walrss/walrss

run: build
	mkdir -p run
	cd run && ../bin/walrss

fmt:
	go fmt github.com/codemicro/walrss/...

templates:
	qtc -skipLineComments -ext qtpl.html -dir walrss/internal/http/views

