package git_test

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/cwdfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/datacontext/attrs/tmpcache"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/cpi"
	rgit "ocm.software/ocm/api/oci/extensions/repositories/git"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess"
	ocmlog "ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/refmgmt"
)

var _ = Describe("ctf management", func() {
	var remoteRepo *git.Repository

	var tmp vfs.FileSystem
	var workspace vfs.FileSystem

	var repoDir string
	var repoURL string

	ocmlog.Context().AddRule(logging.NewConditionRule(logging.TraceLevel, refmgmt.ALLOC_REALM))

	ctx := oci.New()

	BeforeEach(func() {
		path := GinkgoT().TempDir()
		tmp = Must(cwdfs.New(osfs.New(), path))
		tmpcache.Set(ctx, &tmpcache.Attribute{Path: ".", Filesystem: tmp})
		vfsattr.Set(ctx, tmp)

		Expect(tmp.Mkdir("repo", 0o700)).To(Succeed())
		repoDir = path + filepath.PathSeparatorString + "repo"
		repoURL = "file://" + repoDir

		Expect(tmp.Mkdir("workspace", 0o700)).To(Succeed())
		workspace = Must(cwdfs.New(tmp, "workspace"))
	})

	BeforeEach(func() {
		remoteRepo = Must(git.PlainInit(repoDir, true))
	})

	AfterEach(func() {
		Expect(vfs.Cleanup(tmp)).To(Succeed())
	})

	It("instantiate git based ctf", func() {
		repo := Must(rgit.Create(ctx, accessobj.ACC_CREATE, repoURL, accessio.RepresentationFileSystem(workspace)))
		ns := Must(repo.LookupNamespace("test"))

		testData := []byte("testdata")

		aa := NewArtifact(ns, testData)

		Expect(aa.Close()).To(Succeed())
		Expect(ns.Close()).To(Succeed())
		Expect(repo.Close()).To(Succeed())

		commits := Must(remoteRepo.CommitObjects())
		validAdd := 0
		validSync := 0
		var messages []string
		Expect(commits.ForEach(func(commit *object.Commit) error {
			if expected := rgit.GenerateCommitMessageForArtifact(rgit.OperationAdd, aa); commit.Message == expected {
				validAdd++
			}
			if expected := rgit.GenerateCommitMessageForArtifact(rgit.OperationSync, aa); commit.Message == expected {
				validSync++
			}
			messages = append(messages, commit.Message)
			return nil
		})).To(Succeed())

		Expect(validAdd).To(Equal(1),
			fmt.Sprintf(
				"expected exactly one commit with message %q, got %d commits with messages:\n%v",
				rgit.GenerateCommitMessageForArtifact(rgit.OperationAdd, aa),
				validAdd,
				messages,
			))
		Expect(validSync).To(Equal(1),
			fmt.Sprintf(
				"expected exactly one commit with message %q, got %d commits with messages:\n%v",
				rgit.GenerateCommitMessageForArtifact(rgit.OperationAdd, aa),
				validAdd,
				messages,
			))
	})
})

func NewArtifact(n cpi.NamespaceAccess, data []byte) cpi.ArtifactAccess {
	art := Must(n.NewArtifact())
	Expect(art.AddLayer(blobaccess.ForData(mime.MIME_OCTET, data), nil)).To(Equal(0))
	desc := Must(art.Manifest())
	Expect(desc).NotTo(BeNil())

	Expect(desc.Layers[0].Digest).To(Equal(digest.FromBytes(data)))
	Expect(desc.Layers[0].MediaType).To(Equal(mime.MIME_OCTET))
	Expect(desc.Layers[0].Size).To(Equal(int64(8)))

	config := blobaccess.ForData(mime.MIME_OCTET, []byte("{}"))
	desc.Config = *artdesc.DefaultBlobDescriptor(config)
	MustBeSuccessful(n.AddBlob(config))
	MustBeSuccessful(n.AddArtifact(desc))
	return art
}
