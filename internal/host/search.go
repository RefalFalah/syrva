package host

import (
	"sort"
	"strings"
)

// Search matches case-insensitive subsequences in local host metadata.
// Each word must match; no network calls or background indexing are involved.
func Search(hosts []Host, query string) []Host {
	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 {
		return append([]Host(nil), hosts...)
	}
	type match struct {
		host  Host
		score int
	}
	var matches []match
	for _, h := range hosts {
		fields := []string{h.Alias, h.Name, h.Hostname, h.User, strings.Join(h.Tags, " ")}
		total, matched := 0, true
		for _, word := range words {
			best := int(^uint(0) >> 1)
			for i, field := range fields {
				if score, ok := fuzzyScore(word, strings.ToLower(field)); ok {
					best = min(best, score+i*5)
				}
			}
			if best == int(^uint(0)>>1) {
				matched = false
				break
			}
			total += best
		}
		if matched {
			matches = append(matches, match{h, total})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score == matches[j].score {
			return matches[i].host.Alias < matches[j].host.Alias
		}
		return matches[i].score < matches[j].score
	})
	result := make([]Host, 0, len(matches))
	for _, match := range matches {
		result = append(result, match.host)
	}
	return result
}

func fuzzyScore(query, text string) (int, bool) {
	if query == text {
		return -100, true
	}
	if strings.HasPrefix(text, query) {
		return -60, true
	}
	if index := strings.Index(text, query); index >= 0 {
		return -30 + index, true
	}
	q := []rune(query)
	position, score, previous := 0, 0, -1
	for i, r := range []rune(text) {
		if position < len(q) && r == q[position] {
			score += i - previous - 1
			previous = i
			position++
		}
	}
	return score, position == len(q)
}
