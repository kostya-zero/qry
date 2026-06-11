export CGO_ENABLED := "0"

default: run

run *args:
    go run . {{args}}

build:
    go build .
