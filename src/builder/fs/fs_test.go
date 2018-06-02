package fs_test

import (
	"builder/fs"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("fs", func() {
	var (
		tmpDir  string
		subject *fs.Dir
	)
	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "fs-test-")
		Expect(err).To(BeNil())
		subject, err = fs.New(tmpDir)
		Expect(err).To(BeNil())
	})
	AfterEach(func() { os.RemoveAll(tmpDir) })

	Describe("AddReader", func() {
		It("puts the file on disk", func() {
			Expect(subject.AddReader("fred.txt", bytes.NewReader([]byte("some text")))).To(Succeed())
			Expect(ioutil.ReadFile(filepath.Join(tmpDir, "fred.txt"))).To(Equal([]byte("some text")))
		})
		It("creates subdirs", func() {
			Expect(subject.AddReader("jim/jane/fred.txt", bytes.NewReader([]byte("some text")))).To(Succeed())
			Expect(ioutil.ReadFile(filepath.Join(tmpDir, "jim/jane/fred.txt"))).To(Equal([]byte("some text")))
		})

		Context("file already exists", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "fred.txt"), []byte("existing text"), 0644)).To(Succeed())
			})
			It("overrides existing file", func() {
				Expect(subject.AddReader("fred.txt", bytes.NewReader([]byte("some text")))).To(Succeed())
				Expect(ioutil.ReadFile(filepath.Join(tmpDir, "fred.txt"))).To(Equal([]byte("some text")))
			})
		})
	})

	Describe("AddFile", func() {
		var tmpFile *os.File
		BeforeEach(func() {
			var err error
			tmpFile, err = ioutil.TempFile("", "fs-test-file-to-add-")
			Expect(err).To(BeNil())
			_, err = tmpFile.Write([]byte("exciting stuff"))
			Expect(err).To(BeNil())
			Expect(tmpFile.Close()).To(Succeed())
		})
		AfterEach(func() { os.Remove(tmpFile.Name()) })

		It("copies file to dir", func() {
			Expect(subject.AddFile("jim/bob.txt", tmpFile.Name())).To(Succeed())
			Expect(ioutil.ReadFile(filepath.Join(tmpDir, "jim/bob.txt"))).To(Equal([]byte("exciting stuff")))
		})

		Context("file already exists", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "bob.txt"), []byte("existing text"), 0644)).To(Succeed())
			})
			It("overrides existing file", func() {
				Expect(subject.AddFile("bob.txt", tmpFile.Name())).To(Succeed())
				Expect(ioutil.ReadFile(filepath.Join(tmpDir, "bob.txt"))).To(Equal([]byte("exciting stuff")))
			})
		})
	})

	Describe("AddURL", func() {
		var server *ghttp.Server
		BeforeEach(func() {
			server = ghttp.NewServer()
		})
		AfterEach(func() { server.Close() })

		Context("http ok", func() {
			BeforeEach(func() {
				server.AppendHandlers(ghttp.RespondWith(http.StatusOK, "text from http"))
			})
			It("copies file to dir", func() {
				Expect(subject.AddURL("jane/jill.txt", server.URL())).To(Succeed())
				Expect(ioutil.ReadFile(filepath.Join(tmpDir, "jane/jill.txt"))).To(Equal([]byte("text from http")))
			})
		})

		Context("http 404", func() {
			BeforeEach(func() {
				server.AppendHandlers(ghttp.RespondWith(http.StatusNotFound, "not found"))
			})
			It("returns error", func() {
				Expect(subject.AddURL("jane/jill.txt", server.URL())).To(MatchError(fmt.Sprintf("Non 200 status code: 404: %s", server.URL())))
				Expect(filepath.Join(tmpDir, "jane/jill.txt")).ToNot(BeAnExistingFile())
			})
		})
	})

	Describe("Exists", func() {
		Context("file exists", func() {
			It("returns true", func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "bob.txt"), []byte("existing text"), 0644)).To(Succeed())
				Expect(subject.Exists("bob.txt")).To(BeTrue())
			})
		})
		Context("file DOESNT exist", func() {
			It("returns false", func() {
				Expect(subject.Exists("bob.txt")).To(BeFalse())
			})
		})
	})

	Describe("DeleteOld", func() {
		var filename string
		BeforeEach(func() {
			filename = filepath.Join(tmpDir, "a", "b", "file.txt")
			Expect(os.MkdirAll(filepath.Dir(filename), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filename, []byte("text"), 0644)).To(Succeed())
		})
		It("deletes files which weren't added this run", func() {
			Expect(subject.DeleteOld()).To(Succeed())
			Expect(filename).ToNot(BeAnExistingFile())
		})
		It("leaves files which were updated", func() {
			Expect(subject.AddBytes("a/b/file.txt", []byte("new"))).To(Succeed())
			Expect(subject.DeleteOld()).To(Succeed())
			Expect(filename).To(BeAnExistingFile())
		})
		It("leaves new files", func() {
			Expect(subject.AddBytes("file_2.txt", []byte("new"))).To(Succeed())
			Expect(subject.DeleteOld()).To(Succeed())
			Expect(filepath.Join(tmpDir, "file_2.txt")).To(BeAnExistingFile())
		})
	})
})
