BINARY_NAME=noob_text

build: 
	@templ generate && go build -o bin/$(BINARY_NAME)

run: build
	@./bin/$(BINARY_NAME)

clean:
	@rm -f bin/$(BINARY_NAME)

tailwind:
	npx tailwindcss -i public/input.css -o public/output.css

css:
	bunx tailwindcss -i public/input.css -o public/output.css --watch
