# Upgrading to Varnish 9

This guide is for operators of existing `VarnishCluster` deployments running images based on Varnish 6.x or 7.x who are moving to an operator release that ships **Varnish 9** (currently **9.0.3** from [packages.varnish-software.com](https://packages.varnish-software.com/)).

Read the whole guide before starting: some changes (VCL syntax, custom image UIDs) must be prepared **before** rolling any pods.

## TL;DR checklist

- [ ] Review your VCL against the [Varnish 7](https://varnish-cache.org/docs/7.0/whats-new/upgrading-7.0.html) and [Varnish 9](https://varnish-cache.org/docs/9.0/whats-new/upgrading-9.0.html) upgrade notes; test compilation against a Varnish 9 image.
- [ ] If you use custom images or persistent volumes: prepare for the **UID/GID change from 997 to 1000**.
- [ ] Upgrade all three pod images (`varnish`, `varnish-controller`, `varnish-metrics-exporter`) **together**, using the same operator release tag.
- [ ] Plan for a **cold cache** during the rollout.
- [ ] Pick an update strategy (`OnDelete` default, `RollingUpdate`, or `DelayedRollingUpdate`) that matches your traffic pattern.
- [ ] After the rollout, verify metrics and dashboards (see [Monitoring](monitoring.md)).

## Images and version coupling

Container image tags follow the **operator release** (e.g. `0.38.0`), not the Varnish semver — see [Custom container images](custom-images.md). All three images in a Varnish pod must be built against the same Varnish version:

| Container | Why it must match |
| --------- | ----------------- |
| `varnish` | Runs `varnishd` itself. |
| `varnish-controller` | Uses `varnishadm` / `varnishstat` and `libvarnishapi` (`libvarnishapi.so.3` on 9.x). |
| `varnish-metrics-exporter` | Uses `varnishstat` + `libvarnishapi` to scrape counters. |

Mixing a 9.x `varnishd` with sidecars built against 6.x/7.x breaks VCL reloads, readiness checks, and metrics. If you override any image in `spec.varnish`, update all overridden images in the same change.

By default (no image overrides), the operator derives pod images from its own deployment image, so upgrading the operator Helm release is enough — but the pods still need to be restarted to pick up the new images (see [Rollout](#rollout) below).

## UID/GID change: 997 → 1000

Older operator releases ran Varnish pod containers as UID/GID **997**. Varnish 9 images use the **`varnish` user (UID/GID 1000)** from the Varnish Software packages, and the operator now sets `runAsUser` / `runAsGroup` / `fsGroup` to **1000** on the StatefulSet.

Impact:

- **Custom images** must contain a user with UID/GID 1000 with read access to the shared workdir and secret volumes. See the [compatibility requirements](custom-images.md#compatibility-requirements-for-custom-images).
- **Persistent volumes** (e.g. if you use the `file` storage backend or mount extra volumes into Varnish pods): files owned by 997 may become unreadable/unwritable. `fsGroup: 1000` handles volumes that support fsGroup-based relabeling; for volumes that do not (e.g. `hostPath`, some NFS setups), chown the data to 1000 before or during the rollout.
- **PodSecurityPolicies / security admission** rules pinned to UID 997 need updating.

The default cache workdir is an `emptyDir` recreated with each pod, so it needs no migration.

## VCL changes

Your VCL in the ConfigMap (`spec.vcl`) skips two major versions. Review the upstream upgrade notes for everything between your current and target versions, especially:

- **[Varnish 7](https://varnish-cache.org/docs/7.0/whats-new/upgrading-7.0.html)**: regular expressions moved to **PCRE2** (most patterns work unchanged, but some syntax/behavior differs), stricter VCL parsing, and changed defaults.
- **[Varnish 9](https://varnish-cache.org/docs/9.0/whats-new/upgrading-9.0.html)**: removed deprecated APIs and parameters, backend **probe** behavior changes, and vmod updates.

Practical tips:

- Test-compile your VCL against a Varnish 9 image before touching the cluster:

  ```bash
  docker run --rm -v "$PWD/vcl:/etc/varnish:ro" <your-varnish-9-image> \
    varnishd -C -f /etc/varnish/entrypoint.vcl
  ```

- If you use the operator's default VCL (no ConfigMap yet), no action is needed; the operator seeds a Varnish 9-compatible default.
- After the rollout, watch the `varnish_vcl_compilation_error` metric (see [Monitoring](monitoring.md)); it flips to `1` if a VCL reload fails on the new version.

## Cache and rollout

### Cold cache

Varnish is an in-memory cache and the default workdir is an `emptyDir`; every restarted pod starts with an **empty cache**. Expect elevated backend traffic and higher latency until the cache warms up. If your backends cannot absorb a full cache flush, roll pods gradually (see below) or pre-warm via synthetic traffic.

### Rollout

The default update strategy is **`OnDelete`**: after upgrading the operator (and/or `spec.varnish.image`), pods keep running the old images until you delete them. This gives you full control over pacing:

```bash
kubectl delete pod <varnishcluster-name>-varnish-2   # highest ordinal first
# wait for readiness + cache warm-up, then continue with the next pod
```

StatefulSet ordering applies: with `RollingUpdate`-style strategies pods restart from the highest ordinal to the lowest. Readiness is gated on `varnishadm ping`, so a pod only receives traffic once `varnishd` responds on its admin interface — but readiness does **not** mean the cache is warm.

For automated pacing use **`DelayedRollingUpdate`**, which waits a configurable delay between pod updates so each node can warm up before the next restarts:

```yaml
spec:
  updateStrategy:
    type: DelayedRollingUpdate
    delayedRollingUpdate:
      delaySeconds: 300
```

See [VarnishCluster](varnish-cluster.md) for details on update strategies.

If you need to drain traffic away from a pod before deleting it, remove it from the cache Service selector-based endpoints by cordoning at the LB level or scaling client traffic; the operator does not provide a built-in drain today.

## Monitoring

The bundled metrics exporter ([otto-de/prometheus_varnish_exporter](https://github.com/otto-de/prometheus_varnish_exporter) v1.8.3) scrapes `varnishstat` dynamically, so core metrics keep working on Varnish 9. Some counter names change between Varnish major versions — review any custom alerts and dashboards against a running 9.x pod (`curl <pod>:9131/metrics`).

The optional per-cluster Grafana dashboard (`spec.monitoring.grafanaDashboard`) requires the Prometheus Operator `ServiceMonitor` (`spec.monitoring.prometheusServiceMonitor`) plus Grafana dashboard discovery — see [Monitoring](monitoring.md).

## Verifying the upgrade

```bash
# All pods ready on the new images
kubectl get pods -n <ns> -l varnish-component=varnish -o wide

# Varnish version inside a pod
kubectl exec -n <ns> <pod> -c varnish -- varnishd -V

# VCL compiled and active
kubectl exec -n <ns> <pod> -c varnish-controller -- varnishadm -n /var/lib/varnish vcl.list

# Metrics flowing
kubectl exec -n <ns> <pod> -c varnish-metrics-exporter -- wget -qO- localhost:9131/metrics | head
```

`varnishd -V` should report `varnish-9.0.3`, and responses through the cache Service should carry `Via: 1.1 <pod-name> (Varnish/9.0)`.
