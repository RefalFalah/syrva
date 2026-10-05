package transfer

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/RefalFalah/syrva/internal/host"
)

func TestSCPArgs(t *testing.T) {
	h := host.Host{Alias: "dev", Hostname: "::1", User: "root", Port: 2222, IdentityFile: "~/.ssh/key file"}
	local := filepath.Join(t.TempDir(), "backup file.sql")
	home, _ := os.UserHomeDir()
	want := []string{"-s", "-P", "2222", "-o", "User=root", "-i", filepath.Join(home, ".ssh", "key file"), "--", filepath.ToSlash(local), "[::1]:/tmp/path with space/"}
	got, err := UploadArgs(h, local, "/tmp/path with space/")
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("UploadArgs = %#v, %v", got, err)
	}
	got, err = DownloadArgs(h, "/var/log/error.log", local)
	if err != nil || got[len(got)-2] != "[::1]:/var/log/error.log" || got[len(got)-1] != filepath.ToSlash(local) {
		t.Fatalf("DownloadArgs = %#v, %v", got, err)
	}
}

func TestLocalPathSafety(t *testing.T) {
	for _, path := range []string{"backup:daily.sql", "-option.sql", "./file with space.zip"} {
		got, err := LocalPath(path)
		if err != nil || !filepath.IsAbs(filepath.FromSlash(got)) || strings.HasPrefix(got, "-") {
			t.Fatalf("LocalPath(%q) = %q, %v", path, got, err)
		}
	}
	if runtime.GOOS == "windows" {
		got, err := LocalPath(`C:\Users\developer\My Files\file.zip`)
		if err != nil || got != "C:/Users/developer/My Files/file.zip" {
			t.Fatalf("Windows drive = %q, %v", got, err)
		}
		got, err = LocalPath(`\\server\share\file.zip`)
		if err != nil || got != "//server/share/file.zip" {
			t.Fatalf("Windows UNC = %q, %v", got, err)
		}
	}
}

func TestInvalidSCPPaths(t *testing.T) {
	h := host.Host{Alias: "dev", Hostname: "localhost", User: "root", Port: 22}
	for _, path := range []string{"", "a\nb", "a\x00b"} {
		if _, err := UploadArgs(h, path, "/tmp"); err == nil {
			t.Errorf("accepted local %q", path)
		}
		if _, err := DownloadArgs(h, path, "."); err == nil {
			t.Errorf("accepted remote %q", path)
		}
	}
}
