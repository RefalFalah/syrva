package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/RefalFalah/syrva/internal/host"
	"github.com/spf13/cobra"
)

type prompt struct {
	reader *bufio.Reader
	out    io.Writer
	ctx    context.Context
}

func newPrompt(cmd *cobra.Command) *prompt {
	return &prompt{reader: bufio.NewReader(cmd.InOrStdin()), out: cmd.OutOrStdout(), ctx: cmd.Context()}
}

func (p *prompt) line(label string) (string, error) {
	if err := p.ctx.Err(); err != nil {
		return "", err
	}
	if _, err := fmt.Fprint(p.out, label); err != nil {
		return "", err
	}
	type result struct {
		line string
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		line, err := p.reader.ReadString('\n')
		ch <- result{line, err}
	}()
	select {
	case <-p.ctx.Done():
		return "", p.ctx.Err()
	case r := <-ch:
		if r.err != nil && !(errors.Is(r.err, io.EOF) && len(r.line) > 0) {
			return "", fmt.Errorf("input berakhir; perubahan dibatalkan")
		}
		return strings.TrimSpace(r.line), nil
	}
}

func (p *prompt) ask(label, value string, optional bool) (string, error) {
	if value != "" {
		label += " [" + value + "]"
	}
	input, err := p.line(label + ": ")
	if err != nil {
		return "", err
	}
	if input == "" {
		return value, nil
	}
	if optional && input == "-" {
		return "", nil
	}
	return input, nil
}

func (p *prompt) confirm(label string, defaultYes bool) (bool, error) {
	suffix := " [y/N] "
	if defaultYes {
		suffix = " [Y/n] "
	}
	value, err := p.line(label + suffix)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(value) {
	case "":
		return defaultYes, nil
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("jawab y atau n; perubahan dibatalkan")
	}
}

func readHost(cmd *cobra.Command, current host.Host) (host.Host, error) {
	p := newPrompt(cmd)
	fmt.Fprintln(p.out, "Enter mempertahankan default. Ketik '-' untuk mengosongkan field opsional.")
	for _, field := range []struct {
		label    string
		value    *string
		optional bool
	}{
		{"Alias", &current.Alias, false},
		{"Nama Server", &current.Name, true},
		{"Hostname / IP", &current.Hostname, false},
		{"Username", &current.User, false},
	} {
		value, err := p.ask(field.label, *field.value, field.optional)
		if err != nil {
			return host.Host{}, err
		}
		*field.value = value
	}
	port, err := p.ask("Port", strconv.Itoa(current.Port), false)
	if err != nil {
		return host.Host{}, err
	}
	current.Port, err = strconv.Atoi(port)
	if err != nil {
		return host.Host{}, fmt.Errorf("port harus berupa angka antara 1 dan 65535")
	}
	current.IdentityFile, err = p.ask("SSH Key (path; kosong = SSH Agent/OpenSSH)", current.IdentityFile, true)
	if err != nil {
		return host.Host{}, err
	}
	tags, err := p.ask("Tags (pisahkan dengan koma)", strings.Join(current.Tags, ", "), true)
	if err != nil {
		return host.Host{}, err
	}
	current.Tags = nil
	seen := map[string]bool{}
	for _, tag := range strings.Split(tags, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" && !seen[tag] {
			current.Tags = append(current.Tags, tag)
			seen[tag] = true
		}
	}
	current.WorkingDirectory, err = p.ask("Working Directory", current.WorkingDirectory, true)
	if err != nil {
		return host.Host{}, err
	}
	if err := host.Validate(current); err != nil {
		return host.Host{}, err
	}
	return current, nil
}
