# AWS ECR Repository Creation Plugin

The goal of the plugin is to assure the existence of required AWS Elastic Container Registry repositories, when used with OCM CLI.

## Usage

1. Install the latest version of the plugin with OCM CLI:

    ```console
    ocm install plugin ghcr.io/open-component-model/ocm//ocm.software/plugins/ecrplugin
    ```

2. Configure an `ocmconfig`with credentials:

    ```yaml
    configurations:
      - type: credentials.config.ocm.software
        consumers:
          - identity:
              type: AWS
            credentials:
              - type: Credentials
                properties:
                  awsAccessKeyID: "xxx"
                  awsSecretAccessKey: "yyy"
    ```

3. Do `ocm transfer`. The plugin will be called automatically and make sure the target repository exists,
if the URL of the target repository fits to the pattern defined [here](https://github.com/open-component-model/ocm/blob/main/cmds/ecrplugin/actions/action.go#L110)
