# Signing Component Versions

This tour illustrates the basic functionality to
sign and verify signatures.

It covers two basic scenarios:
- [`sign`](01-basic-signing.go) Create, Sign, Transport and Verify a component version.
- [`context`](02-using-context-settings.go) Using context settings to configure signing and verification in target repo.

## Running the examples

You can just call the main program with some config file option (`--config <file>`) and the name of the scenario.
The config file should have the following content:

```yaml
targetRepository:
  type: CommonTransportFormat
  filePath: /tmp/example06.target.ctf
  fileFormat: directory
  accessMode: 2
ocmConfig: <your ocm config file>
```

The actual version of the example just works with the filesystem
target, because it is not possible to specify credentials for the
target repository in this simple config file. But, if you specific an [OCM config file](../04-working-with-config/README.md) you can
add more credential settings to make target repositories possible
requiring credentials.

## Walkthrough

### Create, Sign, Transport and Verify a component version

As usual, we start with getting access to an OCM context

```go
	ctx := ocm.DefaultContext()
```
Then, we configure this context with optional ocm config defined in our config file.
See [OCM config scenario in tour 04](../04-working-with-config/README.md#standard-configuration-file).

```go
	err := ReadConfiguration(ctx, cfg)
	if err != nil {
		return err
	}
```

To sign a component version we need a private key.
For this example, we just create a local keypair.
To be able to verify later, we should save the public key,
but here we do all this in a single program.

```go
	privkey, pubkey, err := rsa.CreateKeyPair()
	if err != nil {
		return errors.Wrapf(err, "cannot create keypair")
	}
```

<a id="tour06-compose"></a>
And we need a component version to sign.
We compose a component version without a repository, again
(see [tour02 example 2](../02-composing-a-component-version/README.md#composition-environment)).

```go
	cv := composition.NewComponentVersion(ctx, "acme.org/example6", "v0.1.0")

	// just use the same component version setup again
	err = setupVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "version composition")
	}

	fmt.Printf("*** composition version ***\n")
	err = describeVersion(cv)
```

Now, let's sign the component version.
There might be multiple signatures, therefore every signature
has a name (here `acme.org`). Keys are always specified for
a dedicated signature name. The signing process can be influenced by
several options. Here, we just provide the private key to be used in an ad-hoc manner.
[Later](#using-context-settings-to-configure-signing), we will see how everything can be preconfigured in a *signing context*.

```go
	_, err = signing.SignComponentVersion(cv, "acme.org", signing.PrivateKey("acme.org", privkey))
	if err != nil {
		return errors.Wrapf(err, "cannot sign component version")
	}
	fmt.Printf("*** signed composition version ***\n")
	err = describeVersion(cv)
```

Now, we add the signed component version to a target repository.
Here, we just reuse the code from [tour02](../02-composing-a-component-version/README.md#composition-environment)

```go
	fmt.Printf("target repository is %s\n", string(cfg.Target))
	target, err := ctx.RepositoryForConfig(cfg.Target, nil)
	if err != nil {
		return errors.Wrapf(err, "cannot open repository")
	}
	defer target.Close()

	err = target.AddComponentVersion(cv, true)
	if err != nil {
		return errors.Wrapf(err, "cannot store signed version")
	}
```

Let's check the target for the new component version.

```go
	tcv, err := target.LookupComponentVersion("acme.org/example6", "v0.1.0")
	if err != nil {
		return errors.Wrapf(err, "transported version not found")
	}
	defer tcv.Close()

	// please be aware that the signature should be stored.
	fmt.Printf("*** target version in transportation target\n")
	err = describeVersion(tcv)
	if err != nil {
		return errors.Wrapf(err, "describe failed")
	}
```

Please note, that the version now contains a signature.

Finally, we check whether the signature is still valid for the
target version.

```go
	_, err = signing.VerifyComponentVersion(cv, "acme.org", signing.PublicKey("acme.org", pubkey))
	if err != nil {
		return errors.Wrapf(err, "verification failed")
	} else {
		fmt.Printf("verification succeeded\n")
	}
```

### Using Context Settings to Configure Signing

Instead of providing all signing relevant information directly with
the signing or verification calls, it is possible to preconfigure
various information at the OCM context.

As usual, we start with getting access to an OCM context

```go
	ctx := ocm.DefaultContext()
```

Then, we configure this context with optional ocm config defined in our config file.
See [OCM config scenario in tour 04](../04-working-with-config/README.md#standard-configuration-file).

```go
	err := ReadConfiguration(ctx, cfg)
	if err != nil {
		return err
	}
```

To sign a component version we need a private key.
For this example, we again just create a local keypair.
To be able to verify later, we should save the public key,
but here we do all this in a single program.

```go
	privkey, pubkey, err := rsa.CreateKeyPair()
	if err != nil {
		return errors.Wrapf(err, "cannot create keypair")
	}
```

Finally, we create a component version in our target repository. The called
function

```go
	err = prepareComponentInRepo(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot prepare component version in target repo")
	}
```

executes the same coding already shown in the [previous](#tour06-compose) example.

#### Signing Using Manual Context Settings

After this preparation we now configure the signing part of the OCM context.
Every OCM context features a signing registry, which provides available
signers and hashers, but also keys and certificates for various purposes.
It is always asked if a key is required, which is
not explicitly given to a signing/verification call.

This context part is implemented as additional attribute stored along
with the context. Attributes are always implemented as a separate package
containing the attribute structure, its deserialization and
a `Get(Context)` function to retrieve the attribute for the context.
This way new arbitrary attributes for various use cases can be added
without the need to change the context interface.

```go
	siginfo := signingattr.Get(ctx)
```

Now, we manually add the keys to our context.

```go
	siginfo.RegisterPrivateKey("acme.org", privkey)
	siginfo.RegisterPublicKey("acme.org", pubkey)
```

We are prepared now and can sign any component version without specifying further options
in any repository for the signature name `acme.org`. 

Therefore, we just get the component version from the prepared repository

```go
	fmt.Printf("repository is %s\n", string(cfg.Target))
	repo, err := ctx.RepositoryForConfig(cfg.Target, nil)
	if err != nil {
		return errors.Wrapf(err, "cannot open repository")
	}
	defer repo.Close()

	cv, err := repo.LookupComponentVersion("acme.org/example6", "v0.1.0")
	if err != nil {
		return errors.Wrapf(err, "version not found")
	}
	defer cv.Close()
```

and finally sign it. We don't need to present the key, here. It is taken from the
context.

```go
	_, err = signing.SignComponentVersion(cv, "acme.org")
	if err != nil {
		return errors.Wrapf(err, "cannot sign component version")
	}
```

The same way we can just call `VerifyComponentVersion` to
verify the signature.

```go
	_, err = signing.VerifyComponentVersion(cv, "acme.org")
	if err != nil {
		return errors.Wrapf(err, "verification failed")
	} else {
		fmt.Printf("verification succeeded\n")
	}
```

#### Configuring Keys with OCM Configuration File

Manually adding keys to the signing attribute
might simplify the call to possibly multiple signing/verification
calls, but it does not help to provide keys via an external
configuration (for example for using the OCM CLI).
In [tour04](../04-working-with-config/README.md#providing-new-config-object-types) we have seen how arbitrary configuration
possibilities can be added. The signing attribute uses
this mechanism to configure itself by providing an own
configuration object, which can be used to feed keys (and certificates)
into the signing attribute of an OCM context.

```go
	sigcfg := signingattr.New()
```

It provides methods to add elements
like keys and certificates, which convert
these elements into a (de-)serializable form.

```go
	sigcfg.AddPrivateKey("acme.org", privkey)
	sigcfg.AddPublicKey("acme.org", pubkey)

	ocmcfg := configcfg.New()
	ocmcfg.AddConfig(sigcfg)
```

By adding this config to a generic configuration object you get
an OCM config usable to predefine keys for your CLI.

```go
	data, err := runtime.DefaultYAMLEncoding.Marshal(ocmcfg)
	if err != nil {
		return err
	}
	fmt.Printf("ocm config file configuring standard keys:\n--- begin ocmconfig ---\n%s--- end ocmconfig ---\n", string(data))
```

And here is a sample output containing the public and private key.

```yaml
configurations:
- privateKeys:
    acme.org:
      stringdata: |
        -----BEGIN RSA PRIVATE KEY-----
        MIIEowIBAAKCAQEA4B4hgrlrvCYNqwGg+p16qCMmRwA3WhC2AK41DyQzw+VF7Zje
        jX076GZG4FWkezZBj8fLDdtHQsjdqTB0vsvN1jH2xkc3ZKHgCAAj6OG+1upOVZrZ
        rhYCmSayFx00jW8Eozw/v93QvA+tb0S5RhHkoPmD4BP6FYIpFHHcBrtvy+VPdbyn
        qfr4w7wjajmaxkxYN6TJuHCerR6F48Xp1stjW4pOH7fICq6oay07Jxekcs31JEDy
        sh163KRWFg4Uk3zx1lKPJWw69OUHztEAwk1NXL4D4e4o7pOVnGal9F4jGUAGgYV0
        KFgWuSf0h9X7aeYqtdtdh+uL7C4aRjfflrweuwIDAQABAoIBACqc8Ag4E0kB/0VN
        mPst6D2B+Ww0mVGxrblxZjtLyd/sfyBPGbnTXwmwMLfE8PJQfaTF+1DWKbWEFclu
        ojQI8klQ1LgcoGas9LjwteM40R1yDZTvTYZxPus51VDZx71Ap6QV95UWqvKnFHX8
        njG5gzwsVSvNAJcIWaE+iPRqvTYKN5prGQlyPX3AGzc5VpMYbVUeDuJGv6RVQm9n
        CYdsNBMfOWlIlUWWkZYfS2blW1kgANDlnfEE8QSQzVyzqnixFuqyXfKIArcTy1zf
        glv8PViqXWBT6pYnRIMT3UA9wtEJ39XL5N/08SBHSArBI9Cl8H3kdexMMHVfqTbx
        2stOnrkCgYEA4wH9ECpebeaR7sf2IW7tnNcX19FVOU6uJZdnfx8IZ+eO0YGLzf3H
        GKDQrn/gme2GMJhdA8EwzOgotSKwi6ZQQ/NBhRT+3b1bXthiVAzw7IKlYvDGcW8j
        SpyJrukL2WuJdkCeNOu1yFC9hu6OY1vuhsIDX3ZG5V48pv5Trf2fItUCgYEA/L2m
        06k1JTD+OKybPhVM2YuVGdVn0lqJoS77JjGdO4X+T5hmJW+fSLGWeRdexLOup2A5
        qs3GptDXtXa+imCz+F0c+5Up2eWzO3LpYTgh3KcSEwJYM97Ore7IhFLllnn9TKZQ
        hY/28j4WJk9btymlMbCQEkTLsHx6bSAtn6uIY08CgYAGVPPeE5B2uEuxwVoYLKPs
        EAPWPTyHUK6C1epJHXB6lXbiWR6xLDb1dObdNyvonzty940Aoc7eqEsaYlFSU91B
        R/O35pIFVVbLGnYDqii+MBa038ppN5RgbGVav51Y/yriZYMELt7vK3Fd5iGKm/HX
        SGoXq5DmrO48KjPjUoOx0QKBgQD6FmvKi2eEKquTxvBCiW8m4KCkFHAMTPxc2xE3
        JObxrFAN0L5aks4pb1h27/Idb7MF+gh2A/JqxMJ91EcMxE2xF+oC5AGqlEk7LSTT
        x2hyX3taBfrjMLiQkXQmw6Rcts83FmcdEmyIHXlrZSFa58GHnq+g9CQdub6m1q1u
        jpyBrwKBgGDXFUh3B4qZAZZhXSvXEQt5lf+UX16NTWUQLcYDTNqoQdN8kCNeci1N
        oCBQ5/l6EZxyQxGmlmobuqXqDoFygL1yVT1NSvlBoWO871CbAkQLbdlHKa/3Grt0
        f1SJyax3Tu9ulOZMBuV6mI58eP8ldPF5YE2v9okiKpBe/2KtSjVG
        -----END RSA PRIVATE KEY-----
  publicKeys:
    acme.org:
      stringdata: |
        -----BEGIN RSA PUBLIC KEY-----
        MIIBCgKCAQEA4B4hgrlrvCYNqwGg+p16qCMmRwA3WhC2AK41DyQzw+VF7ZjejX07
        6GZG4FWkezZBj8fLDdtHQsjdqTB0vsvN1jH2xkc3ZKHgCAAj6OG+1upOVZrZrhYC
        mSayFx00jW8Eozw/v93QvA+tb0S5RhHkoPmD4BP6FYIpFHHcBrtvy+VPdbynqfr4
        w7wjajmaxkxYN6TJuHCerR6F48Xp1stjW4pOH7fICq6oay07Jxekcs31JEDysh16
        3KRWFg4Uk3zx1lKPJWw69OUHztEAwk1NXL4D4e4o7pOVnGal9F4jGUAGgYV0KFgW
        uSf0h9X7aeYqtdtdh+uL7C4aRjfflrweuwIDAQAB
        -----END RSA PUBLIC KEY-----
  type: keys.config.ocm.software
type: generic.config.ocm.software
```