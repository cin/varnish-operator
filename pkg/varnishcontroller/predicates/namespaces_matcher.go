package predicates

import (
	"github.com/cin/varnish-operator/pkg/logger"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.TypedPredicate[*v1.Pod] = &NamespacesMatcherPredicate{}

type NamespacesMatcherPredicate struct {
	logger     *logger.Logger
	Namespaces []string
}

func NewNamespacesMatcherPredicate(namespaces []string, logr *logger.Logger) *NamespacesMatcherPredicate {
	if logr == nil {
		logr = logger.NewNopLogger()
	}
	return &NamespacesMatcherPredicate{
		logger:     logr,
		Namespaces: namespaces,
	}
}

func (p *NamespacesMatcherPredicate) Create(e event.TypedCreateEvent[*v1.Pod]) bool {
	return p.allow(e.Object.GetNamespace())
}

func (p *NamespacesMatcherPredicate) Delete(e event.TypedDeleteEvent[*v1.Pod]) bool {
	return p.allow(e.Object.GetNamespace())
}

func (p *NamespacesMatcherPredicate) Update(e event.TypedUpdateEvent[*v1.Pod]) bool {
	return p.allow(e.ObjectNew.GetNamespace())
}

func (p *NamespacesMatcherPredicate) Generic(e event.TypedGenericEvent[*v1.Pod]) bool {
	return p.allow(e.Object.GetNamespace())
}

func (p *NamespacesMatcherPredicate) allow(namespace string) bool {
	if len(p.Namespaces) == 0 {
		return true
	}
	return contains(namespace, p.Namespaces)
}

func contains(v string, s []string) bool {
	if s == nil {
		return false
	}

	for _, value := range s {
		if value == v {
			return true
		}
	}

	return false
}
