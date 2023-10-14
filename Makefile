default: help

help:
	@echo "COMMAND:"
	@echo "	make init"
	@echo "	make protoc"
	@echo "	make git-tag"

init:
	@echo "[INIT] install protoc-gen-go"
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

protoc:
	@echo "[PROTOC] build proto files"
	cd net && \
	cd proto && \
	protoc --go_out=. --go_opt=paths=source_relative proto.proto

tag:
	./tag.sh
