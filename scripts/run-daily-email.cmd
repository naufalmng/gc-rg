@echo off
setlocal

set APP_DIR=%GC_RG_WORKDIR%
if "%APP_DIR%"=="" set APP_DIR=C:\gc-rg
set REPORT_DIR=%GC_RG_REPORT_DIR%
if "%REPORT_DIR%"=="" set REPORT_DIR=%APP_DIR%\reports\daily
set REPORT_DATE=%GC_RG_DATE%
if "%REPORT_DATE%"=="" for /f %%i in ('powershell -NoProfile -Command "Get-Date -Format yyyy-MM-dd"') do set REPORT_DATE=%%i

cd /d "%APP_DIR%" || exit /b 1
"%APP_DIR%\bin\gc-rg-generate.exe" --date "%REPORT_DATE%" || exit /b 1
"%APP_DIR%\bin\gc-rg-email.exe" --date "%REPORT_DATE%" --report-dir "%REPORT_DIR%" --send || exit /b 1
