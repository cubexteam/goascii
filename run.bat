@echo off
title goascii

echo.
echo  goascii - Image to ASCII art converter
echo  by SantianDev ^| https://github.com/cubexteam/goascii
echo  -------------------------------------------------------
echo.

where go >nul 2>&1
if %errorlevel% neq 0 (
    echo  [ERROR] Go not found. Download from https://go.dev/dl/
    pause
    exit /b 1
)

if not exist goascii.exe (
    echo  [INFO] Building goascii.exe...
    go build -o goascii.exe ./cmd/goascii
    if %errorlevel% neq 0 (
        echo  [ERROR] Build failed
        pause
        exit /b 1
    )
    echo  [OK] Done!
    echo.
)

echo  [OK] Opening browser...
echo  [OK] Press Ctrl+C in this window to stop the server
echo.

goascii.exe
