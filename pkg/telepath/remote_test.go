package telepath

import (
	"reflect"
	"testing"

	"tractor.dev/toolkit-go/duplex/codec"
	"tractor.dev/toolkit-go/duplex/fn"
	"tractor.dev/toolkit-go/duplex/mux"
	"tractor.dev/toolkit-go/duplex/talk"
)

func fatal(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func equal(t *testing.T, a, b any, message string) {
	if !reflect.DeepEqual(a, b) {
		t.Log(message)
		t.Fail()
	}
}

func makeRemote(value any) (cur Cursor, closer func()) {
	c, s := mux.Pair()
	client := talk.NewPeer(c, codec.CBORCodec{})
	server := talk.NewPeer(s, codec.CBORCodec{})
	server.Server.Handler = fn.HandlerFrom(New(value))
	go client.Respond()
	go server.Respond()
	cur = NewCursor(NewRemote(client, ""))
	closer = func() {
		client.Close()
		server.Close()
	}
	return
}

type TestRoot struct {
	Src *TestObject
	Dst *TestObject
}

type TestObject struct {
	Name string
}

func TestRemotePtrSet(t *testing.T) {
	cur, closer := makeRemote(&TestRoot{
		Src: &TestObject{Name: "ObjectA"},
		Dst: nil,
	})
	defer closer()

	names, err := cur.List()
	fatal(t, err)
	equal(t, names, []string{"Src", "Dst"}, "fields not equal")

	name, err := cur.Select("Src/Name").Value()
	fatal(t, err)
	equal(t, name, "ObjectA", "name not equal")

	fatal(t, cur.Select("Dst").Set(cur.Select("Src")))
	name, err = cur.Select("Dst/Name").Value()
	fatal(t, err)
	equal(t, name, "ObjectA", "name not equal")
}
