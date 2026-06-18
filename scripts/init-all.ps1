# =====================================================
# 综治中心矛盾纠纷管理系统 - 一键初始化脚本 (Windows PowerShell)
# 功能：创建目录、安装前端依赖、检查配置、启动提示
# =====================================================

$ErrorActionPreference = "Continue"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir

Write-Host ""
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "  综治中心矛盾纠纷管理系统 - 初始化脚本" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "[1/6] 创建后端日志目录..." -ForegroundColor Yellow
$LogDir = Join-Path $ProjectRoot "backend\logs"
if (-not (Test-Path $LogDir)) {
    New-Item -ItemType Directory -Path $LogDir -Force | Out-Null
    Write-Host "    创建目录: $LogDir" -ForegroundColor Green
} else {
    Write-Host "    目录已存在: $LogDir" -ForegroundColor Gray
}
Write-Host ""

Write-Host "[2/6] 安装管理端前端依赖..." -ForegroundColor Yellow
$AdminDir = Join-Path $ProjectRoot "frontend\admin"
if (Test-Path $AdminDir) {
    Write-Host "    进入目录: $AdminDir" -ForegroundColor Gray
    Push-Location $AdminDir
    try {
        npm install
        if ($LASTEXITCODE -eq 0) {
            Write-Host "    管理端依赖安装成功" -ForegroundColor Green
        } else {
            Write-Warning "    管理端依赖安装失败，请手动执行 npm install"
        }
    } finally {
        Pop-Location
    }
} else {
    Write-Warning "    管理端目录不存在: $AdminDir"
}
Write-Host ""

Write-Host "[3/6] 安装自助终端前端依赖..." -ForegroundColor Yellow
$KioskDir = Join-Path $ProjectRoot "frontend\kiosk"
if (Test-Path $KioskDir) {
    Write-Host "    进入目录: $KioskDir" -ForegroundColor Gray
    Push-Location $KioskDir
    try {
        npm install
        if ($LASTEXITCODE -eq 0) {
            Write-Host "    自助终端依赖安装成功" -ForegroundColor Green
        } else {
            Write-Warning "    自助终端依赖安装失败，请手动执行 npm install"
        }
    } finally {
        Pop-Location
    }
} else {
    Write-Warning "    自助终端目录不存在: $KioskDir"
}
Write-Host ""

Write-Host "[4/6] 检查数据库配置..." -ForegroundColor Yellow
$ConfigFiles = @(
    (Join-Path $ProjectRoot "backend\gateway\conf\config.yaml"),
    (Join-Path $ProjectRoot "backend\config\config.yaml")
)
foreach ($ConfigFile in $ConfigFiles) {
    if (Test-Path $ConfigFile) {
        Write-Host ""
        Write-Host "    配置文件: $ConfigFile" -ForegroundColor Gray
        Write-Host ""
        Write-Host "    ================================================" -ForegroundColor Yellow
        Write-Host "    ⚠  请修改以下数据库配置（如未修改）：" -ForegroundColor Yellow
        Write-Host "    ================================================" -ForegroundColor Yellow
        Write-Host "      database:"
        Write-Host "        host:     127.0.0.1   ← 修改为你的MySQL/TiDB地址"
        Write-Host "        port:     4000        ← 修改为端口(MySQL=3306, TiDB=4000)"
        Write-Host "        user:     root        ← 修改为数据库用户名"
        Write-Host "        password: 123456      ← 修改为数据库密码"
        Write-Host "        dbname:   dispute_resolve"
        Write-Host ""
        Write-Host "      redis:"
        Write-Host "        host:     127.0.0.1   ← 修改为Redis地址"
        Write-Host "        port:     6379        ← 修改为Redis端口"
        Write-Host ""
        Write-Host "      rocketmq:"
        Write-Host "        nameserver:"
        Write-Host "          - 127.0.0.1:9876    ← 修改为RocketMQ地址"
        Write-Host ""
        break
    }
}
Write-Host ""

Write-Host "[5/6] 数据库初始化步骤..." -ForegroundColor Yellow
Write-Host ""
Write-Host "    执行顺序（重要！）：" -ForegroundColor Cyan
Write-Host ""
Write-Host "    步骤1: 先执行建表脚本" -ForegroundColor White
Write-Host "      mysql -h主机 -u用户 -p密码 < sql\01_init_schema.sql" -ForegroundColor Gray
Write-Host "      或:  mysql -h主机 -u用户 -p密码 -e 'source $ProjectRoot\sql\01_init_schema.sql'" -ForegroundColor Gray
Write-Host ""
Write-Host "    步骤2: 再执行初始化数据脚本" -ForegroundColor White
Write-Host "      mysql -h主机 -u用户 -p密码 < sql\init-data.sql" -ForegroundColor Gray
Write-Host "      或:  mysql -h主机 -u用户 -p密码 -e 'source $ProjectRoot\sql\init-data.sql'" -ForegroundColor Gray
Write-Host ""
Write-Host "    数据库: dispute_resolve" -ForegroundColor Gray
Write-Host "    字符集: utf8mb4" -ForegroundColor Gray
Write-Host ""

Write-Host "[6/6] 启动顺序说明..." -ForegroundColor Yellow
Write-Host ""
Write-Host "    ================================================" -ForegroundColor Cyan
Write-Host "    启动后端服务" -ForegroundColor Cyan
Write-Host "    ================================================" -ForegroundColor Cyan
Write-Host "      cd $ProjectRoot\backend"
Write-Host "      .\start-all.ps1"
Write-Host ""
Write-Host "    服务列表："
Write-Host "      - gateway           :8080  (HTTP网关)"
Write-Host "      - user-service      :8081  (用户服务)"
Write-Host "      - dispute-service   :8082  (纠纷服务)"
Write-Host "      - workflow-service  :8083  (工作流服务)"
Write-Host "      - ai-service        :8084  (AI服务)"
Write-Host "      - notification-service :8085 (通知服务)"
Write-Host ""
Write-Host "    ================================================" -ForegroundColor Cyan
Write-Host "    启动前端服务" -ForegroundColor Cyan
Write-Host "    ================================================" -ForegroundColor Cyan
Write-Host "      cd $ProjectRoot\frontend"
Write-Host "      .\start-all.ps1"
Write-Host ""
Write-Host "    前端列表："
Write-Host "      - admin   管理后台   :5173"
Write-Host "      - kiosk   自助终端   :5174"
Write-Host ""

Write-Host "================================================" -ForegroundColor Green
Write-Host "  默认账号密码（密码均为123456或admin123）" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Green
Write-Host ""
Write-Host "  ┌────────────┬────────────┬─────────────────┐"
Write-Host "  │ 账号       │ 密码       │ 角色             │"
Write-Host "  ├────────────┼────────────┼─────────────────┤"
Write-Host "  │ admin      │ admin123   │ 系统管理员       │"
Write-Host "  ├────────────┼────────────┼─────────────────┤"
Write-Host "  │ director   │ 123456     │ 综治中心主任     │"
Write-Host "  ├────────────┼────────────┼─────────────────┤"
Write-Host "  │ leader     │ 123456     │ 调解组组长       │"
Write-Host "  ├────────────┼────────────┼─────────────────┤"
Write-Host "  │ mediator1  │ 123456     │ 调解员(王)      │"
Write-Host "  ├────────────┼────────────┼─────────────────┤"
Write-Host "  │ mediator2  │ 123456     │ 调解员(赵)      │"
Write-Host "  ├────────────┼────────────┼─────────────────┤"
Write-Host "  │ mediator3  │ 123456     │ 调解员(刘)      │"
Write-Host "  └────────────┴────────────┴─────────────────┘"
Write-Host ""

Write-Host "================================================" -ForegroundColor Green
Write-Host "  初始化脚本执行完成！" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Green
Write-Host ""
Write-Host "  访问地址："
Write-Host "    管理后台: http://localhost:5173"
Write-Host "    自助终端: http://localhost:5174"
Write-Host "    API文档 : http://localhost:8080/swagger/index.html"
Write-Host ""
Write-Host "  如有问题，请检查："
Write-Host "    1. MySQL/TiDB 是否已启动"
Write-Host "    2. Redis 是否已启动"
Write-Host "    3. RocketMQ 是否已启动"
Write-Host "    4. 配置文件 config.yaml 中的连接信息是否正确"
Write-Host ""
