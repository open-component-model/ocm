# Working with Credentials

This tour illustrates the basic handling of credentials
using the OCM library. The library provides
an extensible framework to bring together credential providers
and credential cosunmers in a technology-agnostic way.

It covers four basic scenarios:
- [`basic`](01-using-credentials.go) Writing to a repository with directly specified credentials.
- [`generic`](02-basic-credential-management.go) Using credentials via the credential management.
- [`read`](02-basic-credential-management.go) Read the previously component version using the credential management.
- [`credrepo`](03-credential-repositories.go) Providing credentials via credential repositories.

You can just call the main program with some config file option (`--config <file>`) and the name of the scenario.
The config file should have the following content:

```yaml
repository: ghcr.io/mandelsoft/ocm
username:
password:
```