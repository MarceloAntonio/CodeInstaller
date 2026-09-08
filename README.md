# Codium Installer

Cross-platform installer for [VSCodium](https://vscodium.com/) with my personal setup — curated extensions and a shared `settings.json` — for both Arch Linux and Windows.

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

## Installation

**1. Clone the repository**

```bash
git clone https://github.com/MarceloAntonio/CodeInstaller
cd CodeInstaller
```

**2. Run the installer for your OS**

Arch Linux:
```bash
./install.sh
```

Windows (PowerShell):
```powershell
.\install.ps1
```

> **Note:** Windows blocks script execution by default. If you get a "cannot be loaded because running scripts is disabled" error, run:
> ```powershell
> powershell -ExecutionPolicy Bypass -File .\install.ps1
> ```

---

## Uninstalling

Both scripts accept a `-r` flag to remove VSCodium and its config (with a backup taken first).

```bash
./install.sh -r
```

```powershell
.\install.ps1 -r
```

---

## Configuration

Drop your own `settings.json` inside `config/` and both scripts will pick it up automatically:

```
CodeInstaller/
├── install.sh
├── install.ps1
└── config/
    └── settings.json
```

If `config/settings.json` doesn't exist, that step is simply skipped.

---

## Extensions installed

- [`esbenp.prettier-vscode`](https://open-vsx.org/extension/esbenp/prettier-vscode) — Prettier
- [`Catppuccin.catppuccin-vsc-pack`](https://open-vsx.org/extension/Catppuccin/catppuccin-vsc-pack) — Catppuccin theme pack

---

## Backups

Before overwriting or removing anything, both scripts back up your existing config:

| Platform | Backup location |
| --- | --- |
| Arch Linux | `~/BKP.config` |
| Windows | `%USERPROFILE%\BKP.config` |