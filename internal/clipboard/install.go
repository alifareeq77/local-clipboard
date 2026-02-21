package clipboard

import (
	"log"
	"os"
	"os/exec"
	"runtime"
)

// TryInstall attempts to install a clipboard tool (wl-clipboard or xclip) on Linux.
// Prefers Wayland (wl-clipboard) if WAYLAND_DISPLAY or XDG_SESSION_TYPE=wayland, else xclip.
// Returns true if an install was attempted (caller should re-run Detect()).
func TryInstall() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	preferWayland := os.Getenv("WAYLAND_DISPLAY") != "" || os.Getenv("XDG_SESSION_TYPE") == "wayland"
	var pkg string
	if preferWayland {
		pkg = "wl-clipboard"
	} else {
		pkg = "xclip"
	}

	// Prefer apt (Debian/Ubuntu), then dnf/yum, then pacman. Run via sudo.
	var args []string
	if path, err := exec.LookPath("apt-get"); err == nil {
		args = []string{path, "install", "-y", pkg}
	} else if path, err := exec.LookPath("apt"); err == nil {
		args = []string{path, "install", "-y", pkg}
	} else if path, err := exec.LookPath("dnf"); err == nil {
		args = []string{path, "install", "-y", pkg}
	} else if path, err := exec.LookPath("yum"); err == nil {
		args = []string{path, "install", "-y", pkg}
	} else if path, err := exec.LookPath("pacman"); err == nil {
		args = []string{path, "-S", "--noconfirm", pkg}
	} else {
		log.Printf("no supported package manager (apt/dnf/pacman) found; install %s manually", pkg)
		return false
	}

	sudoPath, err := exec.LookPath("sudo")
	if err != nil {
		log.Printf("sudo not found; run as root or install %s manually: apt install %s", pkg, pkg)
		return false
	}
	installCmd := exec.Command(sudoPath, args...)
	log.Printf("installing clipboard tool: %s (may prompt for password)", pkg)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	installCmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	if err := installCmd.Run(); err != nil {
		log.Printf("install failed: %v; try manually: sudo apt install %s", err, pkg)
		return false
	}
	log.Printf("installed %s", pkg)
	return true
}
