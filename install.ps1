param(
    [switch]$r
)

$ErrorActionPreference = "Stop"

$ScriptDir = $PSScriptRoot
$ConfigRoot = Join-Path $env:APPDATA "VSCodium"
$ConfigDest = Join-Path $ConfigRoot "User"
$BackupDir = Join-Path $env:USERPROFILE "BKP.config"

$Extensions = @(
    "esbenp.prettier-vscode",
    "Catppuccin.catppuccin-vsc-pack"
)

function Write-Step($msg) { Write-Host "-> $msg..." -ForegroundColor Cyan }
function Write-Ok($msg) { Write-Host "[OK] $msg" -ForegroundColor Green }
function Write-Err($msg) { Write-Host "[X] $msg" -ForegroundColor Red }

function Install-VSCodium {
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "    VSCODIUM INSTALLER" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan

    if (-not (Get-Command winget -ErrorAction SilentlyContinue)) {
        Write-Err "winget not found. Install 'App Installer' from the Microsoft Store first."
        exit 1
    }

    Write-Step "Installing VSCodium (winget)"
    winget install -e --id VSCodium.VSCodium --accept-source-agreements --accept-package-agreements --silent
    if ($LASTEXITCODE -ne 0) {
        Write-Err "Failed to install VSCodium"
        exit 1
    }
    Write-Ok "VSCodium installed"

    $codiumCmd = Get-Command codium -ErrorAction SilentlyContinue
    if (-not $codiumCmd) {
        $possiblePaths = @(
            (Join-Path $env:LOCALAPPDATA "Programs\VSCodium\bin\codium.cmd"),
            (Join-Path $env:ProgramFiles "VSCodium\bin\codium.cmd")
        )
        foreach ($p in $possiblePaths) {
            if (Test-Path $p) {
                $env:Path = "$(Split-Path $p);$env:Path"
                break
            }
        }
    }

    Write-Step "Installing extensions"
    foreach ($ext in $Extensions) {
        codium --install-extension $ext
        if ($LASTEXITCODE -eq 0) {
            Write-Ok "Extension installed: $ext"
        } else {
            Write-Err "Failed to install extension: $ext"
        }
    }

    $settingsSource = Join-Path $ScriptDir "config\settings.json"
    if (Test-Path $settingsSource) {
        Write-Step "Installing settings.json"
        New-Item -ItemType Directory -Force -Path $ConfigDest | Out-Null
        $settingsDest = Join-Path $ConfigDest "settings.json"
        if (Test-Path $settingsDest) {
            New-Item -ItemType Directory -Force -Path $BackupDir | Out-Null
            Copy-Item $settingsDest (Join-Path $BackupDir "settings.json") -Force
        }
        Copy-Item $settingsSource $settingsDest -Force
        Write-Ok "settings.json installed"
    }

    Write-Host "========================================" -ForegroundColor Green
    Write-Host "   VSCODIUM SETUP COMPLETED" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "If 'codium' is not recognized, restart your terminal." -ForegroundColor Yellow
}

function Uninstall-VSCodium {
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "    VSCODIUM UNINSTALLER" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan

    Write-Step "Removing VSCodium (winget)"
    winget uninstall -e --id VSCodium.VSCodium
    if ($LASTEXITCODE -eq 0) {
        Write-Ok "VSCodium removed"
    } else {
        Write-Err "VSCodium was not installed or failed to remove"
    }

    $settingsDest = Join-Path $ConfigDest "settings.json"
    if (Test-Path $settingsDest) {
        Write-Step "Backing up settings.json before removing config"
        New-Item -ItemType Directory -Force -Path $BackupDir | Out-Null
        Copy-Item $settingsDest (Join-Path $BackupDir "settings.json") -Force
        Write-Ok "settings.json backed up to $BackupDir"
    }

    if (Test-Path $ConfigRoot) {
        Write-Step "Removing $ConfigRoot"
        Remove-Item -Recurse -Force $ConfigRoot
        Write-Ok "Config directory removed"
    } else {
        Write-Ok "No config directory found, nothing to remove"
    }

    Write-Host "========================================" -ForegroundColor Green
    Write-Host "   VSCODIUM REMOVED" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
}

if ($r) {
    Uninstall-VSCodium
} else {
    Install-VSCodium
}