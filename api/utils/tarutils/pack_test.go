package tarutils_test

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/osfs"

	"ocm.software/ocm/api/utils/tarutils"
)

var _ = Describe("tar utils mapping", func() {
	It("creates config data", func() {
		file := Must(os.CreateTemp("", "tar*"))
		defer func() {
			file.Close()
			os.Remove(file.Name())
		}()

		MustBeSuccessful(tarutils.PackFsIntoTar(osfs.OsFs, "testdata", file, tarutils.TarFileSystemOptions{}))
		file.Close()

		list := Must(tarutils.ListArchiveContent(file.Name()))
		Expect(list).To(HaveExactElements("dir", "dir/dirlink", "dir/link", "dir/regular", "dir/subdir", "dir/subdir/file", "dir2", "dir2/file2", "file"))
	})

	It("creates config data", func() {
		file := Must(os.CreateTemp("", "tar*"))
		defer func() {
			file.Close()
			os.Remove(file.Name())
		}()

		MustBeSuccessful(tarutils.PackFsIntoTar(osfs.OsFs, "testdata", file, tarutils.TarFileSystemOptions{FollowSymlinks: true}))
		file.Close()

		list := Must(tarutils.ListArchiveContent(file.Name()))
		Expect(list).To(ConsistOf("dir", "dir/dirlink", "dir/link", "dir/regular", "dir/subdir", "dir/subdir/file", "dir2", "dir2/file2", "file", "dir/dirlink/file2"))
	})

	It("test ListSortedFilesInDir with non existing path", func() {
		files, err := tarutils.ListSortedFilesInDir(osfs.New(), "/path/doesn't/exist!", true)
		Expect(err).To(HaveOccurred())
		Expect(files).To(BeNil())
		Expect(errors.Is(err, fs.ErrNotExist)).To(BeTrue())
		if runtime.GOOS == "windows" {
			Expect(err.Error()).To(ContainSubstring("The system cannot find the path specified."))
		} else {
			Expect(err.Error()).To(ContainSubstring("no such file or directory"))
		}
	})

	It("test byte-equivalent compressed archives", func() {
		fs := memoryfs.New()
		f1, err := fs.Create("some file")
		Expect(err).ToNot(HaveOccurred())
		_, err = f1.Write([]byte("some content"))
		Expect(err).ToNot(HaveOccurred())
		Expect(f1.Close()).ToNot(HaveOccurred())

		var buf1, buf2 bytes.Buffer

		Expect(tarutils.TgzFs(fs, &buf1, tarutils.TarFileSystemOptions{
			ZeroModTime: true,
		})).To(Succeed())

		Expect(tarutils.TgzFs(fs, &buf2, tarutils.TarFileSystemOptions{
			ZeroModTime: true,
		})).To(Succeed())

		Expect(buf1.Bytes()).To(Equal(buf2.Bytes()))

		var buf3 bytes.Buffer
		Expect(tarutils.TgzFs(fs, &buf3, tarutils.TarFileSystemOptions{})).To(Succeed())
		Expect(buf1.Bytes()).ToNot(Equal(buf3.Bytes()))
	})

	It("normalizes host-specific tar headers", func() {
		if runtime.GOOS == "windows" {
			Skip("file permissions and symlinks differ on windows")
		}
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "regular"), []byte("regular"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "executable"), []byte("executable"), 0o700)).To(Succeed())
		Expect(os.Mkdir(filepath.Join(dir, "subdir"), 0o700)).To(Succeed())
		Expect(os.Symlink("regular", filepath.Join(dir, "link"))).To(Succeed())

		var buf bytes.Buffer
		Expect(tarutils.PackFsIntoTar(osfs.New(), dir, &buf, tarutils.TarFileSystemOptions{
			ZeroModTime:      true,
			NormalizeHeaders: true,
		})).To(Succeed())

		modes := map[string]int64{}
		tr := tar.NewReader(&buf)
		for {
			h, err := tr.Next()
			if err == io.EOF {
				break
			}
			Expect(err).ToNot(HaveOccurred())
			Expect(h.Uid).To(BeZero(), h.Name)
			Expect(h.Gid).To(BeZero(), h.Name)
			Expect(h.Uname).To(BeEmpty(), h.Name)
			Expect(h.Gname).To(BeEmpty(), h.Name)
			modes[h.Name] = h.Mode
		}
		Expect(modes).To(Equal(map[string]int64{
			"regular":    0o644,
			"executable": 0o755,
			"subdir":     0o755,
			"link":       0o777,
		}))
	})
})
