//go:build !windows

package Console

import (
	"fmt"
	"os"
	"os/exec"
)

func stty(args ...string) error {
	cmd := exec.Command("stty", args...)
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func enterRawMode() (func(), bool) {
	for _, f := range []*os.File{os.Stdin, os.Stdout} {
		info, err := f.Stat()
		if err != nil || info.Mode()&os.ModeCharDevice == 0 {
			return nil, false
		}
	}
	if err := stty("-icanon", "-echo", "min", "1"); err != nil {
		return nil, false
	}
	return func() { stty("icanon", "echo") }, true
}

func termSize() (int, int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80, 30
	}
	var rows, cols int
	if _, err := fmt.Sscan(string(out), &rows, &cols); err != nil || cols == 0 {
		return 80, 30
	}
	return cols, rows
}
