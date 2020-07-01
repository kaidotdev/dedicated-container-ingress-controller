package processor

import (
	"context"
	ingressV1 "dedicated-container-ingress-controller/api/v1"
	"dedicated-container-ingress-controller/pkg/server/core"

	"golang.org/x/xerrors"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"

	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const myFinalizerName = "ingress.finalizers.kaidotdev.github.io"

var (
	scheme = runtime.NewScheme() // nolint:gochecknoglobals
)

func init() { // nolint:gochecknoinits
	_ = clientgoscheme.AddToScheme(scheme)

	_ = ingressV1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type ControllerSettings struct {
	DedicatedContainerFactory *core.DedicatedContainerFactory
	Logger                    ILogger
}

type Controller struct {
	mgr    ctrl.Manager
	stopCh chan struct{}
}

func NewController(settings ControllerSettings) (*Controller, error) {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:         scheme,
		LeaderElection: false,
	})
	if err != nil {
		return nil, xerrors.Errorf("failed to create manager: %w", err)
	}
	if err := (&reconciler{
		Client:                    mgr.GetClient(),
		Scheme:                    mgr.GetScheme(),
		Recorder:                  mgr.GetEventRecorderFor("dedicated-container-ingress-controller"),
		dedicatedContainerFactory: settings.DedicatedContainerFactory,
		logger:                    settings.Logger,
	}).setupWithManager(mgr); err != nil {
		return nil, xerrors.Errorf("failed to setup manager: %w", err)
	}
	return &Controller{
		mgr:    mgr,
		stopCh: make(chan struct{}),
	}, nil
}

func (c *Controller) Start() error {
	return c.mgr.Start(c.stopCh)
}

func (c *Controller) Stop(_ context.Context) error {
	close(c.stopCh)
	return nil
}

type reconciler struct {
	client.Client
	Scheme                    *runtime.Scheme
	Recorder                  record.EventRecorder
	dedicatedContainerFactory *core.DedicatedContainerFactory
	logger                    ILogger
}

func (r *reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ingress := &ingressV1.DedicatedContainerIngress{}
	ctx := context.Background()
	if err := r.Client.Get(ctx, req.NamespacedName, ingress); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	host := ingress.Spec.Host
	if ingress.ObjectMeta.DeletionTimestamp.IsZero() {
		if !r.dedicatedContainerFactory.HasEntry(host) {
			podTemplateSpec := ingress.Spec.Template
			podTemplateSpec.Namespace = ingress.Namespace
			r.dedicatedContainerFactory.AddEntry(host, podTemplateSpec)
			r.Recorder.Eventf(ingress, coreV1.EventTypeNormal, "SuccessfulCreated", "Created entry: %q", host)
		}
	} else if containsFinalizerString(ingress.ObjectMeta.Finalizers) {
		r.dedicatedContainerFactory.DeleteEntry(host)
		r.Recorder.Eventf(ingress, coreV1.EventTypeNormal, "SuccessfulDeleted", "Deleted entry: %q", host)

		ingress.ObjectMeta.Finalizers = removeFinalizerString(ingress.ObjectMeta.Finalizers)
		if err := r.Client.Update(ctx, ingress); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *reconciler) setupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ingressV1.DedicatedContainerIngress{}).
		Complete(r)
}

func containsFinalizerString(slice []string) bool {
	for _, item := range slice {
		if item == myFinalizerName {
			return true
		}
	}
	return false
}

func removeFinalizerString(slice []string) (result []string) {
	for _, item := range slice {
		if item == myFinalizerName {
			continue
		}
		result = append(result, item)
	}
	return
}
