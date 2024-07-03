BINARY_NAME=tasca

all: build

build:
	go build -o ${BINARY_NAME} main.go

run:
	go run .

clean:
	go clean
	rm ${BINARY_NAME}
