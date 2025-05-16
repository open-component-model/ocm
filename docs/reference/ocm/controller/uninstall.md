---
title: "ocm controller uninstall - Uninstalls The Ocm-Controller And All Of Its Dependencies"
linkTitle: "controller uninstall"
url: "/docs/cli-reference/controller/uninstall/"
sidebar:
  collapsed: true
menu:
  docs:
    name: "controller uninstall"
---

### Synopsis

```bash
ocm controller uninstall controller
```

### Options

```text
  -u, --base-url string                       the base url to the ocm-controller's release page (default "https://github.com/ocm-controller/releases")
      --cert-manager-base-url string          the base url to the cert-manager's release page (default "https://github.com/cert-manager/cert-manager/releases")
      --cert-manager-release-api-url string   the base url to the cert-manager's API release page (default "https://api.github.com/repos/cert-manager/cert-manager/releases")
      --cert-manager-version string           version for cert-manager (default "v1.13.2")
  -c, --controller-name string                name of the controller that's used for status check (default "ocm-controller")
  -d, --dry-run                               if enabled, prints the downloaded manifest file
  -h, --help                                  help for uninstall
  -n, --namespace string                      the namespace into which the controller is installed (default "ocm-system")
  -a, --release-api-url string                the base url to the ocm-controller's API release page (default "https://api.github.com/repos/open-component-model/ocm-controller/releases")
  -l, --silent                                don't fail on error
  -t, --timeout duration                      maximum time to wait for deployment to be ready (default 1m0s)
  -p, --uninstall-prerequisites               uninstall prerequisites required by ocm-controller
  -v, --version string                        the version of the controller to install (default "latest")
```

### SEE ALSO

#### Parents

* [ocm controller](ocm_controller.md)	 &mdash; Commands acting on the ocm-controller
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

