@echo off

del /f ws2cgi_x64.exe 1>nul 2>nul

del /f ws2cgi_x64.elf 1>nul 2>nul

set GOOS=windows
set GOARCH=amd64
go build -v -o ws2cgi_x64.exe -ldflags "-s -w -buildid=" -trimpath -race ./
if not %errorlevel% == 0 (
echo.
echo program exited with code %errorlevel%.
echo press any key to exit.
pause>nul
)

set GOOS=linux
set GOARCH=amd64
go build -v -o ws2cgi_x64.elf -ldflags "-s -w -buildid=" -trimpath ./

if not %errorlevel% == 0 (
echo.
echo program exited with code %errorlevel%.
echo press any key to exit.
pause>nul
)