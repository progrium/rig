package host

import (
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/progrium/rig/pkg/node"
)

type Subprocess struct {
	*exec.Cmd
	curr *exec.Cmd
}

func (s *Subprocess) OnEnabled() {
	cmd := *s.Cmd
	s.curr = &cmd
	if err := s.curr.Start(); err != nil {
		log.Println(err)
	}
}

func (s *Subprocess) OnDisabled() {
	if s.curr != nil && s.curr.Process != nil {
		pgid, err := syscall.Getpgid(s.curr.Process.Pid)
		s.curr = nil
		if err != nil {
			log.Println(err)
			return
		}
		syscall.Kill(-pgid, syscall.SIGTERM)
	}
}

func SubprocessNode(name string, cmd *exec.Cmd) *node.Raw {
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return node.New(name, node.Attrs{
		"enabled": "false",
		"icon":    "circle-filled",
		"color":   "",
	}, &Subprocess{Cmd: cmd})
}

type PrefixWriter struct {
	mu     sync.Mutex
	Prefix string
}

func (pw *PrefixWriter) Write(p []byte) (n int, err error) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	log.Printf("[%s] %s", pw.Prefix, string(p))
	// // Print the subprocess name as a prefix
	// if _, err := fmt.Fprintf(os.Stdout, "[%s] ", pw.Prefix); err != nil {
	// 	return 0, err
	// }

	// // Write the original content to stderr
	// return fmt.Fprint(os.Stdout, string(p))
	return 0, nil
}
