package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/RefalFalah/syrva/internal/host"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"
)

type Result struct {
	Alias string
	Add   bool
}

type model struct {
	hosts, filtered []host.Host
	search          textinput.Model
	searching       bool
	cursor          int
	width, height   int
	terminalFD      int
	result          Result
}

type resizeTick struct{}

func Select(ctx context.Context, in io.Reader, out io.Writer, hosts []host.Host) (Result, error) {
	input, inputOK := in.(*os.File)
	output, outputOK := out.(*os.File)
	if !inputOK || !outputOK || !term.IsTerminal(int(input.Fd())) || !term.IsTerminal(int(output.Fd())) {
		return Result{}, fmt.Errorf("TUI membutuhkan terminal interaktif. Gunakan syrva list atau syrva <host>.")
	}
	m := newModel(hosts)
	m.terminalFD = int(output.Fd())
	if width, height, err := term.GetSize(m.terminalFD); err == nil {
		m.resize(width, height)
	}
	p := tea.NewProgram(m, tea.WithContext(ctx), tea.WithInput(in), tea.WithOutput(out), tea.WithAltScreen())
	final, err := p.Run()
	if ctx.Err() != nil {
		return Result{}, ctx.Err()
	}
	if err != nil {
		return Result{}, fmt.Errorf("tidak dapat membuka TUI: %w", err)
	}
	if result, ok := final.(model); ok {
		return result.result, nil
	}
	return Result{}, fmt.Errorf("TUI tidak dapat mengembalikan pilihan server")
}

func newModel(hosts []host.Host) model {
	input := textinput.New()
	input.Prompt = "> "
	input.Placeholder = "/ untuk mencari alias, nama, host, user, tags"
	input.CharLimit = 128
	m := model{hosts: hosts, filtered: hosts, search: input, terminalFD: -1}
	m.resize(80, 24)
	return m
}

func tick() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(time.Time) tea.Msg { return resizeTick{} })
}

func (m model) Init() tea.Cmd {
	if m.terminalFD >= 0 {
		return tick()
	}
	return nil
}

func (m *model) resize(width, height int) {
	m.width, m.height = max(1, width), max(1, height)
	m.search.Width = max(1, min(m.width-8, 84))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, nil
	case resizeTick:
		// Polling also handles Windows consoles, where SIGWINCH is unavailable.
		if m.terminalFD >= 0 {
			if width, height, err := term.GetSize(m.terminalFD); err == nil {
				m.resize(width, height)
			}
			return m, tick()
		}
		return m, nil
	case tea.KeyMsg:
		// Terminal readers can group '/' and a rapidly typed/pasted query into
		// one KeyRunes message. Do not discard the search shortcut or its text.
		if !m.searching && !msg.Alt && msg.Type == tea.KeyRunes && len(msg.Runes) > 1 && msg.Runes[0] == '/' {
			m.searching = true
			focus := m.search.Focus()
			msg.Runes = msg.Runes[1:]
			updated, command := m.Update(msg)
			return updated, tea.Batch(focus, command)
		}
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if len(m.filtered) > 0 {
				m.result.Alias = m.filtered[m.cursor].Alias
				return m, tea.Quit
			}
			return m, nil
		case "up", "ctrl+p":
			m.cursor = max(0, m.cursor-1)
			return m, nil
		case "down", "ctrl+n":
			m.cursor = min(max(0, len(m.filtered)-1), m.cursor+1)
			return m, nil
		case "tab":
			m.searching = !m.searching
			if m.searching {
				command := m.search.Focus()
				return m, command
			}
			m.search.Blur()
			return m, nil
		}
		if !m.searching {
			switch msg.String() {
			case "q", "Q":
				return m, tea.Quit
			case "a", "A":
				m.result.Add = true
				return m, tea.Quit
			case "/":
				m.searching = true
				command := m.search.Focus()
				return m, command
			case "j":
				m.cursor = min(max(0, len(m.filtered)-1), m.cursor+1)
			case "k":
				m.cursor = max(0, m.cursor-1)
			}
			return m, nil
		}
	}
	old := m.search.Value()
	var command tea.Cmd
	m.search, command = m.search.Update(msg)
	if m.search.Value() != old {
		m.filtered = host.Search(m.hosts, m.search.Value())
		m.cursor = 0
	}
	return m, command
}

func truncate(value string, width int) string {
	return ansi.Truncate(value, max(1, width), "…")
}

func (m model) View() string {
	limit := lipgloss.NewStyle().MaxWidth(m.width).MaxHeight(m.height)
	if m.width < 24 || m.height < 8 {
		return limit.Render("SYRVA\nPerbesar terminal.\nQ/Esc keluar")
	}
	width := min(m.width-4, 92)
	textWidth := width - 4
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	lines := []string{
		truncate(accent.Render("SYRVA")+dim.Render("  terminal-first server management"), textWidth),
		"", "Search servers", truncate(m.search.View(), textWidth), "",
	}
	itemHeight := 4
	if m.height < 16 {
		lines = []string{accent.Render("SYRVA"), truncate(m.search.View(), textWidth)}
		itemHeight = 2
	}
	status := fmt.Sprintf("%d / %d servers", len(m.filtered), len(m.hosts))
	if len(m.filtered) > 0 {
		status += fmt.Sprintf("  •  pilihan %d", m.cursor+1)
	}
	footer := []string{dim.Render(truncate(status, textWidth))}
	if textWidth < 24 {
		footer = append(footer, "Enter SSH / Find", "A Add  Q Quit")
	} else if textWidth < 44 {
		footer = append(footer, truncate("Enter Connect  / Search", textWidth), truncate("A Add  Q/Esc Quit", textWidth))
	} else {
		footer = append(footer, "Enter Connect  / Search  A Add  Q/Esc Quit")
	}
	available := m.height - 2 - len(lines) - len(footer)
	itemHeight = min(itemHeight, max(1, available))
	pageSize := max(1, available/itemHeight)
	start := (m.cursor / pageSize) * pageSize
	end := min(len(m.filtered), start+pageSize)
	if len(m.filtered) == 0 {
		message := "Tidak ada server yang cocok."
		if len(m.hosts) == 0 {
			message = "Belum ada server. Tekan A untuk menambahkan."
		}
		lines = append(lines, dim.Render(truncate(message, textWidth)))
	}
	for i := start; i < end; i++ {
		h := m.filtered[i]
		prefix, name := "  ", h.DisplayName()
		if h.Name != "" && h.Name != h.Alias {
			name += " [" + h.Alias + "]"
		}
		selected := i == m.cursor
		if selected {
			prefix = "› "
			name = accent.Render(truncate(prefix+name, textWidth))
		} else {
			name = truncate(prefix+name, textWidth)
		}
		lines = append(lines, name)
		if itemHeight > 1 {
			metadata := "  " + h.User + "@" + h.Hostname
			if itemHeight < 4 && len(h.Tags) > 0 {
				metadata += " • " + strings.Join(h.Tags, " • ")
			}
			lines = append(lines, dim.Render(truncate(metadata, textWidth)))
		}
		if itemHeight == 4 {
			lines = append(lines, dim.Render(truncate("  "+strings.Join(h.Tags, " • "), textWidth)), "")
		} else if itemHeight == 3 {
			lines = append(lines, dim.Render(truncate("  "+strings.Join(h.Tags, " • "), textWidth)))
		}
	}
	lines = append(lines, footer...)
	frame := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1).Width(width)
	return limit.Render(frame.Render(strings.Join(lines, "\n")))
}
