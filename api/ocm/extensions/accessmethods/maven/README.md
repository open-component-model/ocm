# `mvn` - Maven artifacts (Java packages, jars) in a Maven (mvn) repository (e.g. mvnrepository.com)

## Synopsis

```yaml
type: maven/v1
```

Provided blobs use the following media type: `application/x-tgz`

### Description

This method implements the access of a resource hosted by a maven repository or a
complete resource set denoted by a GAV (GroupId, ArtifactId, Version).

### Specification Versions

Supported specification version is `v1`

#### Version `v1`

The type specific specification fields are:

- **`repoUrl`** *string*

  Base URL of the Maven (mvn) repository

- **`groupId`** *string*

  The groupId of the Maven (mvn) artifact

- **`artifactId`** *string*

  The artifactId of the Maven (mvn) artifact

- **`version`** *string*

  The version name of the Maven (mvn) artifact

- **`classifier`** *string*

  The optional classifier of the Maven (mvn) artifact

- **`extension`** *string*

  The optional extension of the Maven (mvn) artifact

If classifier/extension is given a dedicated resource is described,
otherwise the complete resource set described by a GAV.
Only complete resource sets can be uploaded again to a Maven repository.

#### Examples

##### Complete resource set denoted by a GAV

```yaml
name: acme.org/complete/gav
version: 0.0.1
provider:
  name: acme.org
resources:
  - name: java-sap-vcap-services
    type: mvnArtifact
    version: 0.0.1
    access:
      type: mvn
      repository: https://repo1.maven.org/maven2
      groupId: com.sap.cloud.environment.servicebinding
      artifactId: java-sap-vcap-services
      version: 0.10.4
```

##### Single pom.xml file

This can't be uploaded again into a Maven repository, but it can be used to describe the dependencies of a project.
The mime type will be `application/xml`.

```yaml
name: acme.org/single/pom
version: 0.0.1
provider:
  name: acme.org
resources:
  - name: sap-cloud-sdk
    type: pom
    version: 0.0.1
    access:
      type: mvn
      repository: https://repo1.maven.org/maven2
      groupId: com.sap.cloud.sdk
      artifactId: sdk-modules-bom
      version: 5.7.0
      classifier: ''
      extension: pom
```

##### Single binary file

In case you want to download and install maven itself, you can use the following example.
This can't be uploaded again into a Maven repository.
The mime type will be `application/gzip`.

```yaml
name: acme.org/bin/zip
version: 0.0.1
provider:
  name: acme.org
resources:
  - name: maven
    type: bin
    version: 0.0.1
    access:
      type: mvn
      repository: https://repo1.maven.org/maven2
      groupId: org.apache.maven
      artifactId: apache-maven
      version: 3.9.6
      classifier: bin
      extension: zip
```
