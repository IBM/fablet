package util

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func UnTar(source []byte, target string) error {
	// Use file as source
	// sourceFile, err := os.Open(source)
	// if err != nil {
	// 	return err
	// }
	// defer sourceFile.Close()

	// Use reader as source
	sourceReader := bytes.NewReader(source)

	// gzipReader, err := gzip.NewReader(sourceFile)
	// if err != nil {
	// 	return err
	// }

	tarReader := tar.NewReader(sourceReader)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		info := hdr.FileInfo()

		if info.IsDir() {
			if err = os.MkdirAll(filepath.Join(target, hdr.Name), info.Mode()); err != nil {
				return err
			}
		} else {
			if err = copyFile(tarReader, filepath.Join(target, hdr.Name), info.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(tarReader *tar.Reader, targetFilePath string, mode os.FileMode) error {
	targetFile, err := os.OpenFile(targetFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	defer targetFile.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(targetFile, tarReader)
	return err
}

// ExistString to find if exist string
func ExistString(strs []string, str string) bool {
	for _, tmp := range strs {
		if tmp == str {
			return true
		}
	}
	return false
}

// Set an easy Set
// TODO performance
// TODO change interface{} key to string, for json.Marshal issue.
// Not routine safe
type Set []interface{}

// NewSet create a set
func NewSet(ks ...interface{}) Set {
	s := Set{}
	for _, k := range ks {
		s.Add(k)
	}
	return s
}

// NewStringSet create a set
func NewStringSet(ks ...string) Set {
	s := Set{}
	for _, k := range ks {
		s.Add(k)
	}
	return s
}

// Add add key
func (s *Set) Add(nv interface{}) {
	if !s.Exist(nv) {
		*s = append(*s, nv)
	}
}

// Exist if nexist key
func (s *Set) Exist(nv interface{}) bool {
	for _, v := range *s {
		if v == nv {
			return true
		}
	}
	return false
}

// StringList get list as strings
// TODO performance
func (s *Set) StringList() []string {
	l := []string{}
	for _, v := range *s {
		l = append(l, fmt.Sprintf("%v", v))
	}
	return l
}
