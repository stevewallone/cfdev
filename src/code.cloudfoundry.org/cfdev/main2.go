package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"code.cloudfoundry.org/cfdev/bosh"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
	"github.com/hooklift/iso9660"
)

func main() {
	garden := client.New(connection.New("tcp", "localhost:8888"))
	b, err := bosh.New(garden)
	if err != nil {
		panic(err)
	}
	say := func(message string, args ...interface{}) {
		fmt.Printf(message+"\n", args...)
	}
	if err := uploadReleases(b, say); err != nil {
		panic(err)
	}
}

type fileUpload struct {
	r io.Reader
	f os.FileInfo
}

func (f *fileUpload) Read(p []byte) (n int, err error) { return f.r.Read(p) }
func (f *fileUpload) Close() error                     { return nil }
func (f *fileUpload) Stat() (os.FileInfo, error)       { return f.f, nil }

func uploadReleases(b *bosh.Bosh, say func(message string, args ...interface{})) error {
	file, err := os.Open("/Users/dgodd/.cfdev/cache/cf-deps.iso")
	if err != nil {
		return err
	}
	r, err := iso9660.NewReader(file)
	if err != nil {
		return err
	}
	releases := make([]*fileUpload, 0, 0)
	for {
		f, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if strings.HasPrefix(f.Name(), "/releases/") && strings.HasSuffix(f.Name(), ".tgz") {
			releases = append(releases, &fileUpload{r: f.Sys().(io.Reader), f: f})
		}
	}
	for idx, fileUpload := range releases {
		say("Upload Release: %d of %d : %s", idx+1, len(releases), fileUpload.f.Name())
		if err := b.UploadRelease(fileUpload); err != nil {
			return err
		}
	}
	file.Close()
	return nil
}
