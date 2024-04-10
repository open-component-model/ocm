# `mvn` - Java packages (jar) in a Maven (mvn) repository (e.g. mvnrepository.com)

### Synopsis
```
type: mvn/v1
```

Provided blobs use the following media type: `application/x-jar`

### Description

This method implements the access of a Java package from a Maven (mvn) repository.

### Specification Versions

Supported specification version is `v1`

#### Version `v1`

The type specific specification fields are:

- **`repository`** *string*

  Base URL of the Maven (mvn) repository

- **`groupId`** *string*

  The groupId of the Maven (mvn) artifact
- 
- **`artifactId`** *string*

  The artifactId of the Maven (mvn) artifact

- **`version`** *string*

  The version name of the Maven (mvn) artifact
