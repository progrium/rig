package misc

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// Must takes a value and an error, and panics if the error is not nil.
// Otherwise, it returns the value.
func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func URLJoin(base string, parts ...string) string {
	base = strings.TrimRight(base, "/")
	for i := range parts {
		parts[i] = strings.Trim(parts[i], "/")
	}
	parts = append([]string{base}, parts...)
	return strings.Join(parts, "/")
}

func ListenAddr() string {
	ln, err := net.Listen("tcp4", ":0")
	if err != nil {
		panic(err)
	}
	ln.Close()
	return ln.Addr().String()
}

func DialUntilOpen(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout exceeded trying to connect")
			}
			time.Sleep(time.Second)
			continue
		}
		defer conn.Close()
		break
	}
	return nil
}

func TryUntilSuccess(timeout time.Duration, interval time.Duration, fn func() error) error {
	deadline := time.Now().Add(timeout)
	for {
		err := fn()
		if err != nil {
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout exceeded")
			}
			time.Sleep(interval)
			continue
		}
		break
	}
	return nil
}
