package v1alpha1

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cin/varnish-operator/pkg/logger"

	"go.uber.org/zap"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var webhookLogger = &logger.Logger{SugaredLogger: zap.NewNop().Sugar()}

func SetWebhookLogger(l *logger.Logger) {
	webhookLogger = l
}

var (
	varnishArgsKeyRegexp  = regexp.MustCompile(`^-\w$`)
	disallowedVarnishArgs = map[string]bool{
		"-f": true,
		"-F": true,
		"-n": true,
		"-S": true,
		"-b": true,
	}
	disallowedVarnishArgsAsString string
)

func init() {
	disallowedVarnishArgsAsArr := make([]string, len(disallowedVarnishArgs))
	i := 0
	for k := range disallowedVarnishArgs {
		disallowedVarnishArgsAsArr[i] = k
		i++
	}
	disallowedVarnishArgsAsString = fmt.Sprintf(`"%s"`, strings.Join(disallowedVarnishArgsAsArr, `", "`))
}

func (vc *VarnishCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(vc).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-caching-ibm-com-v1alpha1-varnishcluster,mutating=true,failurePolicy=fail,groups=caching.ibm.com,resources=varnishclusters,verbs=create;update,versions=v1alpha1,name=mvarnishcluster.kb.io

var _ webhook.Defaulter = &VarnishCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (vc *VarnishCluster) Default() {
	logr := webhookLogger.With(logger.FieldComponent, VarnishComponentMutatingWebhook)
	logr = logr.With(logger.FieldNamespace, vc.Namespace)
	logr = logr.With(logger.FieldVarnishCluster, vc.Name)
	logr.Debug("Mutating webhook has been called")

	var defaultReplicasNumber int32 = 1
	if vc.Spec.Replicas == nil {
		vc.Spec.Replicas = &defaultReplicasNumber
	}
}

// note: change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-caching-ibm-com-v1alpha1-varnishcluster,mutating=false,failurePolicy=fail,groups=caching.ibm.com,resources=varnishclusters,versions=v1alpha1,name=vvarnishcluster.kb.io

var _ webhook.Validator = &VarnishCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (vc *VarnishCluster) ValidateCreate() error {
	logr := webhookLogger.With(logger.FieldComponent, VarnishComponentValidatingWebhook)
	logr = logr.With(logger.FieldNamespace, vc.Namespace)
	logr = logr.With(logger.FieldVarnishCluster, vc.Name)

	logr.Debug("Validating webhook has been called on create request")
	return validateCreateUpdate(vc)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (vc *VarnishCluster) ValidateUpdate(old runtime.Object) error {
	logr := webhookLogger.With(logger.FieldComponent, VarnishComponentValidatingWebhook)
	logr = logr.With(logger.FieldNamespace, vc.Namespace)
	logr = logr.With(logger.FieldVarnishCluster, vc.Name)

	logr.Debug("Validating webhook has been called on update request")
	return validateCreateUpdate(vc)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (vc *VarnishCluster) ValidateDelete() error {
	logr := webhookLogger.With(logger.FieldComponent, VarnishComponentValidatingWebhook)
	logr = logr.With(logger.FieldNamespace, vc.Namespace)
	logr = logr.With(logger.FieldVarnishCluster, vc.Name)

	logr.Debug("Validating webhook has been called on delete request")
	return nil
}

func validateCreateUpdate(vc *VarnishCluster) error {
	if vc.Spec.Varnish != nil {
		if err := validVarnishArgs(vc.Spec.Varnish.Args); err != nil {
			return fieldError(".spec.varnish.args", err)
		}
	}

	if vc.Spec.Service != nil {
		if vc.Spec.Service.Port != nil {
			if err := inAllowedRange(int64(*vc.Spec.Service.Port), 1, 65535); err != nil {
				return fieldError(".spec.service.port", err)
			}
		}
		if vc.Spec.Service.MetricsPort != nil {
			if err := inAllowedRange(int64(*vc.Spec.Service.MetricsPort), 1, 65535); err != nil {
				return fieldError(".spec.service.metricsPort", err)
			}
		}
	}

	if vc.Spec.Backend != nil {
		if vc.Spec.Backend.ZoneBalancing != nil {
			for _, threshold := range vc.Spec.Backend.ZoneBalancing.Thresholds {
				if threshold.Local != nil {
					if err := min(int64(*threshold.Local), 1); err != nil {
						return fieldError(".spec.backend.zoneBalancing.thresholds[].local", err)
					}
				}
				if threshold.Remote != nil {
					if err := min(int64(*threshold.Remote), 1); err != nil {
						return fieldError(".spec.backend.zoneBalancing.thresholds[].remote", err)
					}
				}
				if threshold.Local != nil {
					if err := inAllowedRange(int64(*threshold.Threshold), 1, 100); err != nil {
						return fieldError(".spec.backend.zoneBalancing.thresholds[].threshold", err)
					}
				}
			}

			if err := inAllowedRange(int64(*vc.Spec.Service.Port), 1, 65535); err != nil {
				return fieldError(".spec.service.port", err)
			}
		}
	}

	if vc.Spec.UpdateStrategy != nil && vc.Spec.UpdateStrategy.DelayedRollingUpdate != nil {
		if err := min(int64(vc.Spec.UpdateStrategy.DelayedRollingUpdate.DelaySeconds), 1); err != nil {
			return fieldError(".spec.updateStrategy.delayedRollingUpdate.delaySeconds", err)
		}
	}

	return nil
}

func inAllowedRange(port int64, min, max int64) error {
	if port < min || port > max {
		return errors.Errorf("value should be between %d and %d", min, max)
	}

	return nil
}

func min(value int64, minimum int64) error {
	if value < minimum {
		return errors.Errorf("value should be more than %d", minimum)
	}
	return nil
}

func validVarnishArgs(args []string) error {
	for i := 0; i < len(args); {
		if !varnishArgsKeyRegexp.MatchString(args[i]) {
			return errors.Errorf(
				`varnish args must follow pattern: ["key"[, "value"][,"key"[, "value"]]...] where key follows regexp "%s" and value is optional. eg ["-s", "malloc,1024M", "-p", "default_ttl=3600", "-T", "127.0.0.1:6082"]`,
				varnishArgsKeyRegexp.String(),
			)
		}
		if _, found := disallowedVarnishArgs[args[i]]; found {
			return errors.Errorf("cannot include args %s", disallowedVarnishArgsAsString)
		}
		i++
		if i < len(args) && !varnishArgsKeyRegexp.MatchString(args[i]) {
			i++
		}
	}
	return nil
}

func fieldError(fieldName string, err error) error {
	if err == nil {
		return nil
	}
	return errors.Errorf("field %q is invalid: %s", fieldName, err.Error())
}
