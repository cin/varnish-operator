# Varnish Operator

[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/5895/badge)](https://bestpractices.coreinfrastructure.org/projects/5895)

### Project status: alpha
The project is in development and breaking changes can be introduced.

The purpose of the project is to provide a convenient way to deploy and manage Varnish instances in Kubernetes.

Kubernetes version `>=1.29.0` is supported (see the operator bundle `minKubeVersion`). CI runs e2e against Kubernetes 1.34.3 and 1.35.1, and unit tests use envtest 1.36.0—see [docs/development.md](docs/development.md#kubernetes-versions-in-tests) for why those version numbers differ.

Varnish version 7.x is supported (container images ship Debian trixie packages, currently Varnish 7.7).

Full documentation can be found [here](https://cin.github.io/varnish-operator/)

### Overview

Varnish operator manages Varnish clusters using a CustomResourceDefinition that defines a new Kind called `VarnishCluster`. 

The operator manages the whole lifecycle of the cluster: creating, deleting and keeping the cluster configuration up to date. The operator is responsible for building the VCL configuration using templates defined by the users and keeping the configuration up to date when relevant events occur (backend pod failure, scaling of the deployment, VCL configuration change).

## Features

 * [x] Basic install
 * [x] Full lifecycle support (create/update/delete)
 * [x] Automatic VCL configuration updates (using user defined templates)
 * [x] Prometheus metrics support
 * [x] Scaling
 * [x] Configurable update strategy
 * [x] Persistence (for [file storage backend](https://varnish-cache.org/docs/trunk/users-guide/storage-backends.html#file) support)
 * [ ] Multiple Varnish versions support
 * [ ] Autoscaling

### Further reading

* [QuickStart](https://cin.github.io/varnish-operator/quick-start.html)
* [Contributing](https://cin.github.io/varnish-operator/development.html)
