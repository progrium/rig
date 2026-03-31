package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"tractor.dev/toolkit-go/engine/cli"
	"tractor.dev/toolkit-go/engine/fs/watchfs/watcher"
	"tractor.dev/wanix/fs"
	"tractor.dev/wanix/term"
)

var Version = "dev"

func main() {
	log.SetFlags(log.Lshortfile)

	cmd := &cli.Command{
		Version: Version,
		Usage:   "rig",
		Short:   "Hi",
		Long:    `Hello world\nAgain\n\n`,
	}
	cmd.AddCommand(serveCmd())
	cmd.AddCommand(inspectCmd())

	if err := cli.Execute(context.Background(), cmd, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func allocHook(s *term.Service, rid string) error {
	r, err := s.Get(rid)
	if err != nil {
		return err
	}
	c := exec.Command("/bin/sh")
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}
	prg, err := fs.OpenFile(r, "program", os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	winch, err := r.Open("winch")
	if err != nil {
		return err
	}
	go func() {
		r := bufio.NewScanner(winch)
		for r.Scan() {
			line := strings.Split(strings.TrimSpace(r.Text()), " ")
			cols, err := strconv.ParseUint(line[0], 10, 16)
			if err != nil {
				log.Println("winch:", err)
				continue
			}
			rows, err := strconv.ParseUint(line[1], 10, 16)
			if err != nil {
				log.Println("winch:", err)
				continue
			}
			size := pty.Winsize{
				Cols: uint16(cols),
				Rows: uint16(rows),
			}
			pty.Setsize(ptmx, &size)
		}
		if err := r.Err(); err != nil {
			log.Println("winch:", err)
		}
	}()
	go func() {
		if _, err := io.Copy(prg.(io.Writer), ptmx); err != nil {
			log.Println("ptmx->prg:", err)
		}
	}()
	go func() {
		if _, err := io.Copy(ptmx, prg.(io.Reader)); err != nil {
			log.Println("prg->ptmx:", err) // todo? io.ErrClosed after program exits
		}
	}()
	go func() {
		if err := c.Wait(); err != nil {
			log.Println("cmd:", err)
		}
		if err := ptmx.Close(); err != nil {
			log.Println("ptmx:", err)
		}
		if err := prg.Close(); err != nil {
			log.Println("prg:", err)
		}
	}()
	return nil
}

func setupManifold() {
	srcPath := "/src"
	mainSrc := `package main

import (
	"log"

	"github.com/progrium/rig/pkg/runtime"
	"github.com/progrium/rig/pkg/catalog/debug"
)

func main() {
	log.Println("started")
	runtime.Run(debug.Debug{})
}
`
	os.MkdirAll(srcPath, 0755)
	if err := os.WriteFile(filepath.Join(srcPath, "main.go"), []byte(mainSrc), 0644); err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = srcPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal("tidy:", err)
	}

	rebuild := func() bool {

		log.Println("generating...")
		cmd := exec.Command("go", "generate", "./...")
		cmd.Dir = srcPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatal("generate:", err)
		}

		log.Println("building...")
		cmd = exec.Command("go", "build", "-o", "/bin/main", ".")
		cmd.Dir = srcPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Println("build:", err)
			return false
		}

		return true
	}

	reloadCh := make(chan time.Time)
	w := watcher.New(os.DirFS("/"))
	if err := w.AddRecursive(srcPath); err != nil {
		log.Fatal(err)
	}
	go func() {
		last := map[string]time.Time{}
		for event := range w.Event {
			if event.Op != watcher.Create && event.Op != watcher.Write {
				continue
			}
			// if event.Path == "meta.gen.go" {
			// 	continue
			// }
			if time.Since(last[event.Path]) < 50*time.Millisecond {
				continue
			}
			if filepath.Ext(event.Path) != ".go" {
				continue
			}
			log.Println("event:", event.Path, event.Op)
			last[event.Path] = time.Now()
			go func() {
				if rebuild() {
					reloadCh <- last[event.Path]
				}
			}()
		}
	}()
	go w.Start(100 * time.Millisecond)

	rebuild()

	// firstRun := true
	for {
		finishCh := make(chan error)
		log.Println("starting...")
		cmd := exec.Command("/bin/main")
		cmd.Dir = srcPath
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		// cmd.Env = append(os.Environ(), fmt.Sprintf("MSOCK=%s", workSock))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}

		// go func() {
		// 	var conn net.Conn
		// 	var err error
		// 	timeout := 3 * time.Second
		// 	start := time.Now()
		// 	for {
		// 		conn, err = net.Dial("unix", workSock)
		// 		if err == nil {
		// 			break
		// 		}
		// 		if time.Since(start) > timeout {
		// 			log.Println("logfeed: connection failed after retries:", err)
		// 			return
		// 		}
		// 		time.Sleep(100 * time.Millisecond)
		// 	}

		// }()

		go func() {
			finishCh <- cmd.Wait()
		}()

		select {
		case err := <-finishCh:
			if err != nil {
				log.Fatal(err)
			}
			return
		case changeTime := <-reloadCh:
			log.Println("reloading...")
			pgid, err := syscall.Getpgid(cmd.Process.Pid)
			if err != nil {
				log.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
				log.Fatal(err)
			}
			select {
			case <-ctx.Done(): // sigterm timeout
				log.Println("sending SIGKILL")
				if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
					log.Fatal(err)
				}
			case err := <-finishCh:
				if err != nil {
					log.Fatal(err)
				}
			}
			log.Println("change applied in", time.Since(changeTime))
		}

		// firstRun = false
	}
}
