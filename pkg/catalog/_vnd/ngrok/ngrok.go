package ngrok

import (
	"context"
	"os"
	"strings"

	"github.com/progrium/rig/pkg/catalog/web"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

type Listener struct {
	ngrok.Tunnel
}

func (l *Listener) Activate(ctx context.Context) (err error) {
	l.Tunnel, err = ngrok.Listen(ctx,
		config.HTTPEndpoint(),
		ngrok.WithAuthtoken(os.Getenv("NGROK_TOKEN")),
	)
	return
}

func (l *Listener) Deactivate(ctx context.Context) error {
	if l.Tunnel != nil {
		err := l.Tunnel.Close()
		if err != nil && strings.Contains(err.Error(), "use of closed network connection") {
			return nil
		}
		return err
	}
	return nil
}

func (l *Listener) Provides() web.Listener {
	return l
}
