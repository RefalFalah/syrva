package ssh

import (
	"bufio"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/RefalFalah/syrva/internal/host"
)

type configBlock struct {
	patterns []string
	values   map[string]string
	disabled bool
}

// ParseConfig reads the five MVP directives without executing OpenSSH, Include,
// Match exec, or any other local command. Wildcards supply defaults, not hosts.
func ParseConfig(reader io.Reader, defaultUser string) ([]host.Host, []string, error) {
	blocks := []configBlock{{patterns: []string{"*"}, values: map[string]string{}}}
	aliases := map[string]bool{}
	var warnings []string
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if lineNumber == 1 {
			line = strings.TrimPrefix(line, "\ufeff")
		}
		key, values, err := parseLine(line)
		if err != nil {
			return nil, warnings, fmt.Errorf("SSH config baris %d: %w", lineNumber, err)
		}
		if key == "" {
			continue
		}
		current := &blocks[len(blocks)-1]
		switch key {
		case "host":
			if len(values) == 0 {
				return nil, warnings, fmt.Errorf("SSH config baris %d: Host membutuhkan alias", lineNumber)
			}
			blocks = append(blocks, configBlock{patterns: values, values: map[string]string{}})
			for _, alias := range values {
				if strings.HasPrefix(alias, "!") || strings.ContainsAny(alias, "*?[") {
					continue
				}
				if err := host.ValidateAlias(alias); err != nil {
					warnings = append(warnings, fmt.Sprintf("baris %d: alias %q dilewati: %v", lineNumber, alias, err))
					continue
				}
				aliases[alias] = true
			}
		case "match":
			blocks = append(blocks, configBlock{disabled: true, values: map[string]string{}})
			warnings = append(warnings, fmt.Sprintf("baris %d: blok Match dilewati sampai Host berikutnya", lineNumber))
		case "hostname", "user", "port", "identityfile":
			if current.disabled {
				continue
			}
			if len(values) != 1 || values[0] == "" {
				return nil, warnings, fmt.Errorf("SSH config baris %d: %s membutuhkan satu value (quote path berspasi)", lineNumber, key)
			}
			if _, set := current.values[key]; !set {
				current.values[key] = values[0]
			} else if key == "identityfile" {
				warnings = append(warnings, fmt.Sprintf("baris %d: hanya IdentityFile pertama yang diimpor", lineNumber))
			}
		case "include", "proxyjump", "proxycommand":
			warnings = append(warnings, fmt.Sprintf("baris %d: %s tidak diimpor; MVP hanya menyalin Host/HostName/User/Port/IdentityFile", lineNumber, key))
		default:
			// Other native SSH options are intentionally outside the import model.
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, warnings, fmt.Errorf("tidak dapat membaca SSH config: %w", err)
	}
	keys := make([]string, 0, len(aliases))
	for alias := range aliases {
		keys = append(keys, alias)
	}
	sort.Strings(keys)
	hosts := make([]host.Host, 0, len(keys))
	for _, alias := range keys {
		values := map[string]string{}
		for _, block := range blocks {
			if block.disabled || !matches(block.patterns, alias) {
				continue
			}
			for key, value := range block.values {
				if _, set := values[key]; !set {
					values[key] = value
				}
			}
		}
		h := host.Host{Alias: alias, Name: alias, Hostname: alias, User: defaultUser, Port: 22}
		if value, set := values["hostname"]; set {
			h.Hostname = strings.ReplaceAll(value, "%h", alias)
		}
		if value, set := values["user"]; set {
			h.User = value
		}
		if value, set := values["port"]; set {
			port, err := strconv.Atoi(value)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("server %q dilewati: port tidak berupa angka", alias))
				continue
			}
			h.Port = port
		}
		h.IdentityFile = values["identityfile"]
		if strings.EqualFold(h.IdentityFile, "none") {
			// Empty would re-enable native default identities: do not silently change meaning.
			warnings = append(warnings, fmt.Sprintf("server %q dilewati: IdentityFile none tidak didukung model MVP", alias))
			continue
		}
		if err := host.Validate(h); err != nil {
			warnings = append(warnings, fmt.Sprintf("server %q dilewati: %v", alias, err))
			continue
		}
		hosts = append(hosts, h)
	}
	return hosts, warnings, nil
}

func matches(patterns []string, alias string) bool {
	matched := false
	for _, pattern := range patterns {
		negated := strings.HasPrefix(pattern, "!")
		pattern = strings.TrimPrefix(pattern, "!")
		ok, err := path.Match(strings.ToLower(pattern), strings.ToLower(alias))
		if err != nil || !ok {
			continue
		}
		if negated {
			return false
		}
		matched = true
	}
	return matched
}

func parseLine(line string) (string, []string, error) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", nil, nil
	}
	end := strings.IndexFunc(line, func(r rune) bool { return unicode.IsSpace(r) || r == '=' })
	if end < 0 {
		return strings.ToLower(line), nil, nil
	}
	key := strings.ToLower(line[:end])
	rest := strings.TrimSpace(line[end:])
	rest = strings.TrimSpace(strings.TrimPrefix(rest, "="))
	var values []string
	var token strings.Builder
	var quote rune
	present := false
	runes := []rune(rest)
	flush := func() {
		if present {
			values = append(values, token.String())
			token.Reset()
			present = false
		}
	}
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\\' && i+1 < len(runes) {
			next := runes[i+1]
			if next == '\'' || next == '"' || next == '#' || unicode.IsSpace(next) {
				token.WriteRune(next)
				present = true
				i++
				continue
			}
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				token.WriteRune(r)
			}
			continue
		}
		switch {
		case r == '#':
			flush()
			return key, values, nil
		case r == '\'' || r == '"':
			quote, present = r, true
		case unicode.IsSpace(r):
			flush()
		default:
			token.WriteRune(r)
			present = true
		}
	}
	if quote != 0 {
		return "", nil, fmt.Errorf("quote tidak ditutup")
	}
	flush()
	return key, values, nil
}
