package webfs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"tractor.dev/toolkit-go/engine/fs"
	"tractor.dev/toolkit-go/engine/fs/memfs"
	"tractor.dev/toolkit-go/engine/fs/watchfs"

	esbuild "github.com/evanw/esbuild/pkg/api"
)

var JSXFactory = "m"

func New(fs fs.FS) *FS {
	fsys := &FS{FS: fs}
	fsys.Transform(".jsx", transformJSX)
	fsys.Transform(".tsx", transformTSX)
	fsys.Transform(".ts", transformTSX)
	fsys.Transform(".html", transformScriptJSX)
	return fsys
}

func transformTSX(dst io.Writer, src io.Reader) error {
	b, err := ioutil.ReadAll(src)
	if err != nil {
		return err
	}
	result := esbuild.Transform(string(b), esbuild.TransformOptions{
		Loader:      esbuild.LoaderTSX,
		JSXFactory:  JSXFactory,
		JSXFragment: "",
	})
	if len(result.Errors) > 0 {
		fmt.Println(result.Errors[0], result.Errors[0].Location.File, result.Errors[0].Location.Line, result.Errors[0].Location.Column)
		return fmt.Errorf("TSX transform errors")
	}
	_, err = dst.Write(append([]byte("\n"), result.Code...))
	return err
}

func transformJSX(dst io.Writer, src io.Reader) error {
	b, err := ioutil.ReadAll(src)
	if err != nil {
		return err
	}
	result := esbuild.Transform(string(b), esbuild.TransformOptions{
		Loader:      esbuild.LoaderJSX,
		JSXFactory:  JSXFactory,
		JSXFragment: "",
	})
	if len(result.Errors) > 0 {
		fmt.Println(result.Errors[0], result.Errors[0].Location.File, result.Errors[0].Location.Line, result.Errors[0].Location.Column)
		return fmt.Errorf("JSX transform errors")
	}
	_, err = dst.Write(append([]byte("\n"), result.Code...))
	return err
}

func transformScriptJSX(dst io.Writer, src io.Reader) error {
	_, err := io.Copy(&scriptTransformWriter{Writer: dst, transform: func(b []byte) []byte {
		result := esbuild.Transform(string(b), esbuild.TransformOptions{
			Loader:      esbuild.LoaderJSX,
			JSXFactory:  JSXFactory,
			JSXFragment: "",
		})
		if len(result.Errors) > 0 {
			fmt.Println(result.Errors[0], result.Errors[0].Location.File, result.Errors[0].Location.Line, result.Errors[0].Location.Column)
		}
		return append([]byte("\n"), result.Code...)
	}}, src)
	return err
}

// BUG: this accumulates contents of all script tags
// and writes everything to each successive script tag
type scriptTransformWriter struct {
	io.Writer

	buf       bytes.Buffer
	open      bool
	transform func([]byte) []byte
}

func splitOpen(p []byte) ([]byte, []byte) {
	var empty []byte
	start := bytes.Index(p, []byte("<script"))
	if start == -1 {
		return p, empty
	}
	idx := bytes.IndexByte(p[start:], '>')
	if idx == -1 {
		return p, empty
	}
	pos := start + idx + 1
	return p[:pos], p[pos:]
}

func splitClose(p []byte) ([]byte, []byte) {
	var empty []byte
	token := []byte("</script>")
	start := bytes.Index(p, token)
	if start == -1 {
		return p, empty
	}
	return p[:start], p[start:]
}

func (tw *scriptTransformWriter) Write(p []byte) (n int, err error) {
	var nn int
	buf := p
	for n < len(p) {
		if !tw.open {
			// TODO: what if open tag happens across writes?
			before, after := splitOpen(buf)
			if len(after) == 0 {
				nn, err = tw.Writer.Write(buf)
				n += nn
				return
			}
			tw.open = true
			nn, err = tw.Writer.Write(before)
			n += nn
			if err != nil {
				return
			}
			buf = after
		}

		before, after := splitClose(buf)
		if len(after) == 0 {
			nn, err = tw.buf.Write(buf)
			n += nn
			return
		}
		tw.open = false
		nn, err = tw.buf.Write(before)
		n += nn
		if err != nil {
			return
		}
		_, err = tw.Writer.Write(tw.transform(tw.buf.Bytes()))
		if err != nil {
			return
		}
		buf = after
	}
	return
}

type FS struct {
	fs.FS
	xforms []transform
}

type transform struct {
	suffix string
	fn     func(dst io.Writer, src io.Reader) error
}

func (wfs *FS) Transform(suffix string, fn func(dst io.Writer, src io.Reader) error) {
	wfs.xforms = append(wfs.xforms, transform{
		suffix: suffix,
		fn:     fn,
	})
}

func (wfs *FS) Stat(name string) (fs.FileInfo, error) {
	return fs.Stat(wfs.FS, name)
}

func (wfs *FS) Watch(name string, cfg *watchfs.Config) (*watchfs.Watch, error) {
	_, ok := wfs.FS.(watchfs.WatchFS)
	if !ok {
		return nil, fmt.Errorf("not supported")
	}
	w, err := wfs.Watch(name, cfg)
	// if filepath.Ext(name) == "" && os.IsNotExist(err) {
	// 	exts := []string{".js", ".ts", ".tsx", ".jsx"}
	// 	for _, ext := range exts {
	// 		w, err = wfs.Watch(name+ext, cfg)
	// 		if err == nil {
	// 			break
	// 		}
	// 	}
	// }
	if err != nil {
		return nil, err
	}
	return w, err
}

func (wfs *FS) Open(name string) (fs.File, error) {
	f, err := wfs.FS.Open(name)
	// if filepath.Ext(name) == "" && os.IsNotExist(err) {
	// 	exts := []string{".js", ".ts", ".tsx", ".jsx"}
	// 	for _, ext := range exts {
	// 		f, err = wfs.Open(name + ext)
	// 		if err == nil {
	// 			break
	// 		}
	// 	}
	// }
	if err != nil {
		return nil, err
	}

	for _, xform := range wfs.xforms {
		if strings.HasSuffix(name, xform.suffix) {
			newname := name
			if xform.suffix != ".html" {
				newname = strings.Replace(name, xform.suffix, ".js", 1)
			}
			ff := memfs.NewFileHandle(memfs.CreateFile(newname))
			if err := xform.fn(ff, f); err != nil {
				return nil, err
			}
			ff.Seek(0, 0)
			return ff, nil
		}
	}
	return f, nil
}
