package storage

import "io"

type Storage interface {
	Save(id string, r io.Reader) error
	Get(id string) (io.ReadCloser, error)
	Delete(id string) error
	List() ([]string, error)
}
