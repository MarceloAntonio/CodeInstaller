# Codium Installer

Cross-platform installer for [VSCodium](https://vscodium.com/) with my personal setup — curated extensions and a shared `settings.json` — for both Arch Linux and Windows.

Written as a **single Go program** that detects the OS at compile time via build tags and runs the appropriate logic.

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
| Arch Linux | `git` (installed automatically if missing), Go compiler (to build) |
| Windows | [`winget`](https://github.com/microsoft/winget-cli) (built into Windows 10 2004+ / 11), Go compiler (to build) |

---

## Building

```bash
git clone https://github.com/MarceloAntonio/CodeInstaller
cd CodeInstaller
```

**For Linux (native build):**

```bash
go build -o codium-installer .
```

**For Windows (cross-compile from Linux, or native on Windows):**

```bash
GOOS=windows GOARCH=amd64 go build -o codium-installer.exe .
```

---

## Installation

Run the built binary — it detects your OS automatically:

```bash
./codium-installer
```

```powershell
.\codium-installer.exe
```

> **Note (Linux):** Do not run as root. The program will prompt for `sudo` when needed.

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

Drop your own `settings.json` inside `config/` and the installer will pick it up automatically:

```
CodeInstaller/
├── main.go
├── install_linux.go
├── install_windows.go
├── go.mod
└── config/
    └── settings.json
```

If `config/settings.json` doesn't exist, that step is simply skipped.

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