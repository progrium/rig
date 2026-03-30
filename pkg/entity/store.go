package entity

import "errors"

type Store interface {
	Resolve(id string, skip ...any) E
	Store(e E) error
	Destroy(e E) error
}

type StoreEntity interface {
	E
	SetStore(s Store) error
	GetStore() Store
}

func SetStore(v any, s Store) error {
	if e := ToEntity(v); e != nil {
		if se, ok := e.(StoreEntity); ok {
			return se.SetStore(s)
		}
	}
	return errors.ErrUnsupported
}

func GetStore(v any) Store {
	if e := ToEntity(v); e != nil {
		if se, ok := e.(StoreEntity); ok {
			return se.GetStore()
		}
	}
	return nil
}
