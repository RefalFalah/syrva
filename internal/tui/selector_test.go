package tui

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/RefalFalah/syrva/internal/host"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func key(m model, value tea.KeyMsg) model {
	updated, _ := m.Update(value)
	return updated.(model)
}

func text(value string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)}
}

func TestSelectionAndSearch(t *testing.T) {
	m := newModel([]host.Host{{Alias: "web", Name: "Web"}, {Alias: "api", Tags: []string{"production"}}})
	m = key(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 1 {
		t.Fatal("navigation failed")
	}
	m = key(m, text("/"))
	m = key(m, text("prod"))
	if len(m.filtered) != 1 || m.filtered[0].Alias != "api" {
		t.Fatal("search failed")
	}
	m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.result.Alias != "api" {
		t.Fatal("selection failed")
	}
}

func TestGroupedSearchShortcut(t *testing.T) {
	m := newModel([]host.Host{{Alias: "web"}, {Alias: "api", Tags: []string{"production"}}})
	m = key(m, text("/prod"))
	if !m.searching || m.search.Value() != "prod" || len(m.filtered) != 1 || m.filtered[0].Alias != "api" {
		t.Fatalf("grouped search was lost: %#v", m.filtered)
	}
}

func TestUnicodeMetadataIsPreserved(t *testing.T) {
	m := newModel([]host.Host{{Alias: "dev", Name: "日本 server", User: "root", Hostname: "localhost", Tags: []string{"développement", "团队"}}})
	view := m.View()
	for _, text := range []string{"› 日本 server", "développement", "团队", " • "} {
		if !strings.Contains(view, text) {
			t.Fatalf("Unicode metadata %q was lost in %q", text, view)
		}
	}
	if got := truncate("› sélection • 队伍", 80); got != "› sélection • 队伍" {
		t.Fatalf("Unicode truncate changed text: %q", got)
	}
}

func TestEmptyAndResize(t *testing.T) {
	for _, hosts := range [][]host.Host{nil, {{Alias: strings.Repeat("界", 100), Name: strings.Repeat("é", 100)}}} {
		m := newModel(hosts)
		for _, size := range [][2]int{{0, 0}, {1, 1}, {20, 5}, {24, 8}, {30, 10}, {80, 24}, {150, 60}} {
			updated, _ := m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
			m = updated.(model)
			view := m.View()
			if lipgloss.Width(view) > max(1, size[0]) || lipgloss.Height(view) > max(1, size[1]) {
				t.Fatalf("view exceeds %v: %dx%d", size, lipgloss.Width(view), lipgloss.Height(view))
			}
			if size[0] >= 24 && size[1] >= 8 && !strings.Contains(view, "Quit") {
				t.Fatalf("quit help clipped at %v: %q", size, view)
			}
			m = key(m, tea.KeyMsg{Type: tea.KeyDown})
			m = key(m, tea.KeyMsg{Type: tea.KeyEnter})
		}
	}
}

func TestBubbleTeaProgramQuit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	program := tea.NewProgram(newModel(nil), tea.WithContext(ctx), tea.WithInput(strings.NewReader("q")), tea.WithOutput(io.Discard), tea.WithoutRenderer(), tea.WithoutSignalHandler())
	final, err := program.Run()
	if err != nil || final.(model).result.Alias != "" {
		t.Fatalf("program quit = %v", err)
	}
}

func TestQuitAndAdd(t *testing.T) {
	for _, k := range []tea.KeyMsg{text("q"), {Type: tea.KeyEsc}, {Type: tea.KeyCtrlC}} {
		updated, command := newModel(nil).Update(k)
		if command == nil || updated.(model).result.Alias != "" {
			t.Fatal("quit failed")
		}
	}
	if m := key(newModel(nil), text("A")); !m.result.Add {
		t.Fatal("add failed")
	}
}

func TestNonInteractiveSelector(t *testing.T) {
	_, err := Select(context.Background(), strings.NewReader("q"), io.Discard, nil)
	if err == nil || !strings.Contains(err.Error(), "terminal interaktif") {
		t.Fatalf("Select = %v", err)
	}
}
