@echo off
setlocal
cd /d "%~dp0"
where go >nul 2>nul || (
  echo [ERROR] Go 1.23+ nije pronaden u PATH-u.
  pause
  exit /b 1
)
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0native\build_windows.ps1"
if errorlevel 1 (
  echo.
  echo [ERROR] Build nije uspio.
  pause
  exit /b 1
)
echo.
echo Build je gotov. Pogledaj mapu release.
pause
