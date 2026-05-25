package predicates

import (
	"github.com/cin/varnish-operator/pkg/logger"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type debugPredicate struct {
	logger *logger.Logger
}

func NewDebugPredicate(logr *logger.Logger) predicate.TypedPredicate[client.Object] {
	return &debugPredicate{
		logger: logr,
	}
}

func (p *debugPredicate) Create(e event.TypedCreateEvent[client.Object]) bool {
	p.logger.Debugf("Create event for resource %T: %s/%s", e.Object, e.Object.GetNamespace(), e.Object.GetName())
	return true
}

func (p *debugPredicate) Delete(e event.TypedDeleteEvent[client.Object]) bool {
	p.logger.Debugf("Delete event for resource %T: %s/%s", e.Object, e.Object.GetNamespace(), e.Object.GetName())
	return true
}

func (p *debugPredicate) Update(e event.TypedUpdateEvent[client.Object]) bool {
	p.logger.Debugf("Update event for resource %T: %s/%s", e.ObjectNew, e.ObjectNew.GetNamespace(), e.ObjectNew.GetName())
	return true
}

func (p *debugPredicate) Generic(e event.TypedGenericEvent[client.Object]) bool {
	p.logger.Debugf("Generic event for resource %T: %s/%s", e.Object, e.Object.GetNamespace(), e.Object.GetName())
	return true
}
