// codium-installer: cross-platform VSCodium installer and uninstaller.
//
// Usage:
//
//	codium-installer              Install VSCodium (interactive extension picker)
//	codium-installer -a           Install VSCodium with ALL extensions (skip picker)
//	codium-installer -r           Uninstall VSCodium (with config backup)
//	codium-installer --remove     Same as -r
//	codium-installer --uninstall  Same as -r
package main

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// settings.json is embedded into the binary at compile time.
// This means the binary is fully self-contained — even when downloaded
// standalone via curl/Invoke-WebRequest from a GitHub Release.
//
//go:embed config/settings.json
var embeddedConfig embed.FS


// ---------------------------------------------------------------------------
// ANSI colours
// ---------------------------------------------------------------------------

const (
	colorRed   = "\033[1;31m"
	colorGreen = "\033[1;32m"
	colorCyan  = "\033[1;36m"
	colorReset = "\033[0m"
)

// ---------------------------------------------------------------------------
// Console helpers — replicate the visual style of the original scripts.
// ---------------------------------------------------------------------------

func progress(msg string) {
	fmt.Printf("%s➜ %s...%s\n", colorCyan, msg, colorReset)
}

func success(msg string) {
	fmt.Printf("%s✔ %s%s\n", colorGreen, msg, colorReset)
}

func errorMsg(msg string) {
	fmt.Printf("%s✖ %s%s\n", colorRed, msg, colorReset)
}

func banner(title, color string) {
	fmt.Printf("%s========================================\n", color)
	fmt.Printf("    %s\n", title)
	fmt.Printf("========================================%s\n", colorReset)
}

// ---------------------------------------------------------------------------
// Shared utilities
// ---------------------------------------------------------------------------



// runCmd executes an external command, piping stdout/stderr/stdin to the
// current terminal so the user sees all output (including sudo prompts).
func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// runCmdInDir is like runCmd but runs the command inside a specific directory.
func runCmdInDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// copyFile copies a single file from src to dst, creating any missing
// parent directories along the way. Existing dst is overwritten.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create parent dirs: %w", err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy data: %w", err)
	}
	return out.Close()
}

// ---------------------------------------------------------------------------
// Settings management (shared between platforms)
// ---------------------------------------------------------------------------

// installSettings installs settings.json into the platform's VSCodium user config directory.
// It checks for a local 'config/settings.json' next to the binary to allow overrides,
// otherwise it falls back to the embedded settings.json.
// It backs up any existing file first.
func installSettings(settingsDst, backupDir string) {
	var settingsData []byte
	var err error
	var sourceDesc string

	// 1. Try to read a local override next to the executable
	exe, err := os.Executable()
	localSettings := "config/settings.json"
	if err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			localSettings = filepath.Join(filepath.Dir(resolved), "config", "settings.json")
		}
	}

	if _, statErr := os.Stat(localSettings); statErr == nil {
		settingsData, err = os.ReadFile(localSettings)
		sourceDesc = "local override"
	} else {
		// 2. Fall back to the embedded config
		settingsData, err = embeddedConfig.ReadFile("config/settings.json")
		sourceDesc = "embedded config"
	}

	if err != nil {
		errorMsg(fmt.Sprintf("Could not read settings.json (%s): %v", sourceDesc, err))
		return
	}

	progress(fmt.Sprintf("Installing settings.json (%s)", sourceDesc))

	// Backup existing settings before overwriting.
	if _, err := os.Stat(settingsDst); err == nil {
		backupPath := filepath.Join(backupDir, "settings.json")
		if err := copyFile(settingsDst, backupPath); err != nil {
			errorMsg(fmt.Sprintf("Failed to backup settings.json: %v", err))
		} else {
			success(fmt.Sprintf("Existing settings.json backed up to %s", backupDir))
		}
	}

	// Write the new settings
	if err := os.MkdirAll(filepath.Dir(settingsDst), 0o755); err != nil {
		errorMsg(fmt.Sprintf("Failed to create settings directory: %v", err))
		return
	}

	if err := os.WriteFile(settingsDst, settingsData, 0o644); err != nil {
		errorMsg(fmt.Sprintf("Failed to install settings.json: %v", err))
	} else {
		success("settings.json installed")
	}
}

// backupSettingsBeforeRemoval copies settings.json to the backup directory
// so the user never loses their config during an uninstall.
func backupSettingsBeforeRemoval(settingsPath, backupDir string) {
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return
	}

	progress("Backing up settings.json before removing config")
	backupPath := filepath.Join(backupDir, "settings.json")
	if err := copyFile(settingsPath, backupPath); err != nil {
		errorMsg(fmt.Sprintf("Failed to backup settings.json: %v", err))
	} else {
		success(fmt.Sprintf("settings.json backed up to %s", backupDir))
	}
}

// ---------------------------------------------------------------------------
// Extension installation (shared between platforms)
// ---------------------------------------------------------------------------

// installExtensions calls `codium --install-extension` for every extension
// in the list. A single failure does NOT abort the remaining extensions.
func installExtensions(codiumBin string, extensions []string) {
	if len(extensions) == 0 {
		success("No extensions selected — skipping")
		return
	}

	progress("Installing extensions")
	for _, ext := range extensions {
		if err := runCmd(codiumBin, "--install-extension", ext); err != nil {
			errorMsg(fmt.Sprintf("Failed to install extension: %s", ext))
		} else {
			success(fmt.Sprintf("Extension installed: %s", ext))
		}
	}
}

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

func main() {
	remove := false
	all := false

	for _, arg := range os.Args[1:] {
		switch arg {
		case "-r", "--remove", "--uninstall":
			remove = true
		case "-a", "--all":
			all = true
		case "-h", "--help":
			fmt.Println("Usage: codium-installer          Install VSCodium (interactive extension picker)")
			fmt.Println("       codium-installer -a       Install with ALL extensions (skip picker)")
			fmt.Println("       codium-installer -r       Uninstall VSCodium")
			os.Exit(0)
		default:
			errorMsg(fmt.Sprintf("Unknown option: %s", arg))
			fmt.Println("Usage: codium-installer          Install VSCodium")
			fmt.Println("       codium-installer -a       Install with ALL extensions")
			fmt.Println("       codium-installer -r       Uninstall VSCodium")
			os.Exit(1)
		}
	}

	if remove {
		platformUninstall()
	} else {
		platformInstall(all)
	}
}
