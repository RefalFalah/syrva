package host

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

func Find(hosts map[string]Host, alias string) (Host, error) {
	if h, ok := hosts[alias]; ok {
		h.Alias = alias
		return h, nil
	}
	message := fmt.Sprintf("Server %q tidak ditemukan.", alias)
	var suggestions []string
	for _, h := range Sorted(hosts) {
		if strings.Contains(strings.ToLower(h.Alias), strings.ToLower(alias)) {
			suggestions = append(suggestions, h.Alias)
		}
	}
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}
	if len(suggestions) == 0 && len(alias) <= 128 {
		for _, h := range Search(Sorted(hosts), alias) {
			suggestions = append(suggestions, h.Alias)
			if len(suggestions) == 3 {
				break
			}
		}
	}
	if len(suggestions) == 0 && len(alias) <= 128 {
		type candidate struct {
			alias    string
			distance int
		}
		var candidates []candidate
		for _, h := range Sorted(hosts) {
			d := editDistance(strings.ToLower(alias), strings.ToLower(h.Alias))
			if d <= max(1, min(3, len(alias)/3)) {
				candidates = append(candidates, candidate{h.Alias, d})
			}
		}
		sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].distance < candidates[j].distance })
		for _, c := range candidates[:min(3, len(candidates))] {
			suggestions = append(suggestions, c.alias)
		}
	}
	if len(suggestions) > 0 {
		message += "\n\nMungkin maksud Anda:\n\n  " + strings.Join(suggestions, "\n  ")
	}
	return Host{}, errors.New(message)
}

func editDistance(a, b string) int {
	left, right := []rune(a), []rune(b)
	row := make([]int, len(right)+1)
	for j := range row {
		row[j] = j
	}
	for i, x := range left {
		previous := row[0]
		row[0] = i + 1
		for j, y := range right {
			old := row[j+1]
			cost := 0
			if x != y {
				cost = 1
			}
			row[j+1] = min(row[j]+1, old+1, previous+cost)
			previous = old
		}
	}
	return row[len(right)]
}

func Add(hosts map[string]Host, h Host) error {
	if err := Validate(h); err != nil {
		return err
	}
	if _, exists := hosts[h.Alias]; exists {
		return fmt.Errorf("server %q sudah ada; gunakan syrva edit %s", h.Alias, h.Alias)
	}
	if hosts == nil {
		return fmt.Errorf("daftar host belum diinisialisasi")
	}
	hosts[h.Alias] = h
	return nil
}

func Update(hosts map[string]Host, alias string, h Host) error {
	if _, err := Find(hosts, alias); err != nil {
		return err
	}
	if err := Validate(h); err != nil {
		return err
	}
	if _, exists := hosts[h.Alias]; exists && h.Alias != alias {
		return fmt.Errorf("server %q sudah ada", h.Alias)
	}
	delete(hosts, alias)
	hosts[h.Alias] = h
	return nil
}

func Remove(hosts map[string]Host, alias string) error {
	if _, err := Find(hosts, alias); err != nil {
		return err
	}
	delete(hosts, alias)
	return nil
}
