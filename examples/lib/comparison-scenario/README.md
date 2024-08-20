# Podinfo Provisioning and Consumption Scenarios

This example create, publishes, consumes and localizes
an OCM component version for a helm deployment for the podinfo application.

## Preparation

Copy the files under config into your environment and
set the values according to your needs:

- config.yml

  ```yaml
    username: mandelsoft
    password: ghp_xxxxxx
    repository: ghcr.io/mandelsoft/test
    targetRepository:
      type: OCIRegistry
      baseUrl: ghcr.io/mandelsoft/testtarget
  ```

  Adapt your user and password for your preferred OCI registry.
  In the `targetRepository` you can specify any OCM repository spec.
  To make the scenarios work with OCI image references it must be an OCI
  registry.

- config-OO.yml

  This config uses an OCM config file, you additionally need the `ocmconfig.yaml` file.

  ```yaml
    repository: ghcr.io/mandelsoft/test
    ocmconfig: /home/mandelsoft/comparison/ocmconfig.yaml
    #targetRepository:
    #  type: CommonTransportFormat
    #  filePath: /tmp/comparison.ctf
    #  fileFormat: directory
    #  accessMode: 2
    targetRepository:
    type: OCIRegistry
    baseUrl: ghcr.io/mandelsoft/testtarget
  ```

  The entry `ocmconfig` must point to your ocmconfig file path.

- ocmconfig.yaml

  It is configured to use your docker config for the access to
  OCI registries.
  You just need to create an RSA key pair (for example with `ocm create rsakeypair`) and add the file paths to the appropriate place in the config file.

## Execution

There is one main program usable to call all scenarios.
Just go to this folder and run the main program with
the option `--config <your config file to use>` and the scenario name.

## Scenarios

Scenarios 1-8 use implicitly created RSA keys for signing, and
demonstrate the manual usage of such keys. The other scenarios
use the pre-created keys configured in the ocm config.

1) `create` (*requires `config.yaml`*)

   Compose a component version containing three resources:
   - podinfo image as reference
   - helmchart as reference
   - deployscript taken from `resources/deployscript`

2) `sign`  (*requires `config.yaml`*)

   Sign the component version (implicitly uses 1 to compose it)

3) `write`  (*requires `config.yaml`*)

    Write the signed version to the OCM repository taken
    from config field `repository`.

4) `transport` (*requires `config.yaml`*)

   Transport the content (as value transport) to the OCM
   repository specified in config field `targetRepository`.

5) `verify`  (*requires `config.yaml`*)

    Verify the signature of the component version in the
    target repository

6) `download`  (*requires `config.yaml`*)

   Download the helm chart from the target repository and list the files.

7) `getref`  (*requires `config.yaml`*)

   Determine the OCI reference of the podinfo image in the target
   environment.

8) `deployscript`  (*requires `config.yaml`*)
   Download and print the deployscript taken from the target repository.

9) `localize`  (*requires `config-OO.yaml`*)

   Complete scenario to prepare the deployment provided by the
   component version. Print the values and list the chart files.

Then there are aggregated scenarios:

- `provider` (*requires `config-OO.yaml`*)

  It handles the complete provisioning side, from composition of
  the component version to its publishing to an OCI registry.

- `consumer` (*requires `config-OO.yaml`*)

  It handles the complete scenario to import and verify the component
  version into the target environment. Additionally,  its prints
  all the information from 6-8. It prepares the scene for the `localize` scenario.

- `deploy` (*requires `config-OO.yaml`*)

  It handles the localize scenario and uses the provided information
  to execute a helm install action.
