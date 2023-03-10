## ocm controller install &mdash; Install Either A Specific Or Latest Version Of The Ocm-Controller.

### Synopsis

```
ocm controller install controller {--version v0.0.1}
```

### Options

```
  -u, --base-url string          the base url to the ocm-controller's release page (default "https://github.com/open-component-model/ocm-controller/releases")
  -c, --controller-name string   name of the controller that's used for status check (default "ocm-controller")
  -d, --dry-run                  if enabled, prints the downloaded manifest file
  -h, --help                     help for install
  -n, --namespace string         the namespace into which the controller is installed (default "ocm-system")
  -a, --release-api-url string   the base url to the ocm-controller's API release page (default "https://api.github.com/repos/open-component-model/ocm-controller/releases")
  -t, --timeout duration         maximum time to wait for deployment to be ready (default 1m0s)
  -v, --version string           the version of the controller to install (default "latest")
```

### SEE ALSO

##### Parents

* [ocm controller](ocm_controller.md)	 &mdash; Commands acting on the ocm-controller
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

