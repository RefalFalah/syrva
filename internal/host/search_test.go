package host

import "testing"

func TestSearch(t *testing.T) {
	hosts := []Host{
		{Alias: "websku-dev", Name: "Websku Development", Hostname: "dev.example.com", User: "root", Tags: []string{"development"}},
		{Alias: "api", Name: "Équipe", Hostname: "api.example.com", User: "deploy", Tags: []string{"production"}},
	}
	for _, tt := range []struct {
		query string
		count int
		alias string
	}{
		{"", 2, "websku-dev"}, {"WSdv", 1, "websku-dev"}, {"EXAMPLE", 2, "api"},
		{"prod deploy", 1, "api"}, {"éqp", 1, "api"}, {"nothing", 0, ""},
	} {
		got := Search(hosts, tt.query)
		if len(got) != tt.count || (len(got) > 0 && got[0].Alias != tt.alias) {
			t.Errorf("Search(%q) = %#v", tt.query, got)
		}
	}
}
