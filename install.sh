#!/bin/bash

if [[ $EUID -eq 0 ]]; then
  echo -e "\033[1;31m✖ Do not run this script as root (sudo). The script will prompt for your password when needed.\033[0m"
  exit 1
fi

set -e

RED="\033[1;31m"
GREEN="\033[1;32m"
CYAN="\033[1;36m"
RESET="\033[0m"

DOTFILES_DIR="$(pwd)"
CONFIG_DEST="$HOME/.config/VSCodium/User"
BACKUP_DIR="$HOME/BKP.config"

progress() { echo -e "${CYAN}➜ $1...${RESET}"; }
success() { echo -e "${GREEN}✔ $1${RESET}"; }
error() { echo -e "${RED}✖ $1${RESET}"; }

EXTENSIONS=(
  esbenp.prettier-vscode
  Catppuccin.catppuccin-vsc-pack
)

install() {
  echo -e "${CYAN}"
  echo "========================================"
  echo "    VSCODIUM INSTALLER"
  echo "========================================"
  echo -e "${RESET}"

  progress "Installing build prerequisites (git, base-devel)"
  sudo pacman -S --needed --noconfirm git base-devel
  success "Prerequisites ready"

  progress "Installing vscodium-bin (AUR)"
  rm -rf /tmp/vscodium-bin
  if git clone https://aur.archlinux.org/vscodium-bin.git /tmp/vscodium-bin \
      && cd /tmp/vscodium-bin \
      && makepkg -si --noconfirm; then
    success "vscodium-bin installed"
  else
    error "Failed to install vscodium-bin"
    exit 1
  fi
  cd - >/dev/null
  rm -rf /tmp/vscodium-bin

  progress "Installing extensions"
  for ext in "${EXTENSIONS[@]}"; do
    if codium --install-extension "$ext"; then
      success "Extension installed: $ext"
    else
      error "Failed to install extension: $ext"
    fi
  done

  if [ -f "$DOTFILES_DIR/config/settings.json" ]; then
    progress "Installing settings.json"
    mkdir -p "$CONFIG_DEST"
    if [ -f "$CONFIG_DEST/settings.json" ]; then
      mkdir -p "$BACKUP_DIR"
      cp "$CONFIG_DEST/settings.json" "$BACKUP_DIR/settings.json"
    fi
    cp "$DOTFILES_DIR/config/settings.json" "$CONFIG_DEST/settings.json"
    success "settings.json installed"
  fi

  echo -e "${GREEN}"
  echo "========================================"
  echo "   ✔ VSCODIUM SETUP COMPLETED"
  echo "========================================"
  echo -e "${RESET}"
}

uninstall() {
  echo -e "${CYAN}"
  echo "========================================"
  echo "    VSCODIUM UNINSTALLER"
  echo "========================================"
  echo -e "${RESET}"

  progress "Removing vscodium-bin"
  if sudo pacman -Rns --noconfirm vscodium-bin; then
    success "vscodium-bin removed"
  else
    error "vscodium-bin was not installed or failed to remove"
  fi

  if [ -f "$CONFIG_DEST/settings.json" ]; then
    progress "Backing up settings.json before removing config"
    mkdir -p "$BACKUP_DIR"
    cp "$CONFIG_DEST/settings.json" "$BACKUP_DIR/settings.json"
    success "settings.json backed up to $BACKUP_DIR"
  fi

  if [ -d "$HOME/.config/VSCodium" ]; then
    progress "Removing ~/.config/VSCodium"
    rm -rf "$HOME/.config/VSCodium"
    success "Config directory removed"
  else
    success "No config directory found, nothing to remove"
  fi

  echo -e "${GREEN}"
  echo "========================================"
  echo "   ✔ VSCODIUM REMOVED"
  echo "========================================"
  echo -e "${RESET}"
}

case "$1" in
  -r|--remove|--uninstall)
    uninstall
    ;;
  "")
    install
    ;;
  *)
    error "Unknown option: $1"
    echo "Usage: ./install.sh          Install VSCodium"
    echo "       ./install.sh -r       Uninstall VSCodium"
    exit 1
    ;;
esac