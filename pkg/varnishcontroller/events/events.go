package events

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeEvents "k8s.io/client-go/tools/events"
)

const (
	EventRecorderName = "varnish"

	EventReasonReloadError         EventReason = "ReloadError"
	EventReasonVCLCompilationError EventReason = "VCLCompilationError"
	EventReasonInvalidVCLConfigMap EventReason = "InvalidVCLConfigMap"
	EventReasonBackendIgnored      EventReason = "BackendIgnored"
)

// EventReason is the reason why the event was create. The value appears in the 'Reason' tab of the events list
type EventReason string

// NewEventHandler creates a new event handler that will use the specified recorder
func NewEventHandler(recorder kubeEvents.EventRecorder, podName string) *EventHandler {
	return &EventHandler{
		Recorder: recorder,
		podName:  podName,
	}
}

// EventHandler handles the operations for events
type EventHandler struct {
	Recorder kubeEvents.EventRecorder
	podName  string
}

// Warning creates a 'warning' type event
func (e *EventHandler) Warning(object runtime.Object, reason EventReason, message string) {
	e.Recorder.Eventf(object, nil, v1.EventTypeWarning, string(reason), e.podName, message)
}

// Normal creates a 'normal' type event
func (e *EventHandler) Normal(object runtime.Object, reason EventReason, message string) {
	e.Recorder.Eventf(object, nil, v1.EventTypeNormal, string(reason), e.podName, message)
}
