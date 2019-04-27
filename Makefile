install:
	go install

lint:
	golangci-lint run

update:
	@rm -f -- go.sum
	@go mod tidy
	go get -u
	@go build -o $(TMPDIR)/main
	@git diff -- go.mod || :
