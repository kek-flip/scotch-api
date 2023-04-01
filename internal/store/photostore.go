package store

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type PhotoStore struct {
	path string
}

var errDirNotExist = errors.New("directory does not exist")

func NewPhotoStore(path string) (*PhotoStore, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errDirNotExist
	}

	return &PhotoStore{path}, nil
}

func (ps *PhotoStore) Create(photo []byte, id int) error {
	fullFileName := fmt.Sprintf("%s/%d.jpeg", ps.path, id)
	pf, err := os.Create(fullFileName)
	if err != nil {
		return err
	}
	defer pf.Close()

	if http.DetectContentType(photo) != "image/jpeg" {
		return errors.New("not jpeg image")
	}

	n, err := pf.Write(photo)
	if len(photo) != n {
		os.Remove(fullFileName)
		return err
	}

	return nil
}

func (ps *PhotoStore) FindById(id int) ([]byte, error) {
	fullFileName := fmt.Sprintf("%s/%d.jpeg", ps.path, id)
	pf, err := os.OpenFile(fullFileName, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer pf.Close()

	var p bytes.Buffer
	pw := bufio.NewWriter(&p)
	_, err = io.Copy(pw, pf)
	if err != nil {
		return nil, err
	}

	return p.Bytes(), nil
}

func (ps *PhotoStore) DeleteByName(fileName string) error {
	fullFileName := fmt.Sprintf("%s/%s.jpeg", ps.path, fileName)
	return os.Remove(fullFileName)
}
