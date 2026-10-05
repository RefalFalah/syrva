package host

import (
	"strings"
	"testing"
)

func TestCRUD(t *testing.T) {
	hosts := map[string]Host{}
	h := Host{Alias: "dev", Hostname: "localhost", User: "root", Port: 22}
	if err := Add(hosts, h); err != nil {
		t.Fatal(err)
	}
	if err := Add(hosts, h); err == nil {
		t.Fatal("duplicate accepted")
	}
	h.Alias = "renamed"
	if err := Update(hosts, "dev", h); err != nil {
		t.Fatal(err)
	}
	if _, exists := hosts["dev"]; exists {
		t.Fatal("old alias retained")
	}
	if err := Remove(hosts, "renamed"); err != nil || len(hosts) != 0 {
		t.Fatalf("Remove = %v", err)
	}
	if err := Remove(hosts, "unknown"); err == nil {
		t.Fatal("unknown removed")
	}
}

func TestFindSuggestions(t *testing.T) {
	hosts := map[string]Host{"websku-dev": {}, "websku-prod": {}}
	_, err := Find(hosts, "websku")
	if err == nil || !strings.Contains(err.Error(), "websku-dev") || !strings.Contains(err.Error(), "tidak ditemukan") {
		t.Fatalf("Find error: %v", err)
	}
	_, err = Find(hosts, "websku-deb")
	if err == nil || !strings.Contains(err.Error(), "websku-dev") {
		t.Fatalf("typo suggestion = %v", err)
	}
}

func TestUpdateConflictKeepsOldHost(t *testing.T) {
	h := Host{Alias: "one", Hostname: "localhost", User: "root", Port: 22}
	hosts := map[string]Host{"one": h, "two": h}
	h.Alias = "two"
	if Update(hosts, "one", h) == nil || len(hosts) != 2 || hosts["one"].Alias != "one" {
		t.Fatal("rename replaced existing host")
	}
}
