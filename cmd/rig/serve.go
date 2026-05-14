package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/creack/pty"
	"github.com/google/uuid"
	"github.com/progrium/rig/pkg/webfs"
	"github.com/progrium/rig/web"
	"golang.org/x/net/websocket"
	"tractor.dev/toolkit-go/duplex/mux"
	"tractor.dev/toolkit-go/engine/cli"
	"tractor.dev/wanix"
	"tractor.dev/wanix/fs"
	"tractor.dev/wanix/fs/localfs"
	"tractor.dev/wanix/term"
	"tractor.dev/wanix/vfs/pipe"
	"tractor.dev/wanix/vfs/ramfs"
	"tractor.dev/wanix/web/api"
)

func serveCmd() *cli.Command {
	cmd := &cli.Command{
		Usage: "serve",
		Short: "",
		Long:  ``,
		Run:   serve,
	}
	return cmd
}

func serve(ctx *cli.Context, args []string) {
	ts := term.New()
	ts.AllocHook = allocHook

	k := wanix.New()
	k.AddModule("#pipe", &pipe.Allocator{})
	k.AddModule("#ramfs", &ramfs.Allocator{})
	k.AddModule("#term", ts)

	root, err := k.NewRoot()
	if err != nil {
		log.Fatal(err)
	}

	root.Bind("#term", "term")

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	lfsys, err := localfs.New(cwd)
	if err != nil {
		log.Fatal(err)
	}
	root.Namespace().Bind(lfsys, ".", "root")

	token := uuid.New().String()
	if err := fs.WriteFile(lfsys, "etc/token", []byte(token+"\n"), 0644); err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.URL.Path == "/" {
			w.Header().Set("Set-Cookie", fmt.Sprintf("token=%s", token))
		}
		http.FileServerFS(web.FS).ServeHTTP(w, r)
	})
	http.HandleFunc("/editors/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		editors, err := fs.Sub(web.FS, "editors")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.StripPrefix("/editors/", http.FileServerFS(webfs.New(editors))).ServeHTTP(w, r)
	})

	http.HandleFunc(fmt.Sprintf("/fsys/%s", token), func(w http.ResponseWriter, r *http.Request) {
		websocket.Handler(func(ws *websocket.Conn) {
			ws.PayloadType = websocket.BinaryFrame
			sess := mux.New(ws)
			defer sess.Close()
			api.PortResponder(sess, root)
		}).ServeHTTP(w, r)
	})

	inspectorTarget, err := url.Parse("http://localhost:11010")
	if err != nil {
		log.Fatal(err)
	}
	inspectorProxy := httputil.NewSingleHostReverseProxy(inspectorTarget)
	inspectorProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "proxy connection error", http.StatusBadGateway)
	}
	http.Handle(fmt.Sprintf("/inspector/%s", token), inspectorProxy)

	go setupManifold()

	fmt.Println("Serving web directory on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
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
