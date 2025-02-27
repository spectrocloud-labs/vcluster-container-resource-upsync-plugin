package syncers

import (
	"fmt"
	synccontext "github.com/loft-sh/vcluster/pkg/controllers/syncer/context"
	"github.com/loft-sh/vcluster/pkg/controllers/syncer/translator"
	"github.com/loft-sh/vcluster/pkg/types"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewContainerResourceSyncer(ctx *synccontext.RegisterContext) types.Base {
	return &containerResourceSyncer{
		NamespacedTranslator: translator.NewNamespacedTranslator(ctx, "pod", &corev1.Pod{}),
	}
}

type containerResourceSyncer struct {
	// implicitly uses default PhysicalToVirtual & VirtualToPhysical implementations
	translator.NamespacedTranslator
}

func (s *containerResourceSyncer) Name() string {
	return "container-resource-syncer"
}

func (s *containerResourceSyncer) Resource() client.Object {
	return &corev1.Pod{}
}

var _ types.Starter = &containerResourceSyncer{}

func (s *containerResourceSyncer) ReconcileStart(ctx *synccontext.SyncContext, req ctrl.Request) (bool, error) {
	return false, nil
}

func (s *containerResourceSyncer) ReconcileEnd() {
	// NOOP
}

func (s *containerResourceSyncer) Sync(ctx *synccontext.SyncContext, pObj client.Object, vObj client.Object) (ctrl.Result, error) {
	pPod := pObj.(*corev1.Pod)
	vPod := vObj.(*corev1.Pod)

	updated := s.updateContainerResources(pPod, vPod)
	if updated == nil {
		// no update is needed
		return ctrl.Result{}, nil
	}

	err := ctx.VirtualClient.Update(ctx.Context, updated)
	if err == nil {
		ctx.Log.Infof("updated pod %s/%s", vObj.GetNamespace(), vObj.GetName())
	} else {
		err = fmt.Errorf("failed to update pod %s/%s: %v", vObj.GetNamespace(), vObj.GetName(), err)
	}

	return ctrl.Result{}, err
}

func (s *containerResourceSyncer) SyncDown(ctx *synccontext.SyncContext, vObj client.Object) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsManaged determines if a physical object is managed by the vcluster
func (s *containerResourceSyncer) IsManaged(pObj client.Object) (bool, error) {
	return false, nil
}

type monotonicBool struct {
	modified bool
}

func (s *containerResourceSyncer) updateContainerResources(pObj, vObj *corev1.Pod) *corev1.Pod {
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
