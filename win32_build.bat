@echo off
set GOROOT=C:\Go32
set GOARCH=386
set PATH=C:\Go32\bin;C:\mingw81_32\bin;%PATH%
echo Building, please wait...
go build -i -v -ldflags="-s -w" -o imdbws32.exe
pause