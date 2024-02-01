BINARY_NAME=noob_text

build: 
	@templ generate && go build -o bin/$(BINARY_NAME)

run: build
	@./bin/$(BINARY_NAME)

clean:
	@rm -f bin/$(BINARY_NAME)

tailwind:
	npx tailwindcss -i static/input.css -o static/output.css

css:
	bunx tailwindcss -i static/input.css -o static/output.css --watch