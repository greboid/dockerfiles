package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/greboid/osimage-install/internal/installer"
	"github.com/greboid/osimage-install/internal/target"
)

// prompter asks questions on the terminal. Every question has a
// default (from the command line, or the image), taken on Enter.
type prompter struct {
	in  *bufio.Reader
	out io.Writer
}

func newTermPrompter(stdin *os.File, out io.Writer) *prompter {
	return &prompter{in: bufio.NewReader(stdin), out: out}
}

func (p *prompter) line(q string) (string, error) {
	fmt.Fprint(p.out, q)
	s, err := p.in.ReadString('\n')
	if err != nil && !(errors.Is(err, io.EOF) && s != "") {
		return "", fmt.Errorf("reading answer: %w", err)
	}
	return strings.TrimSpace(s), nil
}

// hostname asks for a hostname; an empty answer returns def.
func (p *prompter) hostname(def string) (string, error) {
	for {
		a, err := p.line(fmt.Sprintf("Hostname [%s]: ", def))
		if err != nil {
			return "", err
		}
		if a == "" {
			return def, nil
		}
		if err := target.ValidateHostname(a); err != nil {
			fmt.Fprintln(p.out, err)
			continue
		}
		return a, nil
	}
}

// ipv6 asks for a static IPv6 address/prefix; an empty answer keeps
// def and "none" clears it.
func (p *prompter) ipv6(def *target.StaticIPv6, gateway string) (*target.StaticIPv6, error) {
	shown := "none"
	if def != nil {
		shown = def.Address.String()
		gateway = def.Gateway.String()
	}
	for {
		a, err := p.line(fmt.Sprintf("Static IPv6 address/prefix, gateway %s [%s]: ", gateway, shown))
		if err != nil {
			return nil, err
		}
		switch strings.ToLower(a) {
		case "":
			return def, nil
		case "none":
			return nil, nil
		}
		s, err := target.ParseStaticIPv6(a, gateway)
		if err != nil {
			fmt.Fprintln(p.out, err)
			continue
		}
		return s, nil
	}
}

// ipv4 asks for a static IPv4 address/prefix and its gateway; an empty
// answer keeps def and "none" clears it.
func (p *prompter) ipv4(def *target.StaticIPv4) (*target.StaticIPv4, error) {
	shown, defGW := "none", ""
	if def != nil {
		shown, defGW = def.Address.String(), def.Gateway.String()
	}
	for {
		a, err := p.line(fmt.Sprintf("Static IPv4 address/prefix, replaces DHCPv4 [%s]: ", shown))
		if err != nil {
			return nil, err
		}
		switch strings.ToLower(a) {
		case "":
			return def, nil
		case "none":
			return nil, nil
		}
		if _, err := target.ParseIPv4Address(a); err != nil {
			fmt.Fprintln(p.out, err)
			continue
		}
		q := "IPv4 gateway: "
		if defGW != "" {
			q = fmt.Sprintf("IPv4 gateway [%s]: ", defGW)
		}
		gw := ""
		for gw == "" {
			if gw, err = p.line(q); err != nil {
				return nil, err
			}
			if gw == "" {
				gw = defGW
			}
		}
		s, err := target.ParseStaticIPv4(a, gw)
		if err != nil {
			fmt.Fprintln(p.out, err)
			continue
		}
		return s, nil
	}
}

// dns asks for DNS servers and the DNSSEC mode; an empty answer keeps
// def's value, "none" clears the servers and "default" the mode.
func (p *prompter) dns(def *target.DNS) (*target.DNS, error) {
	servers, mode := "", ""
	if def != nil {
		servers, mode = def.ServerList(), def.DNSSEC
	}
	if servers == "" {
		servers = "none"
	}
	if mode == "" {
		mode = "default"
	}
	q := fmt.Sprintf("DNS servers, comma-separated [%s]: ", servers)
	for {
		a, err := p.line(q)
		if err != nil {
			return nil, err
		}
		if a == "" {
			a = servers
		}
		if strings.EqualFold(a, "none") {
			a = ""
		}
		if _, err := target.ParseDNS(a, ""); err != nil {
			fmt.Fprintln(p.out, err)
			continue
		}
		servers = a
		break
	}
	q = fmt.Sprintf("DNSSEC (%s, default) [%s]: ", strings.Join(target.DNSSECModes, ", "), mode)
	for {
		a, err := p.line(q)
		if err != nil {
			return nil, err
		}
		if a == "" {
			a = mode
		}
		if a = strings.ToLower(a); a == "default" {
			a = ""
		}
		d, err := target.ParseDNS(servers, a)
		if err != nil {
			fmt.Fprintln(p.out, err)
			continue
		}
		return d, nil
	}
}

func (p *prompter) yesNo(q string, def bool) (bool, error) {
	hint := "[y/N]"
	if def {
		hint = "[Y/n]"
	}
	for {
		a, err := p.line(q + " " + hint + " ")
		if err != nil {
			return false, err
		}
		switch strings.ToLower(a) {
		case "":
			return def, nil
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		}
		fmt.Fprintln(p.out, "answer y or n")
	}
}

// chooseDisk asks which disk to install to when none was given on the
// command line.
func chooseDisk(ctx context.Context, p *prompter, cfg *installer.Config) error {
	cands, err := installer.Candidates(ctx, cfg)
	if err != nil {
		return err
	}
	return pickDisk(p, cfg, cands)
}

// pickDisk prints the candidates (usable ones numbered, refused ones
// with the reason) and asks which disk to install to. The answer is a
// listed number, or any /dev path (the preflight checks it again
// either way).
func pickDisk(p *prompter, cfg *installer.Config, cands []installer.Candidate) error {
	var usable []installer.Candidate
	for _, d := range cands {
		if d.Usable() {
			usable = append(usable, d)
			fmt.Fprintf(p.out, " %d) %-14s %8s  %s\n", len(usable), d.Path, formatSize(d.Size), labels(d))
		} else {
			fmt.Fprintf(p.out, "    %-14s %8s  not usable: %s\n", d.Path, formatSize(d.Size), d.Reason)
		}
	}
	if len(usable) == 0 {
		return errors.New("no disk this installer can use; see the list above")
	}
	hint := fmt.Sprintf("1-%d", len(usable))
	var def string
	if len(usable) == 1 {
		hint, def = "1", "1"
	}
	for {
		a, err := p.line(fmt.Sprintf("Install to which disk? [%s]: ", hint))
		if err != nil {
			return err
		}
		if a == "" {
			a = def
		}
		if n, err := strconv.Atoi(a); err == nil && n >= 1 && n <= len(usable) {
			cfg.Disk = usable[n-1].Path
			return nil
		}
		if strings.HasPrefix(a, "/dev/") {
			cfg.Disk = a
			return nil
		}
		fmt.Fprintf(p.out, "answer %s or a /dev path\n", hint)
	}
}

// labels joins the disk's model/serial and existing content for the
// picker's list.
func labels(d installer.Candidate) string {
	parts := make([]string, 0, 2)
	if d.Ident != "" {
		parts = append(parts, d.Ident)
	}
	if d.Content != "" {
		parts = append(parts, d.Content)
	}
	return strings.Join(parts, ", ")
}

// formatSize renders a byte count for the disk list.
func formatSize(b uint64) string {
	if b >= 1<<40 {
		return fmt.Sprintf("%.1f TiB", float64(b)/(1<<40))
	}
	return fmt.Sprintf("%.0f GiB", float64(b)/(1<<30))
}
