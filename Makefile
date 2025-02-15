.PHONY: make assets
COVERAGE = coverage.out

make:
	go run main.go

publish:
	date
	go mod tidy
	git push heroku main

tailwind:
	./tailwindcss -i ./static/tailwind.css -o ./static/tailwind.min.css --minify

test: 
	go test ./...

test_coverage:
	go test ./... -v -coverprofile=$(COVERAGE)
	go tool cover -html=$(COVERAGE) 