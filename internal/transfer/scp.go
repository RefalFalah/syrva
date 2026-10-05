package transfer

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/RefalFalah/syrva/internal/host"
	"github.com/RefalFalah/syrva/internal/ssh"
)

func baseArgs(h host.Host) ([]string, error) {
	if err := host.Validate(h); err != nil {
		return nil, err
	}
	// -s selects native scp's SFTP mode on OpenSSH 8.7/8.8 and is accepted
	// by modern versions. No legacy remote shell is used to interpret paths.
	args := []string{"-s", "-P", strconv.Itoa(h.Port), "-o", "User=" + h.User}
	if h.IdentityFile != "" {
		key, err := ssh.ExpandPath(h.IdentityFile)
		if err != nil {
			return nil, err
		}
		args = append(args, "-i", key)
	}
	return args, nil
}

func LocalPath(value string) (string, error) {
	if strings.TrimSpace(value) == "" || strings.ContainsFunc(value, unicode.IsControl) {
		return "", fmt.Errorf("path lokal kosong atau mengandung karakter kontrol")
	}
	expanded, err := ssh.ExpandPath(value)
	if err != nil {
		return "", err
	}
	absolute, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("path lokal tidak valid: %w", err)
	}
	// Absolute paths disambiguate filenames containing ':' or starting with '-'.
	// Native Windows OpenSSH supports drive-letter paths with forward slashes.
	return filepath.ToSlash(absolute), nil
}

func remotePath(h host.Host, value string) (string, error) {
	if strings.TrimSpace(value) == "" || strings.ContainsFunc(value, unicode.IsControl) {
		return "", fmt.Errorf("path remote kosong atau mengandung karakter kontrol")
	}
	hostname := ssh.Destination(h)
	if strings.Contains(hostname, ":") {
		hostname = "[" + hostname + "]"
	}
	// Paths are literal argv values, not locally shell-quoted strings.
	return hostname + ":" + value, nil
}

func UploadArgs(h host.Host, local, remote string) ([]string, error) {
	args, err := baseArgs(h)
	if err != nil {
		return nil, err
	}
	local, err = LocalPath(local)
	if err != nil {
		return nil, err
	}
	remote, err = remotePath(h, remote)
	if err != nil {
		return nil, err
	}
	return append(args, "--", local, remote), nil
}

func DownloadArgs(h host.Host, remote, local string) ([]string, error) {
	args, err := baseArgs(h)
	if err != nil {
		return nil, err
	}
	remote, err = remotePath(h, remote)
	if err != nil {
		return nil, err
	}
	local, err = LocalPath(local)
	if err != nil {
		return nil, err
	}
	return append(args, "--", remote, local), nil
}
