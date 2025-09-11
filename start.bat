@echo off
REM ============================
REM Go Backend API Startup Script
REM ============================

REM Navigate to project folder
cd "C:\Users\Eng VarTrick\Documents\GitHub\go_backend_api"

REM ----------------------------
REM Set initial PORT
REM ----------------------------
set PORT=2000

REM ----------------------------
REM Start Go go_backend_server with fallback ports
REM ----------------------------
:START_APP
if exist go_backend_server.exe (
    echo Running existing go_backend_server.exe on port %PORT%...
) else (
    echo go_backend_server.exe not found, building...
    go build -o go_backend_server.exe main.go
    if %ERRORLEVEL% NEQ 0 (
        echo Build failed. Exiting...
        pause
        exit /b 1
    )
    echo Build successful. Running go_backend_server.exe on port %PORT%...
)

REM Run the go_backend_server
go_backend_server.exe

REM If go_backend_server.exe exits with error, try next port
if %ERRORLEVEL% NEQ 0 (
    echo go_backend_server failed to start on port %PORT%, trying next port...
    set /a PORT=%PORT%+1
    timeout /t 2 >nul
    goto START_APP
)

pause
