package predicates

import (
	"github.com/cin/varnish-operator/pkg/logger"
	"github.com/cin/varnish-operator/pkg/varnishcontroller/podutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.TypedPredicate[*v1.Pod] = &LabelMatcherPredicate{}

type LabelMatcherPredicate struct {
	logger   *logger.Logger
	Selector labels.Selector
}

func NewLabelMatcherPredicate(selector labels.Selector, logr *logger.Logger) *LabelMatcherPredicate {
	if logr == nil {
		logr = logger.NewNopLogger()
	}
	return &LabelMatcherPredicate{
		logger:   logr,
		Selector: selector,
	}
}

func (p *LabelMatcherPredicate) Create(e event.TypedCreateEvent[*v1.Pod]) bool {
	return p.Selector.Matches(labels.Set(e.Object.GetLabels()))
}

func (p *LabelMatcherPredicate) Delete(e event.TypedDeleteEvent[*v1.Pod]) bool {
	return p.Selector.Matches(labels.Set(e.Object.GetLabels()))
}

func (p *LabelMatcherPredicate) Update(e event.TypedUpdateEvent[*v1.Pod]) bool {
	if !p.Selector.Matches(labels.Set(e.ObjectNew.GetLabels())) {
		return false
	}

	oldPod := e.ObjectOld
	newPod := e.ObjectNew
	if oldPod == nil || newPod == nil {
		return true
	}

	if len(newPod.Status.PodIP) != 0 && oldPod.Status.PodIP != newPod.Status.PodIP {
		return true
	}

	if len(newPod.Spec.NodeName) != 0 && oldPod.Spec.NodeName != newPod.Spec.NodeName {
		return true
	}

	if podutil.PodReady(*newPod) != podutil.PodReady(*oldPod) {
		return true
	}

	return false
}

func (p *LabelMatcherPredicate) Generic(e event.TypedGenericEvent[*v1.Pod]) bool {
	return p.Selector.Matches(labels.Set(e.Object.GetLabels()))
}
