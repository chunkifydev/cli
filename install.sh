#!/bin/sh
# Installation script for Chunkify CLI

set -e

main() {
    os=$(uname -s)
    arch=$(uname -m)
    version=${1:-latest}

    # Define installation directory
    chunkify_install="${CHUNKIFY_INSTALL:-$HOME/.chunkify}"
    bin_dir="$chunkify_install/bin"
    tmp_dir="$chunkify_install/tmp"
    exe="$bin_dir/chunkify"

    # Create necessary directories
    mkdir -p "$bin_dir"
    mkdir -p "$tmp_dir"

    # Construct download URL based on OS and architecture
    download_url="https://api.chunkify.dev/releases/$os/$arch/$version"

    echo "Downloading Chunkify CLI..."
    curl -q --fail --location --progress-bar --output "$tmp_dir/chunkify.tar.gz" "$download_url"

    echo "Installing Chunkify CLI..."
    # Extract to tmp dir
    tar -C "$tmp_dir" -xzf "$tmp_dir/chunkify.tar.gz"
    chmod +x "$tmp_dir/chunkify"
    
    # Move into place
    mv "$tmp_dir/chunkify" "$exe"
    rm "$tmp_dir/chunkify.tar.gz"

    echo "Chunkify CLI was installed successfully to $exe"
    if command -v chunkify >/dev/null; then
        echo "Run 'chunkify --help' to get started"
    else
        case $SHELL in
        /bin/zsh) shell_profile=".zshrc" ;;
        *) shell_profile=".bash_profile" ;;
        esac
        echo "Manually add the directory to your \$HOME/$shell_profile (or similar)"
        echo "  export CHUNKIFY_INSTALL=\"$chunkify_install\""
        echo "  export PATH=\"\$CHUNKIFY_INSTALL/bin:\$PATH\""
        echo "Run '$exe --help' to get started"
    fi
}

main "$1" 