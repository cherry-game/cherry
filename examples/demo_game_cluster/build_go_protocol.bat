@echo off

echo build go protocol file...
cd internal
cd protocol
protoc --gogo_out=plugins=grpc:../pb/ --gogo_opt=paths=source_relative *.proto
echo build go proto complete!