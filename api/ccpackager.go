/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"fmt"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
)

// Functions copied from packager.go of fabric-sdk-go, to support different Node chaincode package.

// Descriptor ...
type Descriptor struct {
	name string
	fqp  string
}

// NewNodeCCPackage creates new go lang chaincode package
func NewJavaCCPackage(chaincodeRealPath string) (*resource.CCPackage, error) {
	descriptors, err := findSource(chaincodeRealPath)
	if err != nil {
		return nil, err
	}
	tarBytes, err := generateTarGz(descriptors)
	if err != nil {
		return nil, err
	}
	ccPkg := &resource.CCPackage{Type: pb.ChaincodeSpec_JAVA, Code: tarBytes}
	return ccPkg, nil
}

// NewNodeCCPackage creates new go lang chaincode package
func NewNodeCCPackage(chaincodeRealPath string) (*resource.CCPackage, error) {
	descriptors, err := findSource(chaincodeRealPath)
	if err != nil {
		return nil, err
	}
	tarBytes, err := generateTarGz(descriptors)
	if err != nil {
		return nil, err
	}
	ccPkg := &resource.CCPackage{Type: pb.ChaincodeSpec_NODE, Code: tarBytes}
	return ccPkg, nil
}

func findSource(filePath string) ([]*Descriptor, error) {
	var descriptors []*Descriptor
	err := filepath.Walk(filePath,
		func(path string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fileInfo.Mode().IsRegular() && isSource(path) {
				relPath, err := filepath.Rel(filePath, path)
				if err != nil {
					return err
				}
				descriptors = append(descriptors, &Descriptor{name: relPath, fqp: path})
			}
			return nil
		})

	return descriptors, err
}

func isSource(filePath string) bool {
	return true
}

func generateTarGz(descriptors []*Descriptor) ([]byte, error) {
	var codePackage bytes.Buffer
	gw := gzip.NewWriter(&codePackage)
	tw := tar.NewWriter(gw)
	for _, v := range descriptors {
		err := packEntry(tw, gw, v)
		if err != nil {
			err1 := closeStream(tw, gw)
			if err1 != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("packEntry failed and close error %s", err1))
			}
			return nil, errors.Wrap(err, "packEntry failed")
		}
	}
	err := closeStream(tw, gw)
	if err != nil {
		return nil, errors.Wrap(err, "closeStream failed")
	}
	return codePackage.Bytes(), nil

}

func closeStream(tw io.Closer, gw io.Closer) error {
	err := tw.Close()
	if err != nil {
		return err
	}
	err = gw.Close()
	return err
}

func packEntry(tw *tar.Writer, gw *gzip.Writer, descriptor *Descriptor) error {
	file, err := os.Open(descriptor.fqp)
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Errorf("error file close %s", err)
		}
	}()

	if stat, err := file.Stat(); err == nil {
		// now lets create the header as needed for this file within the tarball
		header := new(tar.Header)
		header.Name = descriptor.name
		header.Size = stat.Size()
		header.Mode = int64(stat.Mode())
		// Use a deterministic "zero-time" for all date fields
		header.ModTime = time.Time{}
		header.AccessTime = time.Time{}
		header.ChangeTime = time.Time{}
		// write the header to the tarball archive
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if _, err := io.Copy(tw, file); err != nil {
			return err
		}
		if err := tw.Flush(); err != nil {
			return err
		}
		if err := gw.Flush(); err != nil {
			return err
		}

	}
	return nil
}
