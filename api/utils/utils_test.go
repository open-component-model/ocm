package utils_test

import (
	"archive/tar"
	"bytes"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils"
)

var _ = Describe("utils", func() {
	Context("WriteFileToTARArchive", func() {
		It("should write file", func() {
			fname := "testfile"
			content := []byte("testcontent")

			archiveBuf := bytes.NewBuffer([]byte{})
			tw := tar.NewWriter(archiveBuf)

			Expect(utils.WriteFileToTARArchive(fname, bytes.NewReader(content), tw)).To(Succeed())
			Expect(tw.Close()).To(Succeed())

			tr := tar.NewReader(archiveBuf)
			fheader, err := tr.Next()
			Expect(err).ToNot(HaveOccurred())
			Expect(fheader.Name).To(Equal(fname))

			actualContentBuf := bytes.NewBuffer([]byte{})
			_, err = io.Copy(actualContentBuf, tr)
			Expect(err).ToNot(HaveOccurred())
			Expect(actualContentBuf.Bytes()).To(Equal(content))

			_, err = tr.Next()
			Expect(err).To(Equal(io.EOF))
		})

		It("should write empty file", func() {
			fname := "testfile"

			archiveBuf := bytes.NewBuffer([]byte{})
			tw := tar.NewWriter(archiveBuf)

			Expect(utils.WriteFileToTARArchive(fname, bytes.NewReader([]byte{}), tw)).To(Succeed())
			Expect(tw.Close()).To(Succeed())

			tr := tar.NewReader(archiveBuf)
			fheader, err := tr.Next()
			Expect(err).ToNot(HaveOccurred())
			Expect(fheader.Name).To(Equal(fname))

			actualContentBuf := bytes.NewBuffer([]byte{})
			contentLenght, err := io.Copy(actualContentBuf, tr)
			Expect(err).ToNot(HaveOccurred())
			Expect(contentLenght).To(Equal(int64(0)))

			_, err = tr.Next()
			Expect(err).To(Equal(io.EOF))
		})

		It("should return error if filename is empty", func() {
			tw := tar.NewWriter(bytes.NewBuffer([]byte{}))
			contentReader := bytes.NewReader([]byte{})
			Expect(utils.WriteFileToTARArchive("", contentReader, tw)).To(MatchError("filename must not be empty"))
		})

		It("should return error if contentReader is nil", func() {
			tw := tar.NewWriter(bytes.NewBuffer([]byte{}))
			Expect(utils.WriteFileToTARArchive("testfile", nil, tw)).To(MatchError("contentReader must not be nil"))
		})

		It("should return error if outArchive is nil", func() {
			contentReader := bytes.NewReader([]byte{})
			Expect(utils.WriteFileToTARArchive("testfile", contentReader, nil)).To(MatchError("archiveWriter must not be nil"))
		})
	})
})
