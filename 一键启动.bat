@echo off
chcp 65001 >nul
if exist "%~dp0release\windows-amd64\非遗华裳.exe" (
  call "%~dp0release\windows-amd64\一键启动.bat"
) else (
  call "%~dp0开发启动.bat"
)
exit /b %errorlevel%
