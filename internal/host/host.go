package host

import (
	"fmt"
	"net"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

type Host struct {
	Alias            string             `yaml:"-"`
	Name             string             `yaml:"name,omitempty"`
	Hostname         string             `yaml:"host"`
	User             string             `yaml:"user"`
	Port             int                `yaml:"port"`
	IdentityFile     string             `yaml:"identity_file,omitempty"`
	Tags             []string           `yaml:"tags,omitempty"`
	WorkingDirectory string             `yaml:"working_directory,omitempty"`
	Commands         map[string]Command `yaml:"commands,omitempty"`
	Tunnels          map[string]Tunnel  `yaml:"tunnels,omitempty"`
}

type Command struct {
	Description string `yaml:"description,omitempty"`
	Command     string `yaml:"command"`
}

type Tunnel struct {
	LocalPort  int    `yaml:"local_port"`
	RemoteHost string `yaml:"remote_host"`
	RemotePort int    `yaml:"remote_port"`
}

var aliasPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
var userPattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.@\\-]*\$?$`)
var hostnamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

func ValidateAlias(alias string) error {
	if len(alias) > 128 || !aliasPattern.MatchString(alias) {
		return fmt.Errorf("alias harus dimulai dengan huruf/angka dan hanya berisi huruf, angka, titik, '-' atau '_' (maksimal 128 karakter)")
	}
	switch strings.ToLower(alias) {
	case "add", "edit", "remove", "list", "import", "help", "completion", "version":
		return fmt.Errorf("alias %q digunakan oleh command Syrva", alias)
	}
	return nil
}

func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port harus antara 1 dan 65535")
	}
	return nil
}

func ValidateHostname(value string) error {
	hostname := value
	if strings.HasPrefix(value, "[") || strings.HasSuffix(value, "]") {
		if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
			return fmt.Errorf("kurung IPv6 tidak lengkap")
		}
		hostname = value[1 : len(value)-1]
	}
	if net.ParseIP(hostname) != nil {
		return nil
	}
	// IPv6 link-local zone identifiers are accepted by native OpenSSH.
	if ip, zone, ok := strings.Cut(hostname, "%"); ok && net.ParseIP(ip) != nil {
		if strings.Contains(ip, ":") && hostnamePattern.MatchString(zone) {
			return nil
		}
	}
	if len(value) > 253 || !hostnamePattern.MatchString(value) {
		return fmt.Errorf("hostname/IP tidak valid; gunakan hostname atau IP tanpa username, URL, spasi, atau karakter shell")
	}
	return nil
}

func hasControl(value string) bool {
	return strings.ContainsFunc(value, unicode.IsControl)
}

func Validate(h Host) error {
	if err := ValidateAlias(h.Alias); err != nil {
		return err
	}
	if err := ValidateHostname(h.Hostname); err != nil {
		return err
	}
	if !userPattern.MatchString(h.User) {
		return fmt.Errorf("username wajib diisi dan tidak boleh berisi spasi atau karakter shell")
	}
	if err := ValidatePort(h.Port); err != nil {
		return err
	}
	if hasControl(h.Name) || hasControl(h.IdentityFile) || hasControl(h.WorkingDirectory) {
		return fmt.Errorf("nama, path SSH key, dan working directory tidak boleh berisi karakter kontrol")
	}
	if strings.Contains(h.IdentityFile, "-----BEGIN ") && strings.Contains(h.IdentityFile, "PRIVATE KEY-----") {
		return fmt.Errorf("SSH Key harus berupa path file, bukan isi private key")
	}
	for _, tag := range h.Tags {
		if strings.TrimSpace(tag) == "" || hasControl(tag) {
			return fmt.Errorf("tag tidak boleh kosong atau berisi karakter kontrol")
		}
	}
	for name, command := range h.Commands {
		if !aliasPattern.MatchString(name) || strings.TrimSpace(command.Command) == "" {
			return fmt.Errorf("preset %q: nama atau command tidak valid", name)
		}
		if strings.ContainsRune(command.Command, '\x00') || hasControl(command.Description) {
			return fmt.Errorf("preset %q berisi karakter yang tidak valid", name)
		}
	}
	for name, tunnel := range h.Tunnels {
		if !aliasPattern.MatchString(name) {
			return fmt.Errorf("nama tunnel %q tidak valid", name)
		}
		if err := ValidateTunnel(tunnel); err != nil {
			return fmt.Errorf("tunnel %q: %w", name, err)
		}
	}
	return nil
}

func ValidateTunnel(t Tunnel) error {
	if err := ValidatePort(t.LocalPort); err != nil {
		return fmt.Errorf("local_port: %w", err)
	}
	if err := ValidatePort(t.RemotePort); err != nil {
		return fmt.Errorf("remote_port: %w", err)
	}
	return ValidateHostname(t.RemoteHost)
}

func (h Host) DisplayName() string {
	if h.Name != "" {
		return h.Name
	}
	return h.Alias
}

func Sorted(hosts map[string]Host) []Host {
	aliases := make([]string, 0, len(hosts))
	for alias := range hosts {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	result := make([]Host, 0, len(aliases))
	for _, alias := range aliases {
		h := hosts[alias]
		h.Alias = alias
		result = append(result, h)
	}
	return result
}
