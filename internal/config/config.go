package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/RefalFalah/syrva/internal/host"
	"go.yaml.in/yaml/v3"
)

type Store struct {
	Dir string
}

type settings struct {
	Version int `yaml:"version"`
}

type hostFile struct {
	Hosts map[string]host.Host `yaml:"hosts"`
}

func DefaultDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("direktori konfigurasi user tidak ditemukan: %w", err)
	}
	return filepath.Join(dir, "syrva"), nil
}

func Open(dir string) (*Store, error) {
	if dir == "" {
		var err error
		dir, err = DefaultDir()
		if err != nil {
			return nil, err
		}
	}
	return &Store{Dir: filepath.Clean(dir)}, nil
}

// Ensure creates missing files only; existing user configuration is never replaced.
func (s *Store) Ensure() error {
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return fmt.Errorf("tidak dapat membuat direktori config %q: %w", s.Dir, err)
	}
	for _, file := range []struct{ name, content string }{
		{"config.yaml", "version: 1\n"},
		{"hosts.yaml", "hosts: {}\n"},
	} {
		path := filepath.Join(s.Dir, file.name)
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("tidak dapat membuat %q: %w", path, err)
		}
		_, writeErr := io.WriteString(f, file.content)
		closeErr := f.Close()
		if err := errors.Join(writeErr, closeErr); err != nil {
			return fmt.Errorf("tidak dapat menulis %q: %w", path, err)
		}
	}
	return nil
}

func decode(data []byte, out any) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(out); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("gunakan satu dokumen YAML saja")
	}
	return nil
}

var yamlLine = regexp.MustCompile(`line (\d+)`)

func yamlError(file string, err error) error {
	location := ""
	if match := yamlLine.FindStringSubmatch(err.Error()); len(match) == 2 {
		location = " (baris " + match[1] + ")"
	}
	// YAML type errors can contain raw scalar values. Do not echo user data or
	// accidentally pasted secrets; retain the useful filename and line instead.
	return fmt.Errorf("%s tidak valid%s. Periksa sintaks, nama field, dan tipe value; password/isi private key tidak didukung", file, location)
}

func resolveNode(node *yaml.Node) *yaml.Node {
	for node != nil && node.Kind == yaml.AliasNode {
		node = node.Alias
	}
	return node
}

func nodeHasKey(node *yaml.Node, key string) bool {
	node = resolveNode(node)
	if node == nil {
		return false
	}
	if node.Kind == yaml.SequenceNode {
		for _, child := range node.Content {
			if nodeHasKey(child, key) {
				return true
			}
		}
	}
	if node.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value == key || (node.Content[i].Value == "<<" && nodeHasKey(node.Content[i+1], key)) {
				return true
			}
		}
	}
	return false
}

func (s *Store) Load() (map[string]host.Host, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	settingsPath := filepath.Join(s.Dir, "config.yaml")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("tidak dapat membaca %q: %w", settingsPath, err)
	}
	cfg := settings{Version: 1}
	if err := decode(data, &cfg); err != nil {
		return nil, yamlError("config.yaml", err)
	}
	if cfg.Version != 1 {
		return nil, fmt.Errorf("versi config %d belum didukung (gunakan version: 1)", cfg.Version)
	}
	path := filepath.Join(s.Dir, "hosts.yaml")
	data, err = os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("tidak dapat membaca %q: %w", path, err)
	}
	file := hostFile{}
	if err := decode(data, &file); err != nil {
		return nil, yamlError("hosts.yaml", err)
	}
	if file.Hosts == nil {
		file.Hosts = make(map[string]host.Host)
	}
	// Inspect nodes only to distinguish an omitted port (22) from an invalid explicit 0.
	var nodes struct {
		Hosts map[string]yaml.Node `yaml:"hosts"`
	}
	if err := yaml.Unmarshal(data, &nodes); err != nil {
		return nil, yamlError("hosts.yaml", err)
	}
	for _, h := range host.Sorted(file.Hosts) {
		node := nodes.Hosts[h.Alias]
		resolved := resolveNode(&node)
		if resolved == nil || resolved.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("server %q harus berupa mapping YAML", h.Alias)
		}
		if !nodeHasKey(&node, "port") {
			h.Port = 22
		}
		if err := host.Validate(h); err != nil {
			return nil, fmt.Errorf("server %q: %w", h.Alias, err)
		}
		file.Hosts[h.Alias] = h
	}
	return file.Hosts, nil
}

func (s *Store) Save(hosts map[string]host.Host) error {
	for _, h := range host.Sorted(hosts) {
		if err := host.Validate(h); err != nil {
			return fmt.Errorf("server %q: %w", h.Alias, err)
		}
	}
	if err := s.Ensure(); err != nil {
		return err
	}
	if hosts == nil {
		hosts = make(map[string]host.Host)
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(hostFile{Hosts: hosts}); err != nil {
		return fmt.Errorf("tidak dapat mengubah host ke YAML: %w", err)
	}
	if err := enc.Close(); err != nil {
		return err
	}
	return atomicWrite(filepath.Join(s.Dir, "hosts.yaml"), buf.Bytes())
}

// Write in the same directory so rename stays on the same filesystem, including Windows.
func atomicWrite(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".syrva-*.tmp")
	if err != nil {
		return fmt.Errorf("tidak dapat menyimpan config: %w", err)
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(0o600); err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err == nil {
		err = os.Rename(tmp, path)
	}
	if err != nil {
		return fmt.Errorf("tidak dapat menyimpan %q (config lama tetap dipertahankan): %w", path, err)
	}
	return nil
}
