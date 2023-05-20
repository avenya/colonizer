build: install
	go build  -o ../terraform/colonizer .

install:
	go mod tidy
	go install

