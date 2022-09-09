@echo off

echo build client protocol file...
cd internal
cd protocol
protoc --gogo_out=plugins=grpc:../pb/ --gogo_opt=paths=source_relative *.proto
echo build client proto complete!


echo all task is finished!