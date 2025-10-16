package accessobj_test

import (
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/tmpcache"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/mime"
)

type Source struct {
	blobaccess.DataSource
	count int
}

func NewData(data string) *Source {
	return &Source{blobaccess.DataAccessForString(data), 0}
}

func (s *Source) Reader() (io.ReadCloser, error) {
	s.count++
	return s.DataSource.Reader()
}

func (s *Source) Count() int {
	return s.count
}

type WriteAtSource struct {
	accessio.DataWriter
	data  []byte
	count int
}

func NewWriteAtSource(data string) *WriteAtSource {
	w := &WriteAtSource{data: []byte(data), count: 0}
	w.DataWriter = accessio.NewWriteAtWriter(w.write)
	return w
}

func (s *WriteAtSource) write(writer io.WriterAt) error {
	s.count++
	_, err := writer.WriteAt(s.data, 0)
	return err
}

func (s *WriteAtSource) Count() int {
	return s.count
}

var _ = Describe("cached blob", func() {
	var fs vfs.FileSystem
	var ctx datacontext.Context

	BeforeEach(func() {
		var err error
		fs, err = osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		ctx = datacontext.New(nil)
		vfsattr.Set(ctx, fs)
		tmpcache.Set(ctx, &tmpcache.Attribute{Path: "/tmp", Filesystem: fs})
	})

	AfterEach(func() {
		vfs.Cleanup(fs)
	})

	It("reads data source only once", func() {
		src := NewData("testdata")
		blob := accessobj.CachedBlobAccessForDataAccess(ctx, mime.MIME_TEXT, src)

		Expect(blob.Size()).To(Equal(int64(8)))
		Expect(blob.Digest().String()).To(Equal("sha256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"))
		Expect(blob.Get()).To(Equal([]byte("testdata")))
		Expect(blob.Get()).To(Equal([]byte("testdata")))
		Expect(src.Count()).To(Equal(1))
		dir, err := vfs.ReadDir(fs, "/tmp")
		Expect(err).To(Succeed())
		Expect(len(dir)).To(Equal(1))
		Expect(blob.Close()).To(Succeed())

		dir, err = vfs.ReadDir(fs, "/tmp")
		Expect(err).To(Succeed())
		Expect(len(dir)).To(Equal(0))
	})

	It("reads write at source only once", func() {
		src := NewWriteAtSource("testdata")
		blob := accessobj.CachedBlobAccessForWriter(ctx, mime.MIME_TEXT, src)

		Expect(blob.Size()).To(Equal(int64(8)))
		Expect(blob.Digest().String()).To(Equal("sha256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"))
		Expect(blob.Get()).To(Equal([]byte("testdata")))
		Expect(blob.Get()).To(Equal([]byte("testdata")))
		Expect(src.Count()).To(Equal(1))
		dir, err := vfs.ReadDir(fs, "/tmp")
		Expect(err).To(Succeed())
		Expect(len(dir)).To(Equal(1))
		Expect(blob.Close()).To(Succeed())

		dir, err = vfs.ReadDir(fs, "/tmp")
		Expect(err).To(Succeed())
		Expect(len(dir)).To(Equal(0))
	})
})
