# Codium Installer

Cross-platform installer for [VSCodium](https://vscodium.com/) with my personal setup — curated extensions and a shared `settings.json` — for both Arch Linux and Windows.

Written as a **single Go program** that detects the OS at compile time via build tags and runs the appropriate logic.

---

## Quick install (pre-built binaries)

Download the latest release for your platform from the [Releases page](https://github.com/MarceloAntonio/CodeInstaller/releases/latest).

### Linux

```bash
# Download
curl -L -o codium-installer https://github.com/MarceloAntonio/CodeInstaller/releases/latest/download/codium-installer-linux-amd64
```
```bash
# Make executable
chmod +x codium-installer
```
```bash
# Run (installs VSCodium + extensions + settings)
./codium-installer
```

### Windows (PowerShell)

```powershell
# Download
Invoke-WebRequest -Uri "https://github.com/MarceloAntonio/CodeInstaller/releases/latest/download/codium-installer-windows-amd64.exe" -OutFile "codium-installer.exe"
```
```powershell
# Run (installs VSCodium + extensions + settings)
.\codium-installer.exe
```

> **Note:** If you want the bundled `settings.json`, clone the repo and place the binary next to the `config/` folder, or build from source (see below).

---

## What it does

- Installs VSCodium
  - **Arch Linux:** builds `vscodium-bin` from the AUR (no AUR helper required)
  - **Windows:** installs via `winget`
- Installs a fixed set of extensions
- Applies `config/settings.json` to VSCodium's user settings
- Backs up any existing config before overwriting it
- Can fully remove everything with a single flag

---

## Requirements

| Platform | Requirements |
| --- | --- |
| Arch Linux | `git` (installed automatically if missing) |
| Windows | [`winget`](https://github.com/microsoft/winget-cli) (built into Windows 10 2004+ / 11) |

---

## Building from source

```bash
git clone https://github.com/MarceloAntonio/CodeInstaller && cd CodeInstaller
```

**For Linux (native build):**

```bash
go build -o codium-installer .

```

```bash
./codium-installer
```
**For Windows (cross-compile from Linux, or native on Windows):**

```bash
GOOS=windows GOARCH=amd64 go build -o codium-installer.exe .
```

---

## Uninstalling

Pass the `-r` flag to remove VSCodium and its config (with a backup taken first):

```bash
./codium-installer -r
```

```powershell
.\codium-installer.exe -r
```

`--remove` and `--uninstall` are also accepted.

---

## Configuration

Drop your own `settings.json` inside `config/` next to the binary and the installer will pick it up automatically:

```
CodeInstaller/
├── codium-installer       (or .exe)
├── config/
│   └── settings.json
└── ...
```

If `config/settings.json` doesn't exist next to the binary, that step is simply skipped.

---

## Extensions installed

| Category | Extension |
| --- | --- |
| Core | `esbenp.prettier-vscode`, `Catppuccin.catppuccin-vsc-pack`, `eamodio.gitlens`, `usernamehw.errorlens`, `foxundermoon.shell-format` |
| Go | `golang.go` |
| Rust | `rust-lang.rust-analyzer` |
| Java | `redhat.java`, `vscjava.vscode-maven` |
| Python | `ms-python.python`, `charliermarsh.ruff` |
| C/C++ | `llvm-vs-code-extensions.vscode-clangd`, `vadimcn.vscode-lldb` |
| JS/TS | `dbaeumer.vscode-eslint` |

---

## Backups

Before overwriting or removing anything, the installer backs up your existing config:

| Platform | Backup location |
| --- | --- |
| Arch Linux | `~/BKP.config/settings.json` |
| Windows | `%USERPROFILE%\BKP.config\settings.json` |

---

## Releases

Releases are built automatically by GitHub Actions whenever a version tag is pushed:

```bash
git tag v1.0.0
git push origin v1.0.0
```

This creates a GitHub Release with pre-built binaries for Linux (amd64) and Windows (amd64).