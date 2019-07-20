TARGETS := ddns ddns-tool
CGO := 0
GO111MODULE := on

build: $(TARGETS)

.PHONY: ddns
ddns:
	go build -v -o bin/ddns cmd/ddns/main.go

.PHONY: ddns-tool
ddns-tool:
	go build -v -o bin/ddns-tool cmd/ddns-tool/main.go


clean:
	rm -rf bin/*
