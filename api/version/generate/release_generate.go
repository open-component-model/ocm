package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/mandelsoft/filepath/pkg/filepath"
)

func Error(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(1)
}

func main() {
	args := os.Args[1:]
	nodev := false
	for len(args) > 0 {
		if !strings.HasPrefix(args[0], "-") {
			break
		}
		switch args[0] {
		case "--no-dev":
			nodev = true
		default:
			Error(fmt.Errorf("ilvalid option %q", args[0]))
		}
		args = args[1:]
	}
	if len(args) == 0 {
		Error(fmt.Errorf("missing argument"))
	}
	cmd := args[0]

	pre := ""
	if len(args) > 1 {
		pre = args[1]
	}
	vers_file := "VERSION"

	var data []byte
	dir := "."
	verpath := ""
	err := os.ErrClosed
	for err != nil {
		dir, err = filepath.Abs(dir)
		if err == nil {
			if !filepath.IsRoot(dir) {
				verpath = filepath.Join(dir, vers_file)
				data, err = os.ReadFile(verpath)
			} else {
				err = fmt.Errorf("no %q file found", vers_file)
			}
		}
		dir += "/.."
	}
	if err != nil {
		Error(fmt.Errorf("cannot read version file %q: %w", vers_file, err))
	}
	raw := strings.TrimSpace(string(data))
	v := raw
	if i := strings.Index(raw, "-"); i >= 0 {
		v = raw[:i]
		found := raw[i+1:]
		if pre == "" {
			pre = found
		}
	}

	if nodev && pre == "dev" {
		Error(fmt.Errorf("dev release not possible"))
	}

	nonpre := semver.MustParse(v)
	if pre != "" {
		_ = semver.MustParse(v + "-" + pre)
	}

	v = strings.TrimPrefix(v, "v")

	//nolint:forbidigo // Logger not needed for this command.
	switch cmd {
	case "print-semver":
		fmt.Print(nonpre)
	case "print-major-minor":
		fmt.Printf("%d.%d", nonpre.Major(), nonpre.Minor())
	case "print-version":
		fmt.Print(v)
	case "print-rc-version":
		if pre == "" {
			fmt.Print(v)
		} else {
			fmt.Printf("%s-%s", v, pre)
		}
	case "bump-minor":
		next := nonpre.IncMinor().String() + "-dev"
		fmt.Printf("%s", next)
	case "bump-patch":
		next := nonpre.IncPatch().String() + "-dev"
		fmt.Printf("%s", next)
	default:
		Error(fmt.Errorf("invalid command %q", cmd))
	}
}
