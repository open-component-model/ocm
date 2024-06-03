package tarutils_test

import (
	"io/fs"
	"os"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/osfs"

	"github.com/open-component-model/ocm/pkg/utils/tarutils"
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

})
