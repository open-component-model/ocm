# Creating OCM Content and Credential Handling

Typically, reading content from public repositories does not require any
credentials. But at least adding component versions to an OCM repository
is not possible without them.

## Direct Credential Provisioning

The most simple way to access repositories with credentials is to specify
the credentials directly for the lookup call:

```go
  cfg, err := helper.ReadConfig(CFG)
  if err != nil {
	  return err
  }
  
  repoSpec := ocireg.NewRepositorySpec("ghcr.io/mandelsoft/ocm", nil)

  repo, err := octx.RepositoryForSpec(repoSpec, cfg.GetCredentials())
  if err != nil {
	  return err
  }
  defer repo.Close()
```

Credentials are given by an object of type [`credentials.Credentials`](../../pkg/contexts/credentials/interface.go).
This is basically a set of string attributes. For OCM repositories based on OCI
registries two attributes are used:
- `credentials.ATTR_USERNAME` the username
- `credentials.ATTR_PASSWORD` the password

After there is a repository object with appropriate write permissions, it is
possible to add content. As before, first, a new version
access object is requested for a component in this repository.

```go
  comp, err := repo.LookupComponent(cfg.Component)
  if err != nil {
	  return errors.Wrapf(err, "cannot lookup component %s", cfg.Component)
  }
  defer comp.Close()

  compvers, err := comp.NewVersion(cfg.Version, true)
  if err != nil {
	  return errors.Wrapf(err, "cannot create new version %s", cfg.Version)
  }
  defer compvers.Close()
```

This object now describes the access to a new component version. It is
not yet pushed to the repository.
Using various interface methods, it is possible to configure
the content for this new version. The example below just sets the provider
information and adds a
single [resource artifact](../../docs/ocm/model.md#resources) consisting of
some text.

```go
  compvers.GetDescriptor().Provider = metav1.Provider{Name: "mandelsoft"}
  
  err=compvers.SetResourceBlob(
	  &compdesc.ResourceMeta{
		  ElementMeta: compdesc.ElementMeta{
			  Name: "test",
		  }, 
		  Type:     resourcetypes.BLOB, 
		  Relation: metav1.LocalRelation,
	  }, 
	  accessio.BlobAccessForString(mime.MIME_TEXT, "testdata"), 
	  "", nil,
  )
  if err != nil {
	  return errors.Wrapf(err, "cannot add resource")
  }
```

After the component version is prepared, it can finally be added to
the repository.

```go
  if err=comp.AddVersion(compvers); err != nil {
	  return errors.Wrapf(err, "cannot add new version")
  }
```

The complete example can be found [here](cred1/example.go).

## Indirect Credential Provisioning

If only a single repository is used during the access of a component version,
the direct provisioning of credentials might be sufficient.

But even a read access might require credentials. A main task of a component 
version is to describe resource artifacts and provide access to the content of
those artifacts. They may be located not only in the
actually used component repository but in any other kind of repository, which is 
supported by an access method.

So, accessing content of artifacts, as described in the main example, might
require access to any repository described by an access specification 
used in a component version. 

Therefore, it might be required to access credentials deep in the functionality
of the library, depending on data found during the processing.
These credentials cannot be passed to the initial repository lookup. To solve
this problem a second type of context is used. The OCM context refers to
a *credentials context*.

A credentials context can be used to store credentials for dedicated purposes and
to describe access to credential stores (also called repositories in the context
of this library).

If some code requires access to dedicated credentials, it specifies the
required kind of credentials by a consumer id object. Such an object 
encapsulates the intended usage context.

A consumer id is basically a set of usage context specific string attributes. All
consumer ids always feature the attribute `type`, describing the kind of
context.

For example, to describe the request for credentials for
an [OCI registry](../../pkg/contexts/oci/identity/identity.go) and repository,
the type value is `oci.CONSUMER_TYPE`. Additionally, the following
attributes are used to fully describe the usage context.

- `ID_SCHEME` the URL scheme used to access the repository
- `ID_HOSTNAME` the hostname used to access the repository
- `ID_PORT` the port number (as string) used to access the repository
- `ID_PATHPREFIX` the namespace prefix used to access the repository
  (for OCI this is an OCI repository path)

The credentials context now allows specifying credentials
for subsets of such identity specifications. When requesting 
credentials for a repository those specifications are
checked with a type specific identity matcher to find credentials best
matching the desired usage context. For example, for the OCI context, the
used matcher tries to match the longest possible path prefix.
If the credentials setting omits an identity attribute, the setting is valid
for all possible values.

In the example, the credentials for the target repository can be specified
as follows:

```go
  octx := ocm.DefaultContext()
  
  octx.CredentialsContext().SetCredentialsForConsumer(
	  credentials.ConsumerIdentity{
		  identity.ID_TYPE: identity.CONSUMER_TYPE, 
		  identity.ID_HOSTNAME: "ghcr.io", 
		  identity.ID_PATHPREFIX: "mandelsoft",
	  }, 
	  cfg.GetCredentials(), 
  )
```

The given pattern does not specify all the possible attributes, therefore, it 
matches, for example, for all ports.

After the credentials have been configured for the used credentials context,
the repository lookup does not need to specify explicit credentials anymore,
as before.

```go
  repoSpec := ocireg.NewRepositorySpec("ghcr.io/mandelsoft/ocm", nil)

  repo, err := octx.RepositoryForSpec(repoSpec)
  if err != nil {
	  return err
  }
  defer repo.Close()
```

This mechanism explicitly works for the implicit credential requests when 
requesting resource content stored in foreign repositories requiring
authentication as well as for the initial repository lookups.

The complete example can be found [here](cred2/example.go).

## Using Credential Repositories

Instead of configuring credentials at the credentials context,
it is possible to access standard credential stores, also.

Similar to the OCM context, the credentials context is capable to manage 
the access to arbitrary credential stores, as ong as their type is supported by
the actual program context. Again, there are types for the
repository/store specification and instantiation methods to map a specification
to a repository object, which then allows accessing the named credentials found
in this repository. The supported types can be dynamically registed by the
program context.

Those repositories might require credentials, again. This is handled the same
way the credentials for the OCM/OCI repositories are handled. Some initial
credentials required to access a store must be configured for the credentials
context prior to accessing the desired credential store, which then requests
credentials via its consumer identity.

Depending on the type of the credential store, the mapping of their content
to consumer ids is done automatically.

An example for such a behaviour is the support for the docker config
file.

```go
  spec:=dockerconfig.NewRepositorySpec("~/.docker/config.json", true)

  _, err = octx.CredentialsContext().RepositoryForSpec(spec)
  if err != nil {
	  return errors.Wrapf(err, "cannot access default docker config")
  }
```

The supported repository types can be found in sub packages of the
`credentials/repositories` package.

The specification object allows configuring automatic consumer propagation.
The docker credentials are always intended for OCI repositories, therefore
it is possible to generate the consumer ids for the described credentials
based on the data contained in the configuration file. Creating the
repository for the credentials context additionally configures all the possible
consumer ids.

For more general stores, which just store credentials without a formal
type specific context, the mapping must be explicitly done as part of the 
coding. This can be done by listing the found credentials by using
a method on the credential repository object. The example above just ignores
this result value, because it uses the auto-propagation mode.

The complete example can be found [here](cred3/example.go).
