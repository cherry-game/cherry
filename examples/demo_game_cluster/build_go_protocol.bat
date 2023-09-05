@echo off

echo build go protocol file...
cd internal
cd protocol
protoc --go_out=plugins=grpc:../pb/ --go_opt=paths=source_relative *.proto
echo build go proto complete!