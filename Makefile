deps:
	# install all dependencies required for running application
	go version
	go env

	# installing golang dependencies using golang modules
	go mod download # ensure dependencies are present
	go mod verify # ensure dependencies are present
	go mod tidy # ensure go.mod is sane

lint:
	gofmt  -w=true -s=true -l=true ./
	golint ./...
	go vet ./...

# https://go.dev/blog/govulncheck
# install it by go install golang.org/x/vuln/cmd/govulncheck@latest
vuln:
	which govulncheck
	govulncheck ./...

check: lint
	go test -v -coverprofile=cover.out ./...

test: check

start:
	go run example/cmd/main.go
