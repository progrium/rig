package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"tractor.dev/toolkit-go/engine/fs/watchfs/watcher"
)

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

func dialManifold() {
	conn, err := net.Dial("unix", "/var/run/manifold.sock")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	conn.Write([]byte("hello\n"))

}
