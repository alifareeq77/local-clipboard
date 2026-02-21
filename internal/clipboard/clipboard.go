package clipboard

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

// Cmd describes a clipboard read or write command (e.g. wl-paste, xclip).
type Cmd struct {
	Name string
	Args []string
}

// Detect returns a command to read from the clipboard and optionally one to write.
// Prefers wl-clipboard, then xclip, then xsel.
func Detect() (read Cmd, write *Cmd, err error) {
	if _, err := exec.LookPath("wl-paste"); err == nil {
		if _, err := exec.LookPath("wl-copy"); err == nil {
			w := &Cmd{Name: "wl-copy"}
			return Cmd{Name: "wl-paste"}, w, nil
		}
		return Cmd{Name: "wl-paste"}, nil, nil
	}
	if _, err := exec.LookPath("xclip"); err == nil {
		w := &Cmd{Name: "xclip", Args: []string{"-selection", "clipboard"}}
		return Cmd{Name: "xclip", Args: []string{"-o", "-selection", "clipboard"}}, w, nil
	}
	if _, err := exec.LookPath("xsel"); err == nil {
		w := &Cmd{Name: "xsel", Args: []string{"--clipboard", "--input"}}
		return Cmd{Name: "xsel", Args: []string{"--clipboard", "--output"}}, w, nil
	}
	return Cmd{}, nil, errors.New("no supported clipboard command found (install wl-clipboard, xclip, or xsel)")
}

// EnsureDetect returns clipboard read/write commands, attempting to install a tool (e.g. wl-clipboard or xclip) on Linux if none is found.
func EnsureDetect() (read Cmd, write *Cmd, err error) {
	read, write, err = Detect()
	if err == nil {
		return read, write, nil
	}
	if TryInstall() {
		read, write, err = Detect()
	}
	return read, write, err
}

// Read runs the read command and returns the clipboard text.
func Read(cmd Cmd) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c := exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	out, err := c.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Write runs the write command to set the clipboard to text.
func Write(cmd *Cmd, text string) error {
	if cmd == nil {
		return errors.New("clipboard write command not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c := exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	c.Stdin = strings.NewReader(text)
	return c.Run()
}
