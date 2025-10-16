package install_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	. "ocm.software/ocm/api/helper/builder"
	. "ocm.software/ocm/api/helper/env"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/tools/toi"
	"ocm.software/ocm/api/ocm/tools/toi/drivers/mock"
	"ocm.software/ocm/api/ocm/tools/toi/install"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
)

const (
	COMPONENT = "acme.org/test"
	VERSION   = "0.1.0"
)

type Driver struct {
	install.Driver
	Found *install.Operation
}

func NewDriver() *Driver {
	driver := &Driver{}
	driver.Driver = mock.New(func(op *install.Operation) (*install.OperationResult, error) {
		driver.Found = op
		return &install.OperationResult{}, nil
	})
	return driver
}

var _ = Describe("Transfer handler", func() {
	var env *Builder
	var driver *Driver

	cid1 := credentials.NewConsumerIdentity("test", hostpath.ID_HOSTNAME, "test.de")
	creds1 := credentials.NewCredentials(common.Properties{"user": "test", "password": "pw"})

	BeforeEach(func() {
		env = NewBuilder(FileSystem(memoryfs.New(), ""))

		env.OCMCommonTransport("ctf", accessio.FormatDirectory, func() {
			env.ComponentVersion(COMPONENT, VERSION, func() {
				env.Provider("acme.org")
				env.Resource("package", VERSION, toi.TypeTOIPackage, v1.LocalRelation, func() {
					env.BlobData(mime.MIME_YAML, []byte(""))
				})
			})
		})

		driver = NewDriver()
	})

	It("gets credentials", func() {
		env.CredentialsContext().SetCredentialsForConsumer(cid1, creds1)

		c := Must(credentials.CredentialsForConsumer(env.OCMContext().CredentialsContext(), cid1))
		Expect(c.Properties()).To(Equal(creds1.Properties()))
	})

	It("executes with credential substitution", func() {
		env.CredentialsContext().SetCredentialsForConsumer(cid1, creds1)

		p, _ := common.NewBufferedPrinter()

		mapping := `
testparam: (( merge ))
creds: (( hasCredentials("mycred") ? [getCredentials("mycred")]  :[]  ))
`
		spec := &toi.PackageSpecification{
			CredentialsRequest: toi.CredentialsRequest{
				Credentials: map[string]toi.CredentialsRequestSpec{
					"mycred": {
						ConsumerId:  cid1,
						Description: "test",
						Optional:    false,
					},
				},
			},
			Executors: []toi.Executor{
				{
					Actions: []string{"install"},
					Image: &toi.Image{
						Ref: "a/b:v1",
					},
					ParameterMapping: []byte(mapping),
				},
			},
		}

		credspec := &toi.Credentials{
			Credentials: map[string]toi.CredentialSpec{
				"mycred": {
					ConsumerId: cid1,
				},
			},
		}

		params := `
testparam: value
`

		repo := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, "/ctf", 0, env))
		defer Close(repo)
		cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv)

		Must(install.ExecuteAction(p, driver, "install", spec, credspec, []byte(params), env, cv, nil))

		effparams := Must(driver.Found.Files[install.InputParameters].Get())
		Expect(string(effparams)).To(StringEqualTrimmedWithContext(`
creds:
- password: pw
  user: test
testparam: value
`))
	})

	It("executes with credential property substitution", func() {
		env.CredentialsContext().SetCredentialsForConsumer(cid1, creds1)

		p, _ := common.NewBufferedPrinter()

		mapping := `
testparam: (( merge ))
creds: (( hasCredentials("mycred") ? getCredentials("mycred", "user")  :"" ))
`
		spec := &toi.PackageSpecification{
			CredentialsRequest: toi.CredentialsRequest{
				Credentials: map[string]toi.CredentialsRequestSpec{
					"mycred": {
						ConsumerId:  cid1,
						Description: "test",
						Optional:    false,
					},
				},
			},
			Executors: []toi.Executor{
				{
					Actions: []string{"install"},
					Image: &toi.Image{
						Ref: "a/b:v1",
					},
					ParameterMapping: []byte(mapping),
				},
			},
		}

		credspec := &toi.Credentials{
			Credentials: map[string]toi.CredentialSpec{
				"mycred": {
					ConsumerId: cid1,
				},
			},
		}

		params := `
testparam: value
`

		repo := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, "/ctf", 0, env))
		defer Close(repo)
		cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv)

		Must(install.ExecuteAction(p, driver, "install", spec, credspec, []byte(params), env, cv, nil))

		effparams := Must(driver.Found.Files[install.InputParameters].Get())
		Expect(string(effparams)).To(StringEqualTrimmedWithContext(`
creds: test
testparam: value
`))
	})

	It("executes with single credential property substitution", func() {
		creds1 := credentials.NewCredentials(common.Properties{"user": "test"})

		env.CredentialsContext().SetCredentialsForConsumer(cid1, creds1)

		p, _ := common.NewBufferedPrinter()

		mapping := `
testparam: (( merge ))
creds: (( hasCredentials("mycred") ? getCredentials("mycred", "*")  :"" ))
`
		spec := &toi.PackageSpecification{
			CredentialsRequest: toi.CredentialsRequest{
				Credentials: map[string]toi.CredentialsRequestSpec{
					"mycred": {
						ConsumerId:  cid1,
						Description: "test",
						Optional:    false,
					},
				},
			},
			Executors: []toi.Executor{
				{
					Actions: []string{"install"},
					Image: &toi.Image{
						Ref: "a/b:v1",
					},
					ParameterMapping: []byte(mapping),
				},
			},
		}

		credspec := &toi.Credentials{
			Credentials: map[string]toi.CredentialSpec{
				"mycred": {
					ConsumerId: cid1,
				},
			},
		}

		params := `
testparam: value
`

		repo := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, "/ctf", 0, env))
		defer Close(repo)
		cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv)

		Must(install.ExecuteAction(p, driver, "install", spec, credspec, []byte(params), env, cv, nil))

		effparams := Must(driver.Found.Files[install.InputParameters].Get())
		Expect(string(effparams)).To(StringEqualTrimmedWithContext(`
creds: test
testparam: value
`))
	})

	It("executes with optional credential substitution without credentials", func() {
		env.CredentialsContext().SetCredentialsForConsumer(cid1, creds1)

		p, _ := common.NewBufferedPrinter()

		mapping := `
testparam: (( merge ))
creds: (( hasCredentials("mycred") ? [getCredentials("mycred")]  :[]  ))
`
		spec := &toi.PackageSpecification{
			CredentialsRequest: toi.CredentialsRequest{
				Credentials: map[string]toi.CredentialsRequestSpec{
					"mycred": {
						ConsumerId:  cid1,
						Description: "test",
						Optional:    true,
					},
				},
			},
			Executors: []toi.Executor{
				{
					Actions: []string{"install"},
					Image: &toi.Image{
						Ref: "a/b:v1",
					},
					ParameterMapping: []byte(mapping),
				},
			},
		}

		credspec := &toi.Credentials{}

		params := `
testparam: value
`

		repo := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, "/ctf", 0, env))
		defer Close(repo)
		cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv)

		Must(install.ExecuteAction(p, driver, "install", spec, credspec, []byte(params), env, cv, nil))

		effparams := Must(driver.Found.Files[install.InputParameters].Get())
		Expect(string(effparams)).To(StringEqualTrimmedWithContext(`
creds: []
testparam: value
`))
	})
})
