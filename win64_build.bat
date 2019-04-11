@echo off
set GOARCH=amd64
set PATH=C:\mingw81_64\bin;%PATH%
echo Building, please wait...
go build -i -v -o imdbws64.exe
pause