package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo/v2/types"

	v1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	vcapi "github.com/cin/varnish-operator/api/v1alpha1"
	"github.com/cin/varnish-operator/pkg/logger"
	"k8s.io/client-go/rest"

	"go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const defaultE2EKubeconfig = "e2e-tests-kubeconfig"

func e2eRestConfig() (*rest.Config, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		if _, err := os.Stat(defaultE2EKubeconfig); err == nil {
			abs, err := filepath.Abs(defaultE2EKubeconfig)
			if err != nil {
				return nil, err
			}
			kubeconfig = abs
		}
	}

	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	return ctrl.GetConfig()
}

func e2eConfigHint(err error) string {
	return fmt.Sprintf(`%v

E2E tests need a Kubernetes cluster with the operator installed (not unit/envtest).

From the repo root, either run the full workflow:
  make e2e-tests

Or prepare the cluster and re-run tests:
  ./hack/create_dev_cluster.sh
  KUBECONFIG=./e2e-tests-kubeconfig go test ./tests
  ./hack/delete_dev_cluster.sh

For an existing cluster, set KUBECONFIG or install the operator per docs/development.md.`, err)
}

const (
	debugLogsDir = "/tmp/debug-logs/"
)

var (
	k8sClient         client.Client
	restConfig        *rest.Config
	kubeClient        *kubernetes.Clientset
	tailLines         int64 = 30
	operatorPodLabels       = map[string]string{"operator": "varnish-operator"}
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(func(o *zap.Options) { o.DestWriter = GinkgoWriter }))
	logr := logger.NewLogger("console", zapcore.DebugLevel)
	By("bootstrapping test environment")

	var err error
	err = vcapi.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	restConfig, err = e2eRestConfig()
	if err != nil {
		logr.Fatalf("%s", e2eConfigHint(err))
	}

	k8sClient, err = client.New(restConfig, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	// Create client test. We use kubernetes package bc currently only it has GetLogs method.
	kubeClient, err = kubernetes.NewForConfig(restConfig)
	Expect(err).ToNot(HaveOccurred())
})

var _ = JustAfterEach(func() {
	if CurrentSpecReport().Failed() && CurrentSpecReport().State != types.SpecStateInterrupted {
		fmt.Printf("Test failed! Collecting diags just after failed test in %s:%d\n", CurrentSpecReport().FileName(), CurrentSpecReport().LineNumber())

		Expect(os.MkdirAll(debugLogsDir, 0777)).To(Succeed())
		vcList := &vcapi.VarnishClusterList{}
		Expect(k8sClient.List(context.Background(), vcList)).To(Succeed())

		GinkgoWriter.Println("Gathering log info for Varnish Operator")
		showPodLogs(operatorPodLabels, "varnish-operator")

		for _, vc := range vcList.Items {
			GinkgoWriter.Printf("Gathering log info for VarnishCluster %s/%s\n", vc.Namespace, vc.Name)
			podList := &v1.PodList{}
			Expect(k8sClient.List(context.Background(), podList, client.InNamespace(vc.Namespace))).To(Succeed())

			for _, pod := range podList.Items {
				fmt.Println("Pod: ", pod.Name, " Status: ", pod.Status.Phase)
				for _, container := range pod.Status.ContainerStatuses {
					fmt.Println("Container: ", container.Name, " Ready: ", container.State)
				}
			}

			showPodLogs(map[string]string{vcapi.LabelVarnishOwner: vc.Name}, vc.Namespace)
			showPodLogs(vc.Spec.Backend.Selector, vc.Namespace)
		}

		showClusterEvents()
	}
})
