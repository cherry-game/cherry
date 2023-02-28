@echo off

if exist outjs (
  rd /s /q outjs
)

md outjs

protoc --proto_path=internal/protocol/ --js_out=import_style=commonjs,binary:outjs/ internal/protocol/*.proto


set oujsDir="%cd%"\outjs
setlocal enableextensions enabledelayedexpansion
set alljs=
for /R %oujsDir% %%f in (*.js) do (
  set alljs=!alljs! %%f
)
set "alljs=%alljs:~1%"


echo %alljs%
browserify %alljs% --outfile nodes/web/static/pb.js