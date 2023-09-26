build:
	@go build -o bin/noobtext.exe

run: build
	@bin/noobtext.exe