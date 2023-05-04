## ocm credential-handling &mdash; Provisioning Of Credentials For Credential Consumers

### Description


Because of the dynamic nature of the OCM area there are several kinds of
credential consumers with potentially completely different kinds of credentials.
Therefore, a common uniform credential management is required, capable to serve
all those use cases.

This is achieved by establishing a credential request mechanism based on
generic consumer identities and credential property sets.
On the one hand every kind of credential consumer uses a dedicated consumer
type (string). Additionally, it defines a set of properties further describing
the target/context credentials are required for.

On the other hand credentials can be defined for such sets of identities
with partial sets of properties (see [ocm configfile](ocm_configfile.md)). A credential
request is then matched against the available credential settings using matchers,
which might be specific for dedicated kinds of requests. For example, a hostpath
matcher matches a path prefix for a <code>pathprefix</code> property.

The best matching set of credential properties is then returned to the
credential consumer, which checks for the expected credential properties.

The following credential consumer types are used:
  - <code>Buildcredentials.ocm.software</code>: Gardener config credential matcher
    
    It matches the <code>Buildcredentials.ocm.software</code> consumer type and additionally acts like
    the <code>hostpath</code> type.
    
    Credential consumers of the consumer type Buildcredentials.ocm.software evaluate the following credential properties:
    
      - <code>key</code>: secret key use to access the credential server
    

  - <code>Github</code>: GitHub credential matcher
    
    This matcher is a hostpath matcher.
    
    Credential consumers of the consumer type Github evaluate the following credential properties:
    
      - <code>token</code>: GitHub personal access token
    

  - <code>HelmChartRepository</code>: Helm chart repository
    
    It matches the <code>HelmChartRepository</code> consumer type and additionally acts like 
    the <code>hostpath</code> type.
    
    Credential consumers of the consumer type HelmChartRepository evaluate the following credential properties:
    
    - **<code>username</code>**: basic auth user name.
    - **<code>password</code>**: basic auth password.
    - **<code>certificate</code>**: TLS client certificate.
    - **<code>privateKey</code>**: TLS private key.

  - <code>OCIRegistry</code>: OCI registry credential matcher
    
    It matches the <code>OCIRegistry</code> consumer type and additionally acts like 
    the <code>hostpath</code> type.
    
    Credential consumers of the consumer type OCIRegistry evaluate the following credential properties:
    
      - <code>username</code>: the basic auth user name
      - <code>password</code>: the basic auth password
      - <code>identityToken</code>: the bearer token used for non-basic auth authorization
    

  - <code>S3</code>: S3 credential matcher
    
    This matcher is a hostpath matcher.
    
    Credential consumers of the consumer type S3 evaluate the following credential properties:
    
      - <code>awsAccessKeyID</code>: AWS access key id
      - <code>awsSecretAccessKey</code>: AWS secret for access key id
      - <code>token</code>: AWS access token (alternatively)
    

\
Those consumer types provide their own matchers, which are often based
on some standard generic matches. Those generic matchers and their
behaviours are described in the following list:
  - <code>exact</code>: exact match of given pattern set
  - <code>hostpath</code>: Host and path based credential matcher
    
    This matcher works on the following properties:
    
    - *<code>type</code>* (required if set in pattern): the identity type 
    - *<code>hostname</code>* (required if set in pattern): the hostname of a server
    - *<code>port</code>* (optional): the port of a server
    - *<code>pathprefix</code>* (optional): a path prefix to match. The 
      element with the most matching path components is selected (separator is <code>/</code>).
    

  - <code>partial</code>: complete match of given pattern ignoring additional attributes



### SEE ALSO

##### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm configfile</b>](ocm_configfile.md)	 &mdash; configuration file

