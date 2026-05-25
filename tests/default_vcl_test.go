package tests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	vcapi "github.com/cin/varnish-operator/api/v1alpha1"

	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Exercises the operator-seeded ConfigMap (entrypoint.vcl + backends.vcl.tmpl), not a user-supplied VCL.
var _ = Describe("operator default VCL", func() {
	vcNamespace := "default"
	vcName := "default-vcl-test"
	configMapName := "e2e-operator-default-vcl"
	objMeta := metav1.ObjectMeta{
		Namespace: vcNamespace,
		Name:      vcName,
	}
	backendResponse := "DEFAULT-VCL-BACKEND"
	backendLabels := map[string]string{"app": "default-vcl-backend"}
	backendDeploymentName := "default-vcl-backend"
	varnishPodLabels := map[string]string{
		vcapi.LabelVarnishOwner:      vcName,
		vcapi.LabelVarnishComponent: vcapi.VarnishComponentVarnish,
	}

	backendsDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backendDeploymentName,
			Namespace: vcNamespace,
			Labels:    backendLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: proto.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: backendLabels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: backendLabels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "backend",
							Image: "hashicorp/http-echo",
							Ports: []v1.ContainerPort{
								{
									Name:          "web",
									Protocol:      v1.ProtocolTCP,
									ContainerPort: 5678,
								},
							},
							Args: []string{fmt.Sprintf("-text=%s", backendResponse)},
						},
					},
				},
			},
		},
	}

	backendPort := intstr.FromInt(5678)
	vc := &vcapi.VarnishCluster{
		ObjectMeta: objMeta,
		Spec: vcapi.VarnishClusterSpec{
			Backend: &vcapi.VarnishClusterBackend{
				Selector: backendLabels,
				Port:     &backendPort,
			},
			Service: &vcapi.VarnishClusterService{
				Port: proto.Int32(9090),
			},
			Varnish: &vcapi.VarnishClusterVarnish{
				ImagePullPolicy: v1.PullNever,
				Controller: &vcapi.VarnishClusterVarnishController{
					ImagePullPolicy: v1.PullNever,
				},
				MetricsExporter: &vcapi.VarnishClusterVarnishMetricsExporter{
					ImagePullPolicy: v1.PullNever,
				},
			},
			VCL: &vcapi.VarnishClusterVCL{
				ConfigMapName:      proto.String(configMapName),
				EntrypointFileName: proto.String("entrypoint.vcl"),
			},
		},
	}

	AfterEach(func() {
		By("deleting created resources")
		Expect(k8sClient.DeleteAllOf(context.Background(), &vcapi.VarnishCluster{}, client.InNamespace(vcNamespace))).To(Succeed())
		Expect(k8sClient.DeleteAllOf(context.Background(), &appsv1.Deployment{}, client.InNamespace(vcNamespace), client.MatchingLabels(backendLabels))).To(Succeed())
		Expect(k8sClient.DeleteAllOf(context.Background(), &v1.ConfigMap{}, client.InNamespace(vcNamespace), client.MatchingLabels(map[string]string{
			vcapi.LabelVarnishOwner: vcName,
		}))).To(Succeed())
		waitForPodsTermination(vcNamespace, varnishPodLabels)
		waitForPodsTermination(vcNamespace, backendLabels)
		waitUntilVarnishClusterRemoved(vcName, vcNamespace)
	})

	It("seeds ConfigMap VCL and serves traffic on Varnish 9", func() {
		Expect(k8sClient.Create(context.Background(), backendsDeployment)).To(Succeed())
		Expect(k8sClient.Create(context.Background(), vc)).To(Succeed())

		cm := &v1.ConfigMap{}
		Eventually(func() error {
			return k8sClient.Get(context.Background(), types.NamespacedName{
				Name: configMapName, Namespace: vcNamespace,
			}, cm)
		}, time.Minute, time.Second*2).Should(Succeed())

		By("operator created default VCL files")
		Expect(cm.Data).To(HaveKey("entrypoint.vcl"))
		Expect(cm.Data).To(HaveKey("backends.vcl.tmpl"))
		entrypoint := cm.Data["entrypoint.vcl"]
		Expect(entrypoint).To(ContainSubstring("return (ok);"))
		Expect(entrypoint).To(ContainSubstring("return (hash);"))
		Expect(entrypoint).To(ContainSubstring("import var;"))
		Expect(entrypoint).To(ContainSubstring(`include "backends.vcl"`))
		Expect(cm.Data["backends.vcl.tmpl"]).To(ContainSubstring("import directors;"))

		By("backend pods become ready")
		waitForPodsReadiness(vcNamespace, backendLabels)
		By("varnish pods become ready")
		waitForPodsReadiness(vcNamespace, varnishPodLabels)

		pf := portForwardPod(vcNamespace, varnishPodLabels, []string{"6081:6081"})
		defer pf.Close()

		By("default /heartbeat endpoint")
		Eventually(func() (int, error) {
			resp, err := http.Get("http://localhost:6081/heartbeat")
			if err != nil {
				return 0, err
			}
			defer func() { _ = resp.Body.Close() }()
			return resp.StatusCode, nil
		}, time.Minute, time.Second*2).Should(Equal(200))

		By("default /liveness endpoint with healthy backend")
		Eventually(func() (int, error) {
			resp, err := http.Get("http://localhost:6081/liveness")
			if err != nil {
				return 0, err
			}
			defer func() { _ = resp.Body.Close() }()
			return resp.StatusCode, nil
		}, time.Minute, time.Second*2).Should(Equal(200))

		By("cached backend response with X-Varnish-Cache")
		var resp *http.Response
		Eventually(func() (int, error) {
			var err error
			resp, err = http.Get("http://localhost:6081/cached-path")
			if err != nil {
				return 0, err
			}
			return resp.StatusCode, nil
		}, time.Second*30, time.Second*2).Should(Equal(200))
		Expect(resp.Header.Get("X-Varnish-Cache")).To(Equal("MISS"))
		body, err := io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.TrimSpace(string(body))).To(Equal(backendResponse))

		resp, err = http.Get("http://localhost:6081/cached-path")
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))
		Expect(resp.Header.Get("X-Varnish-Cache")).To(Equal("HIT"))
	})
})
