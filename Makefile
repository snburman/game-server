.PHONY: make assets
COVERAGE = coverage.out

make:
	go run main.go

publish:
	date
	git push heroku main

assets:
	go run cmd/generate/main.go

tailwind:
	./tailwindcss -i ./static/tailwind.css -o ./static/tailwind.min.css --minify

test: 
	go test ./...

test_coverage:
	go test ./... -v -coverprofile=$(COVERAGE)
	go tool cover -html=$(COVERAGE) 