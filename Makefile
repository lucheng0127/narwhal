build_dir=build

all: dir narwhald cover

dir: dir_build dir_cover

dir_build:
	mkdir -p build

dir_cover:
	mkdir -p cover

cover: dir_cover test

test:
	go test -gcflags=-l -coverprofile cover/cover.out ./pkg/...
	go tool cover -html=./cover/cover.out -o cover/cover.html

narwhald:
	go build -o ${build_dir}/narwhald cmd/narwhald/narwhald.go

narwhal:
	go build -o ${build_dir}/narwhal cmd/narwhal/narwhal.go

clean:
	rm -rf ${build_dir}/*
	rm -rf cover