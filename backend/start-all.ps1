$ErrorActionPreference = "Stop"

$BackendDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ServicesDir = Join-Path $BackendDir "services"
$GatewayDir = Join-Path $BackendDir "gateway"

$Services = @(
    @{ Name = "user-service"; Port = 9001; Path = Join-Path $ServicesDir "user-service" },
    @{ Name = "dispute-service"; Port = 9002; Path = Join-Path $ServicesDir "dispute-service" },
    @{ Name = "workflow-service"; Port = 9003; Path = Join-Path $ServicesDir "workflow-service" },
    @{ Name = "ai-service"; Port = 9004; Path = Join-Path $ServicesDir "ai-service" },
    @{ Name = "notification-service"; Port = 9005; Path = Join-Path $ServicesDir "notification-service" }
)

$Processes = @()

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  综治中心矛盾纠纷管理系统 - 后端启动" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

function Stop-ExistingProcess {
    param($Port)
    $process = netstat -ano | findstr ":$Port" | findstr "LISTENING"
    if ($process) {
        $pid = ($process -split '\s+')[-1]
        Write-Host "  端口 $Port 被占用，正在终止进程 PID: $pid" -ForegroundColor Yellow
        Stop-Process -Id $pid -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 1
    }
}

Write-Host "[1/6] 检查并清理端口..." -ForegroundColor Green
foreach ($svc in $Services) {
    Stop-ExistingProcess $svc.Port
}
Stop-ExistingProcess 8080
Write-Host "  端口清理完成" -ForegroundColor Green
Write-Host ""

Write-Host "[2/6] 启动微服务..." -ForegroundColor Green
foreach ($svc in $Services) {
    Write-Host "  正在启动 $($svc.Name) (端口: $($svc.Port))..." -ForegroundColor Yellow
    
    $process = Start-Process -FilePath "go" `
        -ArgumentList "run", "main.go" `
        -WorkingDirectory $svc.Path `
        -PassThru `
        -NoNewWindow `
        -RedirectStandardOutput (Join-Path $BackendDir "logs\$($svc.Name).log") `
        -RedirectStandardError (Join-Path $BackendDir "logs\$($svc.Name)-err.log")
    
    $Processes += @{ Name = $svc.Name; Process = $process; Port = $svc.Port }
    Start-Sleep -Seconds 2
    
    if ($process.HasExited) {
        Write-Host "  ✗ $($svc.Name) 启动失败" -ForegroundColor Red
        exit 1
    }
    Write-Host "  ✓ $($svc.Name) 启动成功，PID: $($process.Id)" -ForegroundColor Green
}
Write-Host ""

Write-Host "[3/6] 启动 API Gateway (端口: 8080)..." -ForegroundColor Green
$gatewayLog = Join-Path $BackendDir "logs\gateway.log"
$gatewayErrLog = Join-Path $BackendDir "logs\gateway-err.log"
$gatewayProcess = Start-Process -FilePath "go" `
    -ArgumentList "run", "main.go" `
    -WorkingDirectory $GatewayDir `
    -PassThru `
    -NoNewWindow `
    -RedirectStandardOutput $gatewayLog `
    -RedirectStandardError $gatewayErrLog

$Processes += @{ Name = "gateway"; Process = $gatewayProcess; Port = 8080 }
Start-Sleep -Seconds 3

if ($gatewayProcess.HasExited) {
    Write-Host "  ✗ Gateway 启动失败" -ForegroundColor Red
    exit 1
}
Write-Host "  ✓ Gateway 启动成功，PID: $($gatewayProcess.Id)" -ForegroundColor Green
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  所有服务启动成功！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "  服务列表：" -ForegroundColor White
foreach ($p in $Processes) {
    Write-Host "    • $($p.Name) - 端口: $($p.Port), PID: $($p.Process.Id)" -ForegroundColor Gray
}
Write-Host ""
Write-Host "  API 地址: http://localhost:8080" -ForegroundColor Cyan
Write-Host ""
Write-Host "  日志目录: $BackendDir\logs\" -ForegroundColor Gray
Write-Host ""
Write-Host "  按 Ctrl+C 停止所有服务" -ForegroundColor Yellow
Write-Host ""

try {
    while ($true) {
        foreach ($p in $Processes) {
            if ($p.Process.HasExited) {
                Write-Host "  ⚠ $($p.Name) 已意外退出，退出码: $($p.Process.ExitCode)" -ForegroundColor Red
            }
        }
        Start-Sleep -Seconds 5
    }
}
finally {
    Write-Host ""
    Write-Host "正在停止所有服务..." -ForegroundColor Yellow
    foreach ($p in $Processes) {
        if (-not $p.Process.HasExited) {
            Stop-Process -Id $p.Process.Id -Force -ErrorAction SilentlyContinue
            Write-Host "  ✓ 已停止 $($p.Name)" -ForegroundColor Green
        }
    }
    Write-Host "所有服务已停止" -ForegroundColor Green
}
