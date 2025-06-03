@echo off
setlocal

echo 🌤️  Weather WebService Demo - Starting...
echo.

REM Check if Flight Simulator is likely running
echo 🔍 Checking for Flight Simulator...
tasklist /fi "imagename eq FlightSimulator.exe" 2>NUL | find /i "FlightSimulator.exe" >NUL
if "%ERRORLEVEL%"=="0" (
    echo ✅ Flight Simulator is running
) else (
    echo ⚠️  Warning: Flight Simulator not detected
    echo    Please ensure MSFS is running with an aircraft loaded
)
echo.

echo 🚀 Building and starting weather web service...
echo    Open your browser to: http://localhost:8080
echo    Press Ctrl+C to stop the server
echo.

go run main.go

pause
