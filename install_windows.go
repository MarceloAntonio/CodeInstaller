//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ---------------------------------------------------------------------------
// Windows — install
// ---------------------------------------------------------------------------

func platformInstall() {
	banner("VSCODIUM INSTALLER", colorCyan)

	// 1. Verify winget is available.
	if _, err := exec.LookPath("winget"); err != nil {
		errorMsg("winget not found. Install 'App Installer' from the Microsoft Store first.")
		os.Exit(1)
	}

	// 2. Install VSCodium via winget.
	progress("Installing VSCodium (winget)")
	if err := runCmd("winget", "install", "-e",
		"--id", "VSCodium.VSCodium",
		"--accept-source-agreements",
		"--accept-package-agreements",
		"--silent",
	); err != nil {
		errorMsg(fmt.Sprintf("Failed to install VSCodium: %v", err))
		os.Exit(1)
	}
	success("VSCodium installed")

	// 3. Locate the codium binary — it may not be in PATH right after install.
	codiumBin := findCodium()

	// 4. Install extensions.
	installExtensions(codiumBin)

	// 5. Copy settings.json (with backup of existing one).
	dir, err := exeDir()
	if err != nil {
		errorMsg(fmt.Sprintf("Could not determine executable directory: %v", err))
	} else {
		appData := os.Getenv("APPDATA")
		userProfile := os.Getenv("USERPROFILE")
		installSettings(
			filepath.Join(dir, "config", "settings.json"),
			filepath.Join(appData, "VSCodium", "User", "settings.json"),
			filepath.Join(userProfile, "BKP.config"),
		)
	}

	banner("✔ VSCODIUM SETUP COMPLETED", colorGreen)
	fmt.Printf("\033[1;33mIf 'codium' is not recognized, restart your terminal.\033[0m\n")
}

// ---------------------------------------------------------------------------
// Windows — uninstall
// ---------------------------------------------------------------------------

func platformUninstall() {
	banner("VSCODIUM UNINSTALLER", colorCyan)

	appData := os.Getenv("APPDATA")
	userProfile := os.Getenv("USERPROFILE")
	configRoot := filepath.Join(appData, "VSCodium")
	settingsPath := filepath.Join(configRoot, "User", "settings.json")
	backupDir := filepath.Join(userProfile, "BKP.config")

	// 1. Remove via winget.
	progress("Removing VSCodium (winget)")
	if err := runCmd("winget", "uninstall", "-e", "--id", "VSCodium.VSCodium"); err != nil {
		errorMsg("VSCodium was not installed or failed to remove")
	} else {
		success("VSCodium removed")
	}

	// 2. Backup settings before wiping the config directory.
	backupSettingsBeforeRemoval(settingsPath, backupDir)

	// 3. Remove the entire config directory.
	if info, err := os.Stat(configRoot); err == nil && info.IsDir() {
		progress(fmt.Sprintf("Removing %s", configRoot))
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

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// findCodium locates the codium.cmd binary. It first checks PATH, then falls
// back to known default installation directories.
func findCodium() string {
	// Already in PATH?
	if p, err := exec.LookPath("codium"); err == nil {
		return p
	}

	// Check well-known install locations.
	candidates := []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "VSCodium", "bin", "codium.cmd"),
		filepath.Join(os.Getenv("ProgramFiles"), "VSCodium", "bin", "codium.cmd"),
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Last resort — hope it shows up in PATH at call time.
	return "codium"
}
