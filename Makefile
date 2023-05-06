all: maquiaBot4000

maquiaBot4000:
	go build -v

test:
	go test -v ./...

checkFmt:
	[ -z "$$(git ls-files | grep '\.go$$' | xargs gofmt -l)" ] || (exit 1)

clean:
	rm -f maquiaBot4000

.PHONY: all clean test checkFmt

