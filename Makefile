.PHONY: make assets

make:
	go run main.go

assets:
	go run cmd/generate/main.go

tailwind:
	./tailwindcss -i ./static/tailwind.css -o ./static/tailwind.min.css --minify