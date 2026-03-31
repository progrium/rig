package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/google/uuid"
	"github.com/progrium/rig/net/tcp"
	"github.com/progrium/rig/web"
	"golang.org/x/net/websocket"
	"tractor.dev/toolkit-go/duplex/mux"
	"tractor.dev/toolkit-go/engine/cli"
	"tractor.dev/toolkit-go/engine/fs"
	"tractor.dev/wanix"
	"tractor.dev/wanix/fs/localfs"
	"tractor.dev/wanix/term"
	"tractor.dev/wanix/vfs/pipe"
	"tractor.dev/wanix/vfs/ramfs"
	"tractor.dev/wanix/web/api"
)

func serveCmd() *cli.Command {
	cmd := &cli.Command{
		Usage: "serve",
		Short: "Hi",
		Long:  `Hello world\nAgain\n\n`,
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
	k.AddModule("#net/tcp", tcp.New())

	root, err := k.NewRoot()
	if err != nil {
		log.Fatal(err)
	}

	root.Bind("#term", "term")
	root.Bind("#net", "net")

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
	fs := http.FileServerFS(web.FS)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Set-Cookie", fmt.Sprintf("token=%s", token))
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fs.ServeHTTP(w, r)
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
