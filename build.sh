#!/usr/bin/env bash
# Builds CharSayWhat into bin/.
#   ./build.sh            native Linux build      -> bin/charsaywhat
#   ./build.sh windows    cross-compile Windows   -> bin/charsaywhat.exe
#   ./build.sh all        both
#   ./build.sh install    Linux build, then install binary, icons and a .desktop
#                         launcher under ~/.local (PREFIX=... to change)
#   ./build.sh icons      regenerate icons/*.png from charsaywhat.png (ImageMagick)
#
# The Linux build uses Gio, which needs cgo and these dev packages (Debian/Ubuntu):
#   sudo apt install gcc pkg-config libwayland-dev libx11-dev libx11-xcb-dev \
#     libxkbcommon-x11-dev libgles2-mesa-dev libegl1-mesa-dev libffi-dev \
#     libxcursor-dev libxfixes-dev libvulkan-dev
# The build checks for these first. Gio's opt-out tags drop some of them, but
# the result only runs where that backend isn't needed:
#   TAGS=novulkan ./build.sh        no Vulkan (Gio falls back to OpenGL; fine)
#   TAGS=nox11 ./build.sh           Wayland only: does NOT run in an X11 session
# The Windows build is pure Go and needs no extra packages.
set -euo pipefail
cd "$(dirname "$0")"
mkdir -p bin

# check_linux_deps fails early, naming the missing packages, rather than
# letting cgo fail halfway or producing a binary that can't open a window.
check_linux_deps() {
	local tags=" ${TAGS:-} " missing=()
	local -A pkgs=(
		[egl]=libegl1-mesa-dev [xkbcommon]=libxkbcommon-dev
		[wayland-egl]=libwayland-dev [wayland-client]=libwayland-dev [wayland-cursor]=libwayland-dev
	)
	if [[ $tags != *" nox11 "* ]]; then
		pkgs+=([x11]=libx11-dev [xkbcommon-x11]=libxkbcommon-x11-dev [x11-xcb]=libx11-xcb-dev
			[xcursor]=libxcursor-dev [xfixes]=libxfixes-dev)
	fi
	[[ $tags == *" novulkan "* ]] || pkgs+=([vulkan]=libvulkan-dev)
	for pc in "${!pkgs[@]}"; do
		pkg-config --exists "$pc" 2>/dev/null || missing+=("${pkgs[$pc]}")
	done
	if ((${#missing[@]})); then
		echo "missing Gio build dependencies; install with:" >&2
		echo "  sudo apt install $(printf '%s\n' "${missing[@]}" | sort -u | tr '\n' ' ')" >&2
		exit 1
	fi
}

# Window class / Wayland app_id. Gio defaults it to the executable's name;
# pinning it keeps a renamed binary matched to charsaywhat.desktop.
APP_ID=charsaywhat
ICON_SIZES=(16 24 32 48 64 128 256)

build_linux() {
	check_linux_deps
	go build -trimpath -tags "${TAGS:-}" \
		-ldflags="-s -w -X gioui.org/app.ID=$APP_ID" -o bin/charsaywhat .
	echo "Built bin/charsaywhat"
	set_file_icon bin/charsaywhat
}

# set_file_icon gives one file a custom icon in GVfs file managers (Caja,
# Nautilus, Nemo). ELF binaries can't carry an icon, so this is per-file
# metadata on this machine only; harmless where unsupported.
set_file_icon() {
	command -v gio >/dev/null || return 0
	gio set "$1" metadata::custom-icon "file://$PWD/icons/charsaywhat-256.png" 2>/dev/null || true
}

build_icons() {
	command -v convert >/dev/null || { echo "convert (ImageMagick) not found" >&2; exit 1; }
	mkdir -p icons
	for s in "${ICON_SIZES[@]}"; do
		convert charsaywhat.png -filter Lanczos -resize "${s}x${s}" -strip "PNG32:icons/charsaywhat-$s.png"
	done
	echo "Wrote icons/charsaywhat-{$(IFS=,; echo "${ICON_SIZES[*]}")}.png"
}

install_linux() {
	local prefix="${PREFIX:-$HOME/.local}"
	install -Dm755 bin/charsaywhat "$prefix/bin/charsaywhat"
	for s in "${ICON_SIZES[@]}"; do
		install -Dm644 "icons/charsaywhat-$s.png" "$prefix/share/icons/hicolor/${s}x${s}/apps/charsaywhat.png"
	done
	install -Dm644 charsaywhat.desktop "$prefix/share/applications/charsaywhat.desktop"
	command -v gtk-update-icon-cache >/dev/null &&
		gtk-update-icon-cache -qtf "$prefix/share/icons/hicolor" 2>/dev/null || true
	command -v update-desktop-database >/dev/null &&
		update-desktop-database -q "$prefix/share/applications" 2>/dev/null || true
	set_file_icon "$prefix/bin/charsaywhat"
	echo "Installed to $prefix (binary, icons, applications/charsaywhat.desktop)"
	case ":$PATH:" in *":$prefix/bin:"*) ;; *) echo "note: $prefix/bin is not on PATH" ;; esac
}

build_windows() {
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
		go build -trimpath -ldflags="-s -w" -o bin/charsaywhat.exe .
	echo "Built bin/charsaywhat.exe"
}

case "${1:-linux}" in
linux) build_linux ;;
windows) build_windows ;;
all)
	build_linux
	build_windows
	;;
install)
	build_linux
	install_linux
	;;
icons) build_icons ;;
*)
	echo "usage: $0 [linux|windows|all|install|icons]" >&2
	exit 1
	;;
esac
