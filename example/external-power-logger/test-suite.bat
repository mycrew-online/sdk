@echo off
setlocal

echo 🔌 External Power Logger - Test Suite
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

REM Show menu
echo 📋 Choose test mode:
echo    1. Standard Logger (VISUAL_FRAME period)
echo    2. Period Test - VISUAL_FRAME 
echo    3. Period Test - SECOND
echo    4. Period Test - ON_SET
echo    5. Build only
echo    6. Exit
echo.

set /p choice="Enter your choice (1-6): "

if "%choice%"=="1" goto standard
if "%choice%"=="2" goto visual_frame
if "%choice%"=="3" goto second
if "%choice%"=="4" goto on_set
if "%choice%"=="5" goto build_only
if "%choice%"=="6" goto exit

echo ❌ Invalid choice
pause
exit /b 1

:build_only
echo 🔨 Building applications...
go build -o external-power-logger.exe main.go
if errorlevel 1 (
    echo ❌ Build failed for main.go
    pause
    exit /b 1
)
echo ✅ Built external-power-logger.exe

go build -o period-test.exe period-test.go
if errorlevel 1 (
    echo ❌ Build failed for period-test.go
    pause
    exit /b 1
)
echo ✅ Built period-test.exe
echo.
echo ✅ All builds successful
pause
exit /b 0

:standard
echo 🔨 Building standard logger...
go build -o external-power-logger.exe main.go
if errorlevel 1 (
    echo ❌ Build failed
    pause
    exit /b 1
)
echo ✅ Build successful
echo.
echo 🚀 Starting Standard External Power Logger...
echo    Uses VISUAL_FRAME period for maximum responsiveness
echo    Press Ctrl+C to stop monitoring
echo.
external-power-logger.exe
goto done

:visual_frame
echo 🔨 Building period test...
go build -o period-test.exe period-test.go
if errorlevel 1 (
    echo ❌ Build failed
    pause
    exit /b 1
)
echo ✅ Build successful
echo.
echo 🚀 Starting Period Test - VISUAL_FRAME...
echo    High frequency updates (30-60 per second)
echo    Press Ctrl+C to stop monitoring
echo.
period-test.exe visual_frame
goto done

:second
echo 🔨 Building period test...
go build -o period-test.exe period-test.go
if errorlevel 1 (
    echo ❌ Build failed
    pause
    exit /b 1
)
echo ✅ Build successful
echo.
echo 🚀 Starting Period Test - SECOND...
echo    Updates once per second
echo    Press Ctrl+C to stop monitoring
echo.
period-test.exe second
goto done

:on_set
echo 🔨 Building period test...
go build -o period-test.exe period-test.go
if errorlevel 1 (
    echo ❌ Build failed
    pause
    exit /b 1
)
echo ✅ Build successful
echo.
echo 🚀 Starting Period Test - ON_SET...
echo    Updates only when external power state changes
echo    Press Ctrl+C to stop monitoring
echo.
period-test.exe on_set
goto done

:done
echo.
echo 👋 External Power Logger stopped

:exit
pause
