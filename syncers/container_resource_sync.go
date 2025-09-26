package syncers

import (
	"fmt"

	"github.com/loft-sh/vcluster/pkg/syncer/synccontext"
	syncertypes "github.com/loft-sh/vcluster/pkg/syncer/types"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewContainerResourceSyncer(ctx *synccontext.RegisterContext) syncertypes.Base {
	return &containerResourceSyncer{}
}

type containerResourceSyncer struct{}

func (s *containerResourceSyncer) Name() string {
	return "container-resource-syncer"
}

func (s *containerResourceSyncer) Resource() client.Object {
	return &corev1.Pod{}
}

func (s *containerResourceSyncer) Syncer() syncertypes.Sync[client.Object] {
	return &containerResourceSync{}
}

type containerResourceSync struct{}

func (s *containerResourceSync) SyncToHost(ctx *synccontext.SyncContext, event *synccontext.SyncToHostEvent[client.Object]) (ctrl.Result, error) {
	// This syncer only works on physical to virtual sync, so we don't need to implement this
	return ctrl.Result{}, nil
}

func (s *containerResourceSync) Sync(ctx *synccontext.SyncContext, event *synccontext.SyncEvent[client.Object]) (ctrl.Result, error) {
	pPod := event.Host.(*corev1.Pod)
	vPod := event.Virtual.(*corev1.Pod)

	updated := s.updateContainerResources(pPod, vPod)
	if updated == nil {
		// no update is needed
		return ctrl.Result{}, nil
	}

	err := ctx.VirtualClient.Update(ctx.Context, updated)
	if err == nil {
		ctx.Log.Infof("updated pod %s/%s", vPod.GetNamespace(), vPod.GetName())
	} else {
		err = fmt.Errorf("failed to update pod %s/%s: %v", vPod.GetNamespace(), vPod.GetName(), err)
	}

	return ctrl.Result{}, err
}

func (s *containerResourceSync) SyncToVirtual(ctx *synccontext.SyncContext, event *synccontext.SyncToVirtualEvent[client.Object]) (ctrl.Result, error) {
	// This syncer only works on physical to virtual sync, so we don't need to implement this
	return ctrl.Result{}, nil
}

type monotonicBool struct {
	modified bool
}

func (s *containerResourceSync) updateContainerResources(pObj, vObj *corev1.Pod) *corev1.Pod {
	updated := vObj.DeepCopy()
	if updated.Annotations == nil {
		updated.Annotations = map[string]string{}
	}

	b := &monotonicBool{}
	for i, c := range pObj.Spec.Containers {
		limits := vObj.Spec.Containers[i].Resources.Limits
		cpu := fmt.Sprintf("limits.cpu.%s", c.Name)
		memory := fmt.Sprintf("limits.memory.%s", c.Name)
		storage := fmt.Sprintf("limits.storage.%s", c.Name)
		ephemeralStorage := fmt.Sprintf("limits.ephemeral-storage.%s", c.Name)

		if limits == nil || limits.Cpu() == nil || limits.Cpu().IsZero() {
			updateMap(updated.Annotations, cpu, c.Resources.Limits.Cpu().String(), b)
		}
		if limits == nil || limits.Memory() == nil || limits.Memory().IsZero() {
			updateMap(updated.Annotations, memory, c.Resources.Limits.Memory().String(), b)
		}
		if limits == nil || limits.Storage() == nil || limits.Storage().IsZero() {
			updateMap(updated.Annotations, storage, c.Resources.Limits.Storage().String(), b)
		}
		if limits == nil || limits.StorageEphemeral() == nil || limits.StorageEphemeral().IsZero() {
			updateMap(updated.Annotations, ephemeralStorage, c.Resources.Limits.StorageEphemeral().String(), b)
		}

		requests := vObj.Spec.Containers[i].Resources.Requests
		cpu = fmt.Sprintf("requests.cpu.%s", c.Name)
		memory = fmt.Sprintf("requests.memory.%s", c.Name)
		storage = fmt.Sprintf("requests.storage.%s", c.Name)
		ephemeralStorage = fmt.Sprintf("requests.ephemeral-storage.%s", c.Name)

		if requests == nil || requests.Cpu() == nil || requests.Cpu().IsZero() {
			updateMap(updated.Annotations, cpu, c.Resources.Requests.Cpu().String(), b)
		}
		if requests == nil || requests.Memory() == nil || requests.Memory().IsZero() {
			updateMap(updated.Annotations, memory, c.Resources.Requests.Memory().String(), b)
		}
		if requests == nil || requests.Storage() == nil || requests.Storage().IsZero() {
			updateMap(updated.Annotations, storage, c.Resources.Requests.Storage().String(), b)
		}
		if requests == nil || requests.StorageEphemeral() == nil || requests.StorageEphemeral().IsZero() {
			updateMap(updated.Annotations, ephemeralStorage, c.Resources.Requests.StorageEphemeral().String(), b)
		}
	}

	if !b.modified {
		return nil
	}
	return updated
}

func updateMap(strMap map[string]string, key, value string, b *monotonicBool) {
	_, found := strMap[key]
	b.modified = !found || b.modified
	strMap[toValidDnsName(key)] = value
}

func toValidDnsName(v string) string {
	if len(v) > 63 {
		v = v[:63]
	}
	return v
}
