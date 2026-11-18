@echo off
chcp 65001 >nul
cd /d "%~dp0"
"%~dp0非遗华裳.exe" %*
if errorlevel 1 pause
