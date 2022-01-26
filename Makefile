all: build

proto_v1:
	cd proto && \
		protoc \
		--go_out=zync/v1 --go_opt=paths=source_relative \
		--go-grpc_out=zync/v1 --go-grpc_opt=paths=source_relative \
		zync.proto

bin:
	mkdir bin

zync: bin proto_v1
	go build -o bin/zync cmd/zync/*.go

zyncd: bin proto_v1
	go build -o bin/zyncd cmd/zyncd/*.go

build: zync zyncd

install: build
	cp bin/zync /usr/local/bin/zync
	cp bin/zyncd /usr/local/bin/zyncd

fmt:
	go fmt ./...
	for file in $$(du -a ./proto | awk '{print $$2}' | grep '\.proto'); do\
		ofile=/tmp/$$(basename $${file}); \
		clang-format --style=file $${file} > $${ofile} && \
		mv $${ofile} $${file}; \
	done

start:
	./bin/zyncd start

stop:
ifneq (, $(shell ls | grep pid))
	kill $(shell cat ./zyncd.pid)
	rm -f zyncd.pid
endif

clean: stop
	rm -f zyncd.*
