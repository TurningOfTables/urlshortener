test:
	go test

test-coverage:
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out

run-test:
	go run main.go database.go -localhost -mode test

run:
	go run main.go database.go -mode production

build:
	go build -o /bin