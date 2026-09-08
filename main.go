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
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

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

// exeDir returns the directory that contains the running executable,
// resolving symlinks so that config/settings.json is always found
// regardless of how the binary was invoked.
//
// It checks two locations (in order):
//  1. The directory containing the executable (for standalone binary deployment)
//  2. The current working directory (for `go run` or running from the project root)
//
// Returns the first directory where config/settings.json actually exists,
// or falls back to the executable directory if neither has it.
func findConfigDir() string {
	candidates := []string{}

	// 1st priority: directory of the executable itself.
	if exe, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			candidates = append(candidates, filepath.Dir(resolved))
		}
	}

	// 2nd priority: current working directory (matches original bash script behaviour).
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}

	settingsRel := filepath.Join("config", "settings.json")
	for _, dir := range candidates {
		if _, err := os.Stat(filepath.Join(dir, settingsRel)); err == nil {
			return dir
		}
	}

	// Fallback: return first candidate even if settings.json is absent
	// (the install step will simply skip when it doesn't find the file).
	if len(candidates) > 0 {
		return candidates[0]
	}
	return "."
}


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

// installSettings copies config/settings.json next to the executable into the
// platform's VSCodium user config directory, backing up any existing file first.
func installSettings(settingsSrc, settingsDst, backupDir string) {
	if _, err := os.Stat(settingsSrc); os.IsNotExist(err) {
		return // no settings.json bundled — skip silently
	}

	progress("Installing settings.json")

	// Backup existing settings before overwriting.
	if _, err := os.Stat(settingsDst); err == nil {
		backupPath := filepath.Join(backupDir, "settings.json")
		if err := copyFile(settingsDst, backupPath); err != nil {
			errorMsg(fmt.Sprintf("Failed to backup settings.json: %v", err))
		} else {
			success(fmt.Sprintf("Existing settings.json backed up to %s", backupDir))
		}
	}

	if err := copyFile(settingsSrc, settingsDst); err != nil {
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
