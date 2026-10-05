package cmd

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/RefalFalah/syrva/internal/host"
)

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func printInfo(out io.Writer, h host.Host) error {
	fmt.Fprintf(out, "%s\n\n", h.DisplayName())
	w := tabwriter.NewWriter(out, 0, 4, 3, ' ', 0)
	for _, row := range [][2]string{
		{"Alias", h.Alias}, {"Host", h.Hostname}, {"User", h.User}, {"Port", strconv.Itoa(h.Port)},
		{"SSH Key", h.IdentityFile}, {"Tags", strings.Join(h.Tags, ", ")},
		{"Working Directory", h.WorkingDirectory},
		{"Commands", strings.Join(sortedKeys(h.Commands), ", ")},
		{"Tunnels", strings.Join(sortedKeys(h.Tunnels), ", ")},
	} {
		value := row[1]
		if value == "" {
			value = "-"
		}
		fmt.Fprintf(w, "%s\t%s\n", row[0], value)
	}
	return w.Flush()
}
