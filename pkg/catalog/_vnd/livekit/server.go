package livekit

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/progrium/rig/pkg/catalog/host"
)

type Server struct {
	cmd *exec.Cmd
}

func (s *Server) Initialize() {
	// ws at ws://localhost:7880
	// apikey is devkey
	// secretkey is secret
	config := fmt.Sprintf(`
bind_addresses:
  - "0.0.0.0"
`)
	cmdpath := "livekit-server"
	binpath := os.Getenv("BINPATH")
	if binpath != "" {
		cmdpath = filepath.Join(binpath, cmdpath)
	}
	cmd := exec.Command(cmdpath, "--config-body", config, "--dev")
	cmd.Stdout = &host.PrefixWriter{Prefix: "livekit-server"}
	cmd.Stderr = &host.PrefixWriter{Prefix: "livekit-server"}
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.Env = os.Environ()
	if binpath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", strings.TrimRight(binpath, "/")))
	} else {
		cmd.Env = append(cmd.Env, "PATH=/opt/homebrew/bin")
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	s.cmd = cmd
}

func (s *Server) Activate(ctx context.Context) (err error) {
	if err := s.cmd.Start(); err != nil {
		return err
	}
	return
}

func (s *Server) Deactivate(ctx context.Context) error {
	if s.cmd != nil && s.cmd.Process != nil {
		pgid, err := syscall.Getpgid(s.cmd.Process.Pid)
		s.cmd = nil
		if err != nil {
			return err
		}
		syscall.Kill(-pgid, syscall.SIGTERM)
	}
	return nil
}
