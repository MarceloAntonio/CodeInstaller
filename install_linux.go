//go:build linux

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// ---------------------------------------------------------------------------
// Linux (Arch) — install
// ---------------------------------------------------------------------------

func platformInstall(selectAll bool) {
	// Refuse to run as root — the script will ask for sudo when needed.
	if os.Geteuid() == 0 {
		errorMsg("Do not run this program as root (sudo). It will prompt for your password when needed.")
		os.Exit(1)
	}

	banner("VSCODIUM INSTALLER", colorCyan)

	// 1. Install prerequisites via pacman.
	progress("Installing build prerequisites (git, base-devel)")
	if err := runCmd("sudo", "pacman", "-S", "--needed", "--noconfirm", "git", "base-devel"); err != nil {
		errorMsg(fmt.Sprintf("Failed to install prerequisites: %v", err))
		os.Exit(1)
	}
	success("Prerequisites ready")

	// 2. Build and install vscodium-bin from the AUR (no AUR helper).
	progress("Installing vscodium-bin (AUR)")
	tmpDir := "/tmp/vscodium-bin"
	os.RemoveAll(tmpDir)

	if err := runCmd("git", "clone", "https://aur.archlinux.org/vscodium-bin.git", tmpDir); err != nil {
		errorMsg(fmt.Sprintf("Failed to clone AUR repo: %v", err))
		os.Exit(1)
	}

	if err := runCmdInDir(tmpDir, "makepkg", "-si", "--noconfirm"); err != nil {
		errorMsg(fmt.Sprintf("Failed to install vscodium-bin: %v", err))
		os.RemoveAll(tmpDir)
		os.Exit(1)
	}
	os.RemoveAll(tmpDir)
	success("vscodium-bin installed")

	// 3. Let the user pick extensions, then install them.
	selected := selectExtensions(selectAll)
	installExtensions("codium", selected)

	// 4. Install settings.json (with backup of existing one).
	home, _ := os.UserHomeDir()
	installSettings(
		filepath.Join(home, ".config", "VSCodium", "User", "settings.json"),
		filepath.Join(home, "BKP.config"),
	)

	banner("✔ VSCODIUM SETUP COMPLETED", colorGreen)
}

// ---------------------------------------------------------------------------
// Linux (Arch) — uninstall
// ---------------------------------------------------------------------------

func platformUninstall() {
	if os.Geteuid() == 0 {
		errorMsg("Do not run this program as root (sudo).")
		os.Exit(1)
	}

	banner("VSCODIUM UNINSTALLER", colorCyan)

	home, _ := os.UserHomeDir()
	configRoot := filepath.Join(home, ".config", "VSCodium")
	settingsPath := filepath.Join(configRoot, "User", "settings.json")
	backupDir := filepath.Join(home, "BKP.config")

	// 1. Remove the package.
	progress("Removing vscodium-bin")
	if err := runCmd("sudo", "pacman", "-Rns", "--noconfirm", "vscodium-bin"); err != nil {
		errorMsg("vscodium-bin was not installed or failed to remove")
	} else {
		success("vscodium-bin removed")
	}

	// 2. Backup settings before wiping the config directory.
	backupSettingsBeforeRemoval(settingsPath, backupDir)

	// 3. Remove the entire config directory.
	if info, err := os.Stat(configRoot); err == nil && info.IsDir() {
		progress("Removing ~/.config/VSCodium")
		if err := os.RemoveAll(configRoot); err != nil {
			errorMsg(fmt.Sprintf("Failed to remove config directory: %v", err))
		} else {
			success("Config directory removed")
		}
	} else {
		success("No config directory found, nothing to remove")
	}

	banner("✔ VSCODIUM REMOVED", colorGreen)
}
