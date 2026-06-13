export CGO_ENABLED := "0"

default: build

build:
    go build .

run *ARGS:
    go run . {{ARGS}}
