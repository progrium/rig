package signal_test

import (
	"fmt"
	"testing"

	"github.com/progrium/rig/pkg/signal"
)

type Node struct {
	Value string

	signal.Dispatcher[Node]
}

func TestSignalSend(t *testing.T) {
	n := &Node{Value: "Foo"}
	ch := make(chan string, 2)
	n.Watch(signal.Func[Node](func(s signal.Signal[Node]) {
		ch <- fmt.Sprintf("%s: %s", s.Receiver.(*Node).Value, s.Name)
	}))
	signal.Send(n, "EventA")
	signal.Send(n, "EventB")
	if v := <-ch; v != "Foo: EventA" {
		t.Fatalf("unexpected watch output: %s", v)
	}
	if v := <-ch; v != "Foo: EventB" {
		t.Fatalf("unexpected watch output: %s", v)
	}
}
