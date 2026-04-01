package realm

import "io"

type Marshaler interface {
	Marshal(w io.Writer, r *R) ([]byte, error)
	Unmarshal(r io.Reader) (*R, error)
}

// FileMarshaler
// MultiFileMarshaler
// FileTreeMarshaler
// SQLiteMarshaler
