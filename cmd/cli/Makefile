all:
	go build -o main main.go

windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o main main.go

linux:
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go

test:
	go test -v .