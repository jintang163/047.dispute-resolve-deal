$ErrorActionPreference = "Continue"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Dispute Resolve - Frontend Launcher    " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

function Start-Frontend {
    param(
        [string]$Name,
        [string]$Dir,
        [string]$Port,
        [string]$Color
    )

    $FullPath = Join-Path $ScriptDir $Dir

    if (-not (Test-Path $FullPath)) {
        Write-Host "[ERROR] $Name directory not found: $FullPath" -ForegroundColor Red
        return $null
    }

    Write-Host "[$Name] Starting on port $Port..." -ForegroundColor $Color

    $LogFile = Join-Path $ScriptDir "$Name.log"
    "" | Out-File -FilePath $LogFile -Encoding UTF8

    $Process = Start-Process -FilePath "powershell.exe" `
        -ArgumentList "-NoExit", "-Command", "
            Set-Location '$FullPath'
            Write-Host '[$Name] Installing dependencies...' -ForegroundColor $Color
            npm install
            Write-Host '[$Name] Starting dev server on port $Port...' -ForegroundColor $Color
            if ('$Name' -eq 'kiosk') {
                npm run dev -- --port $Port
            } else {
                npm run dev
            }
        " `
        -PassThru `
        -RedirectStandardOutput $LogFile `
        -RedirectStandardError $LogFile

    Start-Sleep -Seconds 3

    if ($Process.HasExited) {
        Write-Host "[$Name] Failed to start. Check $LogFile for details." -ForegroundColor Red
        return $null
    }

    Write-Host "[$Name] Started (PID: $($Process.Id), Log: $LogFile)" -ForegroundColor Green
    return $Process
}

Write-Host "[1/3] Starting Admin Frontend..." -ForegroundColor Yellow
$AdminProcess = Start-Frontend -Name "admin" -Dir "admin" -Port "5173" -Color "Cyan"

Write-Host ""
Write-Host "[2/3] Starting Kiosk Frontend..." -ForegroundColor Yellow
$KioskProcess = Start-Frontend -Name "kiosk" -Dir "kiosk" -Port "5174" -Color "Magenta"

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Frontend Services Status               " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

if ($null -ne $AdminProcess) {
    Write-Host "[OK] Admin UI    : http://localhost:5173" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Admin UI failed to start" -ForegroundColor Red
}

if ($null -ne $KioskProcess) {
    Write-Host "[OK] Kiosk UI    : http://localhost:5174" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Kiosk UI failed to start" -ForegroundColor Red
}

Write-Host ""
Write-Host "[INFO] MiniApp: Please use Kuikly CLI to compile and run" -ForegroundColor Yellow
Write-Host "       cd miniapp" -ForegroundColor DarkGray
Write-Host "       kuikly dev --platform wechat     (WeChat MiniApp)" -ForegroundColor DarkGray
Write-Host "       kuikly dev --platform alipay     (Alipay MiniApp)" -ForegroundColor DarkGray
Write-Host "       kuikly build --platform harmonyos (HarmonyOS)" -ForegroundColor DarkGray
Write-Host ""
Write-Host "Press Ctrl+C to stop all services..." -ForegroundColor DarkGray

try {
    while ($true) {
        Start-Sleep -Seconds 1

        if (($null -ne $AdminProcess -and $AdminProcess.HasExited) -or
            ($null -ne $KioskProcess -and $KioskProcess.HasExited)) {
            Write-Host ""
            Write-Host "[WARN] One or more services have stopped." -ForegroundColor Yellow
            break
        }
    }
} finally {
    Write-Host ""
    Write-Host "[SHUTDOWN] Stopping all frontend services..." -ForegroundColor Yellow

    if ($null -ne $AdminProcess -and -not $AdminProcess.HasExited) {
        Stop-Process -Id $AdminProcess.Id -Force -ErrorAction SilentlyContinue
        Write-Host "[SHUTDOWN] Admin UI stopped" -ForegroundColor Gray
    }

    if ($null -ne $KioskProcess -and -not $KioskProcess.HasExited) {
        Stop-Process -Id $KioskProcess.Id -Force -ErrorAction SilentlyContinue
        Write-Host "[SHUTDOWN] Kiosk UI stopped" -ForegroundColor Gray
    }

    Get-Process -Name node -ErrorAction SilentlyContinue | Where-Object {
        $_.MainWindowTitle -like "*admin*" -or $_.MainWindowTitle -like "*kiosk*"
    } | Stop-Process -Force -ErrorAction SilentlyContinue

    Write-Host ""
    Write-Host "[SHUTDOWN] All frontend services stopped." -ForegroundColor Gray
}
