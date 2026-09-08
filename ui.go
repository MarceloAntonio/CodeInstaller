package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// ---------------------------------------------------------------------------
// Extension catalogue — edit this list to add/remove extensions.
//
// Extensions with Default=true are pre-selected in the interactive picker.
// Language-specific extensions default to false so the user opts in.
// ---------------------------------------------------------------------------

// ExtensionItem holds metadata for the interactive picker.
type ExtensionItem struct {
	ID       string
	Category string
	Default  bool
}

var allExtensions = []ExtensionItem{
	// Core / editor — pre-selected
	{ID: "esbenp.prettier-vscode", Category: "Core", Default: true},
	{ID: "Catppuccin.catppuccin-vsc-pack", Category: "Core", Default: true},
	{ID: "eamodio.gitlens", Category: "Core", Default: true},
	{ID: "usernamehw.errorlens", Category: "Core", Default: true},
	{ID: "foxundermoon.shell-format", Category: "Core", Default: true},
	// Go
	{ID: "golang.go", Category: "Go", Default: false},
	// Rust
	{ID: "rust-lang.rust-analyzer", Category: "Rust", Default: false},
	// Java
	{ID: "redhat.java", Category: "Java", Default: false},
	{ID: "vscjava.vscode-maven", Category: "Java", Default: false},
	// Python
	{ID: "ms-python.python", Category: "Python", Default: false},
	{ID: "charliermarsh.ruff", Category: "Python", Default: false},
	// C / C++
	{ID: "llvm-vs-code-extensions.vscode-clangd", Category: "C/C++", Default: false},
	{ID: "vadimcn.vscode-lldb", Category: "C/C++", Default: false},
	// JavaScript / TypeScript
	{ID: "dbaeumer.vscode-eslint", Category: "JS/TS", Default: false},
}

// ---------------------------------------------------------------------------
// Selection state — simple wrapper so we can mutate Selected per item.
// ---------------------------------------------------------------------------

type selItem struct {
	ExtensionItem
	Selected bool
}

// ---------------------------------------------------------------------------
// Interactive checkbox picker
// ---------------------------------------------------------------------------

// selectExtensions shows an interactive checkbox list in the terminal.
// Returns the IDs of the extensions the user selected.
//
// If selectAll is true, it skips the UI and returns every extension.
// If stdin is not a terminal (piped), it returns only the default-selected ones.
func selectExtensions(selectAll bool) []string {
	// --all flag: skip UI, return everything.
	if selectAll {
		ids := make([]string, len(allExtensions))
		for i, ext := range allExtensions {
			ids[i] = ext.ID
		}
		return ids
	}

	// Build mutable selection state.
	items := make([]selItem, len(allExtensions))
	for i, ext := range allExtensions {
		items[i] = selItem{ExtensionItem: ext, Selected: ext.Default}
	}

	// If stdin is not a terminal, return defaults without interactive UI.
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return selectedIDs(items)
	}

	// Put terminal in raw mode so we can read single keypresses.
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return selectedIDs(items)
	}
	defer term.Restore(fd, oldState)

	cursor := 0

	// Initial render.
	renderPicker(items, cursor, false)

	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}

		if n == 1 {
			switch buf[0] {
			case ' ': // toggle
				items[cursor].Selected = !items[cursor].Selected
			case 13: // Enter — confirm
				fmt.Print("\r\n")
				term.Restore(fd, oldState)
				return selectedIDs(items)
			case 'a', 'A': // select all
				for i := range items {
					items[i].Selected = true
				}
			case 'n', 'N': // deselect all
				for i := range items {
					items[i].Selected = false
				}
			case 'q', 27: // q or Esc — cancel, use defaults
				fmt.Print("\r\n")
				term.Restore(fd, oldState)
				for i := range items {
					items[i].Selected = items[i].Default
				}
				return selectedIDs(items)
			case 3: // Ctrl-C
				fmt.Print("\r\n")
				term.Restore(fd, oldState)
				os.Exit(130)
			}
		} else if n == 3 && buf[0] == 27 && buf[1] == '[' {
			switch buf[2] {
			case 'A': // ↑
				if cursor > 0 {
					cursor--
				}
			case 'B': // ↓
				if cursor < len(items)-1 {
					cursor++
				}
			}
		}

		// Redraw.
		renderPicker(items, cursor, true)
	}

	return selectedIDs(items)
}

// ---------------------------------------------------------------------------
// Rendering
// ---------------------------------------------------------------------------

// pickerLines returns how many terminal lines the picker occupies.
func pickerLines(items []selItem) int {
	lines := 3 // header + help + blank
	lastCat := ""
	for _, it := range items {
		if it.Category != lastCat {
			lines++ // category header
			lastCat = it.Category
		}
		lines++ // item row
	}
	return lines
}

// renderPicker draws the entire picker. On redraw it moves the cursor up first.
func renderPicker(items []selItem, cursor int, redraw bool) {
	if redraw {
		total := pickerLines(items)
		fmt.Printf("\033[%dA", total)
	}

	// Header.
	fmt.Printf("\r\033[2K%sSelect extensions to install:%s\r\n", colorCyan, colorReset)
	fmt.Printf("\r\033[2K  ↑↓ move   Space toggle   a all   n none   Enter confirm   q cancel\r\n")
	fmt.Print("\r\033[2K\r\n")

	lastCat := ""
	for i, it := range items {
		// Category header.
		if it.Category != lastCat {
			fmt.Printf("\r\033[2K %s%s:%s\r\n", colorCyan, it.Category, colorReset)
			lastCat = it.Category
		}

		// Checkbox character.
		check := " "
		if it.Selected {
			check = "x"
		}

		// Current row highlight.
		if i == cursor {
			fmt.Printf("\r\033[2K   %s▸ [%s] %s%s\r\n", colorGreen, check, it.ID, colorReset)
		} else {
			fmt.Printf("\r\033[2K     [%s] %s\r\n", check, it.ID)
		}
	}
}

// selectedIDs returns the IDs of all selected items.
func selectedIDs(items []selItem) []string {
	var ids []string
	for _, it := range items {
		if it.Selected {
			ids = append(ids, it.ID)
		}
	}
	return ids
}
