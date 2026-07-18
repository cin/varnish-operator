# Monitoring

## Operator Monitoring

The operator is built using the [Kubebuilder SDK](https://github.com/kubernetes-sigs/kubebuilder) which has built-in support for the Prometheus metrics exporter.

A service, created by the operator's Helm chart, exposes the metrics on port `8329` (named `prometheus-metrics`) and can be used to scrape operator metrics.

Additionally, the operator can install a ServiceMonitor configured to scrape operator metrics and a Grafana dashboard with prebuilt dashboard for the operator. The configuration options for them can be specified under `.monitoring` [values override](operator-configuration.md) field of the Helm chart.

### Monitoring Stack Example

The repo includes an [example helm chart](https://github.com/IBM/varnish-operator/tree/main/config/samples/helm-charts/varnish-operator-monitoring) for a Prometheus and Grafana installation that is configured to scrape metrics from the operator and display included dashboards. It depends on the [Prometheus Operator](https://github.com/coreos/prometheus-operator) so it has to be installed first.

After you have the [operator installed](installation.md), clone the repo and install the helm chart.

```bash
$ #install operator with monitoring enabled
$ helm install varnish-operator varnish-operator/varnish-operator --wait \
--set monitoring.grafanaDashboard.enabled=true \
--set monitoring.grafanaDashboard.datasourceName="Prometheus-varnish-operator" \
--set monitoring.prometheusServiceMonitor.enabled=true \
--set monitoring.prometheusServiceMonitor.labels.app=varnish-operator
$ git clone https://github.com/IBM/varnish-operator.git
$ cd varnish-operator/config/samples/helm-charts/varnish-operator-monitoring
$ helm dep build
$ helm install --name varnish-operator-monitoring .
```

No additional configuration needed. The monitoring stack relies on the labels set for the Service that exposes the operator pods.

Port forward your grafana installation:

```bash
$ kubectl port-forward pod/varnish-operator-monitoring-grafana-6f7ff7f4f9-2pjpj 3000
Forwarding from 127.0.0.1:3000 -> 3000
Forwarding from [::1]:3000 -> 3000
```

You can see your dashboard at `localhost:3000`. The login is `admin`, password is `pass`. You will find a dashboard called `Varnish Operator`.

The chart is not a complete solution and intended to be modified to the end user needs.

## Varnish Monitoring

Each Varnish pod has a [Varnish Prometheus metrics exporter](https://github.com/otto-de/prometheus_varnish_exporter) built-in. The exporter port is exposed by the `VarnishCluster` on port `9131` by default. It can be changed by setting the `spec.service.metricsPort` field in the [`VarnishCluster` spec](varnish-cluster-configuration.md).

The service port can be used to setup metrics scraping using [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator) `ServiceMonitor`.  

The pods itself also expose metrics on port `9131`.

There are also metrics exposed by the Varnish controller on a different port. You'll have to setup your ServiceMonitor to scrape metrics from the port 8235 (or refer by its name `ctrl-metrics` 

One of the metrics can be used to setup alerts when the provided VCL failed to compile. It's called `varnish_vcl_compilation_error` and has the value `0` if the last compilation attempt was successful or `1` in case of failure.

### Bundled Grafana dashboard prerequisites

The operator can install a per-`VarnishCluster` Grafana dashboard (`spec.monitoring.grafanaDashboard`, disabled by default). Its panels filter on the `service`, `pod`, and `namespace` labels that Prometheus attaches when scraping through a `ServiceMonitor` — the raw exporter output (`curl <pod>:9131/metrics`) does not contain them. For the dashboard to show data out of the box you need:

* **Prometheus Operator** with a Prometheus instance selecting the cluster's ServiceMonitor. The easiest way is enabling `spec.monitoring.prometheusServiceMonitor` and setting `.labels` to match your Prometheus `serviceMonitorSelector`.
* **Grafana dashboard discovery** picking up the dashboard ConfigMap. With the [Grafana Helm chart](https://github.com/grafana/helm-charts/tree/main/charts/grafana) sidecar, the ConfigMap's default `grafana_dashboard: "1"` label matches the default sidecar config; override via `spec.monitoring.grafanaDashboard.labels` if yours differs.
* **A matching datasource name**: `spec.monitoring.grafanaDashboard.datasourceName` must equal the name of the Prometheus datasource in Grafana (case-sensitive), otherwise all panels show "no data".

If panels are empty but `kubectl exec <pod> -c varnish-metrics-exporter -- wget -qO- localhost:9131/metrics` returns metrics, check that Prometheus actually scrapes the ServiceMonitor target and that queries like `varnish_up{service="<varnishcluster-name>"}` return series in the Grafana Explore view.

### VarnishCluster with Monitoring Stack Example

The repo has a Helm chart example that installs a simple backend and VarnishCluster to cache requests. Additionally, it installs Prometheus with a pre-configured Grafana instance to monitor it. This chart depends on the Prometheus operator so it must be installed first. 

It can be installed by cloning the repo and installing the chart:

```bash
$ git clone https://github.com/IBM/varnish-operator.git
$ cd varnish-operator/config/samples/helm-charts/varnishcluster-with-monitoring
$ helm dep build
$ helm install --name varnish-test .
```

Port forward the Grafana pod to see the dashboard:

```bash
$ kubectl port-forward pod/varnish-test-grafana-9f584598d-89smp 3000
Forwarding from 127.0.0.1:3000 -> 3000
Forwarding from [::1]:3000 -> 3000
```

You can see your dashboard at `localhost:3000`. The login is `admin`, password is `pass`. You will find a dashboard called `Varnish`.

In another terminal port forward your Varnish service:

```bash
$ kubectl port-forward svc/varnish-test-varnish 8080:80
```

Make some requests to see metrics visualized in your Grafana dashboard.

The chart is not a complete solution and intended to be modified to the end user needs.
