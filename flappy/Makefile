.PHONY=build
build:
	env GOOS=js GOARCH=wasm go build -o main.wasm .

.PHONY=serve
serve:
	go run github.com/mattn/serve@latest -a ':5123' .
