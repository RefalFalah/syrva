package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/RefalFalah/syrva/internal/host"
)

func ExpandPath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home directory tidak ditemukan: %w", err)
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, filepath.FromSlash(path[2:])), nil
	}
	if strings.HasPrefix(path, "~") {
		return "", fmt.Errorf("path ~user tidak didukung; gunakan ~/ atau path lengkap")
	}
	return filepath.FromSlash(path), nil
}

// BaseArgs never alters agent, known_hosts, or host-key checking settings.
func BaseArgs(h host.Host) ([]string, error) {
	if err := host.Validate(h); err != nil {
		return nil, err
	}
	args := []string{"-p", strconv.Itoa(h.Port), "-l", h.User}
	if h.IdentityFile != "" {
		path, err := ExpandPath(h.IdentityFile)
		if err != nil {
			return nil, err
		}
		args = append(args, "-i", path)
	}
	return args, nil
}

func Destination(h host.Host) string {
	return strings.Trim(h.Hostname, "[]")
}

func ConnectArgs(h host.Host) ([]string, error) {
	args, err := BaseArgs(h)
	if err != nil {
		return nil, err
	}
	return append(args, "--", Destination(h)), nil
}
