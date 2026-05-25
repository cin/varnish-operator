# Custom container images

Each `VarnishCluster` pod runs three containers (`varnish`, `varnish-controller`, `varnish-metrics-exporter`). You can point them at your own images via the CR spec. The operator does **not** validate image contents; images must satisfy the [compatibility requirements](#compatibility-requirements-for-custom-images) below or pods will fail readiness checks, VCL reloads, or metrics scraping.

### Image tags vs Varnish version

**Container image tags follow your operator release** (Helm `container.tag`, CI `VERSION`, e.g. `0.38.0` or `local`)—not the upstream Varnish semver. All three images for a given operator release are built and published together with that tag.

The **Varnish daemon version** inside the `varnish` image (today **9.0.3** via `VARNISH_VERSION_NUMBER` at build time) is separate from the OCI tag. The controller and metrics exporter images are operator components; tag them like `varnish-controller:0.38.0`, not `varnish-metrics-exporter:9.0.3`.

## Specifying images on `VarnishCluster`

Set `spec.varnish.image` for the `varnishd` container. Sidecar images default from that name unless overridden explicitly.

```yaml
spec:
  varnish:
    image: registry.example.com/acme/varnish:0.38.0
    imagePullPolicy: IfNotPresent
    imagePullSecret: regcred   # optional, applies to all containers in the pod
    controller:
      image: registry.example.com/acme/varnish-controller:0.38.0
      imagePullPolicy: IfNotPresent
    metricsExporter:
      image: registry.example.com/acme/varnish-metrics-exporter:0.38.0
      imagePullPolicy: IfNotPresent
```

| Field | Purpose |
| ----- | ------- |
| `spec.varnish.image` | `varnishd` container image (repository:tag). |
| `spec.varnish.imagePullPolicy` | Pull policy for the `varnish` container (default: `Always`). |
| `spec.varnish.imagePullSecret` | Secret used to pull images for the StatefulSet pod. |
| `spec.varnish.controller.image` | Controller sidecar; if empty, derived from `varnish.image` (see below). |
| `spec.varnish.metricsExporter.image` | Metrics sidecar; if empty, derived from `varnish.image` (see below). |

### Default image names when fields are omitted

If `spec.varnish.image` is **empty**, the operator uses a **coupled default** derived from the operator Deployment image (`CONTAINER_IMAGE` / Helm `container.image`):

| Operator image | Default `varnish` image | Default controller | Default metrics exporter |
| -------------- | ----------------------- | ------------------ | ------------------------ |
| `registry.io/team/varnish-operator:1.2.0` | `registry.io/team/varnish:1.2.0` | `registry.io/team/varnish-controller:1.2.0` | `registry.io/team/varnish-metrics-exporter:1.2.0` |

The repository path is taken from the operator image; the image name is replaced with `varnish`, and the **same tag** is reused. Controller and exporter names append `-controller` and `-metrics-exporter` to the repository name before the tag.

If you set only `spec.varnish.image`, sidecars use that naming rule automatically:

```yaml
spec:
  varnish:
    image: myregistry/varnish:0.38.0
    # controller  -> myregistry/varnish-controller:0.38.0
    # metrics     -> myregistry/varnish-metrics-exporter:0.38.0
```

Override individual sidecars when your registry naming differs (keep the **same operator release tag** on all three unless you know the builds are paired):

```yaml
spec:
  varnish:
    image: myregistry/varnish:0.38.0
    controller:
      image: myregistry/varnish-ctrl:0.38.0
```

After changing an image tag in a running cluster, set `spec.statefulSet.container.imagePullPolicy: Always` (on the embedded pod template, if you customize it) or delete pods so nodes pull new layers.

## Compatibility requirements for custom images

All three images must target the **same Varnish major/minor** (today: **9.0.3** from [packages.varnish-software.com](https://packages.varnish-software.com/) in this repository’s Dockerfiles). Mixing `varnishd` 9.x with sidecars built against 7.x will break `varnishadm`, `varnishstat`, and the Prometheus exporter.

### `varnish` (`varnishd`) image

| Requirement | Detail |
| ----------- | ------ |
| **Process** | `varnishd` on `PATH` or `/usr/sbin/varnishd` (StatefulSet does not override the image entrypoint). |
| **Vmods** | If you use the operator’s default VCL (`import var;`, `import directors;` in templates), install **`varnish-modules`** for your Varnish version. |
| **User / UID** | Run as non-root user **`varnish` with UID/GID 1000** (Varnish Software packages). The operator sets `runAsUser` / `runAsGroup` / `fsGroup` to **1000**. |
| **Directories** | Writable `/var/lib/varnish` (instance dir `-n`) and `/etc/varnish` (VCL mount). |
| **Args** | The operator injects `-F`, `-n /var/lib/varnish`, `-S /etc/varnish-secret/secret`, `-T 0.0.0.0:6082`, `-b 127.0.0.1:0`, `-a 0.0.0.0:<service.port>`. Do not rely on overriding `-n`, `-f`, `-S`, `-T`, or `-b` via `spec.varnish.args`. |

### `varnish-controller` image

| Requirement | Detail |
| ----------- | ------ |
| **Binary** | `/varnish-controller` as entrypoint (built from this repo’s `cmd/varnish-controller`). |
| **CLI tools** | `varnishadm` and `varnishstat` compatible with the `varnishd` version. |
| **Libraries** | Matching `libvarnishapi` (e.g. `libvarnishapi.so.3` on 9.x). |
| **User** | UID **1000** (`varnish`), with read access to the shared workdir and secret volume. |

### `varnish-metrics-exporter` image

| Requirement | Detail |
| ----------- | ------ |
| **Binary** | `prometheus-varnish-exporter` (this repo builds [otto-de/prometheus_varnish_exporter](https://github.com/otto-de/prometheus_varnish_exporter) v1.8.3+). |
| **CLI tools** | `varnishstat` + matching `libvarnishapi`. |
| **User** | UID **1000** (`varnish`). |
| **Args** | Operator passes **`-n /var/lib/varnish`**; the image must support that flag. |

### VCL and configuration (not in the image, but required)

Custom images do not remove the need for compatible **VCL** in the ConfigMap (`spec.vcl`). Review [Varnish 9 upgrade notes](https://varnish-cache.org/docs/9.0/whats-new/upgrading-9.0.html) when moving from older versions.

## What you cannot change via image alone

These are fixed in the operator today; custom images must adapt to them (or you fork the operator):

- **Security context**: `runAsUser` / `runAsGroup` / `fsGroup` **1000** on Varnish pods.
- **Volume layout**: shared `emptyDir` at `/var/lib/varnish`, VCL under `/etc/varnish`, admin secret at `/etc/varnish-secret/secret`.
- **Ports**: Varnish HTTP from `spec.service.port`, admin CLI on **6082**, metrics on **9131** (defaults).
- **Readiness probe**: `varnishadm -S … -T 127.0.0.1:6082 ping`.

## Building images from this repository

Reference Dockerfiles (Debian trixie + Varnish Software packages):

| Image | Dockerfile | Install script |
| ----- | ---------- | -------------- |
| `varnish` | `Dockerfile.varnishd` | `docker/install-varnish-9.sh` (`minimal`: `varnish` + `varnish-modules`) |
| `varnish-controller` | `Dockerfile.controller` | `install-varnish-9.sh` (`tools`: CLI + libs) + Go binary |
| `varnish-metrics-exporter` | `Dockerfile.exporter` | `tools` + exporter build |

```bash
make docker-build-pod REPO=myregistry VERSION=0.38.0
# myregistry/varnish:0.38.0
# myregistry/varnish-controller:0.38.0
# myregistry/varnish-metrics-exporter:0.38.0
```

Pin a different **Varnish package** inside the `varnish` image (does not change the OCI tag):

```bash
make docker-build-varnish VARNISH_VERSION_NUMBER=9.0.3-1 REPO=myregistry VERSION=0.38.0
```

Build args: `VERSION` (image tag / operator release), `VARNISH_VERSION_NUMBER` (Debian package pin, default `9.0.3-1`), `PROMETHEUS_VARNISH_EXPORTER_VERSION` (default `v1.8.3`).

## Using upstream `library/varnish` images

The official [Docker Hub `varnish` image](https://hub.docker.com/_/varnish) also uses UID **1000** and Varnish Software packages, but it is **not** tested as a drop-in for this operator: entrypoint scripts, bundled VCL, and optional vmods differ from the images built here. Prefer the Dockerfiles in this repo, or replicate their layout (users, paths, `varnishadm`/`varnishstat`, `varnish-modules`) and run `make e2e-tests` before production use.

## Related documentation

- [VarnishCluster configuration](varnish-cluster-configuration.md) — full spec field list
- [Development / local images](development.md) — kind, `make e2e-tests`, local tags
- [Architecture](architecture.md) — pod layout and component roles
