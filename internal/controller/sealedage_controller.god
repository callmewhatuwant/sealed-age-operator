// internal/controller/sealedage_controller.go
package controller

import (
    "context"
    "errors"
    "fmt"
    "io"
    "strings"
    "time"

    age "filippo.io/age"
    "filippo.io/age/armor"

    corev1 "k8s.io/api/core/v1"
    apierrors "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/types"

    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
    "sigs.k8s.io/controller-runtime/pkg/log"

    securityv1alpha1 "github.com/yournamecallmewhatuwant/sealed-age-operator/api/v1alpha1"
)

// SealedAgeReconciler reconciles SealedAge resources.
type SealedAgeReconciler struct {
    client.Client
    Scheme *runtime.Scheme

    // Configurable via CLI flags (see cmd/main.go)
    KeyNamespace string // default: "sealed-age-system"
    KeyLabelKey  string // default: "app"
    KeyLabelVal  string // default: "age-key"
}

// +kubebuilder:rbac:groups=security.age.io,resources=sealedages,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=security.age.io,resources=sealedages/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=security.age.io,resources=sealedages/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch

func (r *SealedAgeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx).WithValues("sealedage", req.NamespacedName)

    // 1. Load the SealedAge resource.
    var cr securityv1alpha1.SealedAge
    if err := r.Get(ctx, req.NamespacedName, &cr); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // 2. List available AGE key Secrets (by namespace and label).
    keyList := &corev1.SecretList{}
    if err := r.List(ctx, keyList,
        client.InNamespace(r.KeyNamespace),
        client.MatchingLabels{r.KeyLabelKey: r.KeyLabelVal},
    ); err != nil {
        logger.Error(err, "failed to list key secrets", "namespace", r.KeyNamespace)
        return ctrl.Result{}, err
    }
    if len(keyList.Items) == 0 {
        logger.Info("no AGE keys found, will retry", "namespace", r.KeyNamespace)
        return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
    }

    // 3. Decrypt each field in spec.encryptedData.
    plain := map[string][]byte{}
    for field, enc := range cr.Spec.EncryptedData {
        b, keyUsed, derr := decryptWithAge(ctx, enc, keyList.Items, cr.Spec.Recipients)
        if derr != nil {
            logger.Error(derr, "failed to decrypt", "field", field)
            return ctrl.Result{}, fmt.Errorf("decrypt %s: %w",

