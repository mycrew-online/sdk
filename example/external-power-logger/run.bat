@echo off
echo 🔌 External Power Logger - Build and Run
echo.

REM Check if Flight Simulator is likely running
echo 🔍 Checking for Flight Simulator...
tasklist /fi "imagename eq FlightSimulator.exe" 2>NUL | find /i "FlightSimulator.exe" >NUL
if "%ERRORLEVEL%"=="0" (
    echo ✅ Flight Simulator is running
) else (
    echo ⚠️  Warning: Flight Simulator not detected - please ensure it's running
    echo    Make sure an aircraft is loaded before starting the logger
)
echo.

REM Build the application
echo 🔨 Building external-power-logger...
go build -o external-power-logger.exe .
if errorlevel 1 (
    echo ❌ Build failed
    pause
    exit /b 1
)
echo ✅ Build successful
echo.

REM Run the application
echo 🚀 Starting External Power Logger...
echo    Press Ctrl+C to stop monitoring
echo    Toggle external power in the aircraft to see state changes
echo.
external-power-logger.exe

echo.
echo 👋 External Power Logger stopped
pause
