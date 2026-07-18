package controller

import (
	"encoding/json"
	"strings"
	"testing"

	vcapi "github.com/cin/varnish-operator/api/v1alpha1"
	"github.com/gogo/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Renders the bundled Grafana dashboard template and validates the output is
// well-formed JSON with the expected PromQL wiring. Catches template syntax
// errors and malformed JSON without needing envtest.
func TestGenerateGrafanaDashboardData(t *testing.T) {
	instance := &vcapi.VarnishCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "test-ns",
		},
		Spec: vcapi.VarnishClusterSpec{
			Monitoring: &vcapi.VarnishClusterMonitoring{
				GrafanaDashboard: &vcapi.VarnishClusterMonitoringGrafanaDashboard{
					Enabled:        true,
					Title:          "Test Dashboard",
					DatasourceName: proto.String("Prometheus"),
				},
			},
		},
	}

	data, err := generateGrafanaDashboardData(instance)
	if err != nil {
		t.Fatalf("failed to render dashboard template: %v", err)
	}

	if len(data) != 1 {
		t.Fatalf("expected exactly one dashboard file, got %d", len(data))
	}

	var rendered string
	for _, v := range data {
		rendered = v
	}

	var dashboard map[string]interface{}
	if err := json.Unmarshal([]byte(rendered), &dashboard); err != nil {
		t.Fatalf("rendered dashboard is not valid JSON: %v", err)
	}

	if title := dashboard["title"]; title != "Test Dashboard" {
		t.Errorf("expected title %q, got %v", "Test Dashboard", title)
	}

	// PromQL exprs live inside JSON strings, so their quotes stay escaped (\") in the raw text.
	for _, substr := range []string{
		`service=\"test-cluster\"`,
		`namespace=\"test-ns\"`,
		`"datasource": "Prometheus"`,
		"{{version}}", // version panel shows the version from the varnish_version const label
	} {
		if !strings.Contains(rendered, substr) {
			t.Errorf("rendered dashboard is missing %q", substr)
		}
	}
}
