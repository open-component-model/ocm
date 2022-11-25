## ocm ocm-accessmethods &mdash; List Of All Supported Access Methods

### Description


Access methods are used to handle the access to the content of artifacts
described in a component version. Therefore, an artifact entry contains
an access specification describing the access attributes for the dedicated
artifact.


The following list describes the supported access methods, their versions
and specification formats.
Typically there is special support for the CLI artifact add commands.
The access method specification can be put below the <code>access</code> field.
If always requires the field <code>type</code> describing the kind and version
shown below.

- Access type <code>S3</code>

  This method implements the access of a blob stored in an S3 bucket.

  The following versions are supported:
  - Version <code>v1</code>
  
    The type specific specification fields are:
    
    - **<code>region</code>** (optional) *string*
    
      OCI repository reference (this artifact name used to store the blob).
    
    - **<code>bucket</code>** *string*
    
      The name of the S3 bucket containing the blob
    
    - **<code>key</code>** *string*
    
      The key of the desired blob
    
    Options used to configure fields: <code>--accessVersion</code>, <code>--bucket</code>, <code>--mediaType</code>, <code>--reference</code>, <code>--region</code>
  

- Access type <code>gitHub</code>

  This method implements the access of the content of a git commit stored in a
  GitHub repository.

  The following versions are supported:
  - Version <code>v1</code>
  
    The type specific specification fields are:
    
    - **<code>repoUrl</code>**  *string*
    
      Repository URL with or without scheme.
    
    - **<code>ref</code>** (optional) *string*
    
      Original ref used to get the commit from
    
    - **<code>commit</code>** *string*
    
      The sha/id of the git commit
    
    Options used to configure fields: <code>--accessHostname</code>, <code>--accessRepository</code>, <code>--commit</code>
  

- Access type <code>localBlob</code>

  This method is used to store a resource blob along with the component descriptor
  on behalf of the hosting OCM repository.
  
  Its implementation is specific to the implementation of OCM
  repository used to read the component descriptor. Every repository
  implementation may decide how and where local blobs are stored,
  but it MUST provide an implementation for this method.
  
  Regardless of the chosen implementation the attribute specification is
  defined globally the same.

  The following versions are supported:
  - Version <code>v1</code>
  
    The type specific specification fields are:
    
    - **<code>localReference</code>** *string*
    
      Repository type specific location information as string. The value
      may encode any deep structure, but typically just an access path is sufficient.
    
    - **<code>mediaType</code>** *string*
    
      The media type of the blob used to store the resource. It may add 
      format information like <code>+tar</code> or <code>+gzip</code>.
    
    - **<code>referenceName</code>** (optional) *string*
    
      This optional attribute may contain identity information used by
      other repositories to restore some global access with an identity
      related to the original source.
    
      For example, if an OCI artifact originally referenced using the
      access method [<code>ociArtifact</code>](../../../../../docs/formats/accessmethods/ociArtifact.md) is stored during
      some transport step as local artifact, the reference name can be set
      to its original repository name. An import step into an OCI based OCM
      repository may then decide to make this artifact available again as 
      regular OCI artifact.
    
    - **<code>globalAccess</code>** (optional) *access method specification*
    
      If a resource blob is stored locally, the repository implementation
      may decide to provide an external access information (independent
      of the OCM model).
    
      For example, an OCI artifact stored as local blob
      can be additionally stored as regular OCI artifact in an OCI registry.
      
      This additional external access information can be added using
      a second external access method specification.
    
    Options used to configure fields: <code>--globalAccess</code>, <code>--hint</code>, <code>--mediaType</code>, <code>--reference</code>
  

- Access type <code>none</code>

  dummy resource with no access


- Access type <code>ociArtifact</code>

  This method implements the access of an OCI artifact stored in an OCI registry.

  The following versions are supported:
  - Version <code>v1</code>
  
    The type specific specification fields are:
    
    - **<code>imageReference</code>** *string*
    
      OCI image/artifact reference following the possible docker schemes:
      - <code>&lt;repo>/&lt;artifact>:&lt;digest>@&lt;tag></code>
      - <code><host>[&lt;port>]/&lt;repo path>/&lt;artifact>:&lt;version>@&lt;tag></code>
    
    Options used to configure fields: <code>--reference</code>
  

- Access type <code>ociBlob</code>

  This method implements the access of an OCI blob stored in an OCI repository.

  The following versions are supported:
  - Version <code>v1</code>
  
    The type specific specification fields are:
    
    - **<code>imageReference</code>** *string*
    
      OCI repository reference (this artifact name used to store the blob).
    
    - **<code>mediaType</code>** *string*
    
      The media type of the blob
    
    - **<code>digest</code>** *string*
    
      The digest of the blob used to access the blob in the OCI repository.
    
    - **<code>size</code>** *integer*
    
      The size of the blob
    
    Options used to configure fields: <code>--digest</code>, <code>--mediaType</code>, <code>--reference</code>, <code>--size</code>
  


### SEE ALSO

##### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client

