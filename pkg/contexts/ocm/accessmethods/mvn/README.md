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

  Base URL of the Maven (mvn) repository.

- **`package`** *string*

  The name of the Maven (mvn) package.

- **`version`** *string*

  The version of the Maven (mvn) package.
