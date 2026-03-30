package web

import "net"

type Listener interface {
	net.Listener

	URL() string
}
