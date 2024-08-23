# Gardener Config Credential Repository

The gardener config credential repository implements the retrieval of secrets in a data format specified by the [gardener concourse utils](https://github.com/gardener/cc-utils/tree/master/model).
It supports either handing in the data via a local json file or retrieve it from a secret server ([see definition](https://github.com/gardener/cc-utils/blob/master/ccc/secrets_server.py#L29)).
