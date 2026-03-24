all: templ
	@go run main.go

air: templ
	@go build -o ./tmp/main .
	
templ:
	@templ generate ./views/...
