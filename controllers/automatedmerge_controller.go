/*
Copyright 2022 Amshuman K R <amshuman.kr@gmail.com>.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/trishanku/heaven/api/v1alpha1"
)

// AutomatedMergeReconciler reconciles a AutomatedMerge object
type AutomatedMergeReconciler struct {
	gitcdReconciler
}

func NewAutomatedMergeReconciler(
	client client.Client,
	scheme *runtime.Scheme,
	defaultGitImage,
	defaultGitcdImage string,
) *AutomatedMergeReconciler {
	return &AutomatedMergeReconciler{
		gitcdReconciler: &gitcdReconcilerImpl{
			Client:            client,
			scheme:            scheme,
			defaultGitImage:   defaultGitImage,
			defaultGitcdImage: defaultGitcdImage,
		},
	}
}

func (c configNames) initForAutomatedMerge(am *v1alpha1.AutomatedMerge) {
	type configSpec struct {
		name       string
		configName *string
	}

	var configs = []configSpec{
		{name: CONFIG_ENTRYPOINTS, configName: am.Spec.App.EntrypointsConfigMapName},
		{name: DEPLOY_AUTOMERGE},
	}

	for _, s := range configs {
		func(s configSpec) {
			if _, ok := c[s.name]; !ok {
				if s.configName != nil {
					c[s.name] = *s.configName
				} else {
					c[s.name] = am.Name + "-" + s.name
				}
			}
		}(s)
	}
}

//+kubebuilder:rbac:groups=controllers.trishanku.org.trishanku.org,resources=automatedmerges,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=controllers.trishanku.org.trishanku.org,resources=automatedmerges/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=controllers.trishanku.org.trishanku.org,resources=automatedmerges/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AutomatedMerge object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *AutomatedMergeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	var (
		l         = log.FromContext(ctx)
		am        = &v1alpha1.AutomatedMerge{}
		c         = make(configNames)
		ensureFns []ensureFunc
	)

	l.Info("Reconciling", "key", req.String())
	defer func() {
		l.Info("Reconciled", "key", req.String(), "res", res, "err", err)
	}()

	if err = r.Get(ctx, req.NamespacedName, am); err != nil {
		if apierrors.IsNotFound(err) {
			res.Requeue = false
		}
		return
	}

	if am.DeletionTimestamp != nil {
		// Rely on owner references and cascade deletion.
		return
	}

	// TODO check if controller and owner references are handled by controller-runtime.

	c.initForAutomatedMerge(am)

	l.Info("configNames", "configNames", c)

	ensureFns = append(ensureFns, r.ensureEntrypointsConfigMap)

	ensureFns = append(ensureFns, r.ensureDeployment(am))

	for _, ensureFn := range ensureFns {
		if err = ensureFn(ctx, am, c); err != nil {
			res.Requeue = true
			return
		}
	}

	return
}

func (r *AutomatedMergeReconciler) generateDeploymentFor(ctx context.Context, am *v1alpha1.AutomatedMerge, c configNames) (d *appsv1.Deployment, err error) {
	var podSpec *corev1.PodSpec

	d = r.newDeployment(
		am,
		am.Spec.App.Replicas,
		&v1alpha1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: am.Annotations,
				Labels:      am.Labels,
			},
		},
		c,
		DEPLOY_AUTOMERGE,
	)

	podSpec = &d.Spec.Template.Spec

	r.appendVolumesForGitcd(podSpec, c)

	r.appendImagePullSecrets(podSpec, &am.Spec.Gitcd)

	r.appendGitcdContainers(podSpec, &am.Spec.Gitcd, GITCD_SUBCOMMAND_PULL, am.Spec.App.CommitterName)

	return
}

func (r *AutomatedMergeReconciler) ensureDeployment(am *v1alpha1.AutomatedMerge) ensureFunc {
	return func(ctx context.Context, owner metav1.Object, c configNames) (err error) {
		var rd *appsv1.Deployment

		if rd, err = r.generateDeploymentFor(ctx, am, c); err != nil {
			return
		}

		err = r.createOrUpdateDeployment(ctx, owner, rd)
		return
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *AutomatedMergeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("automatedmerge").
		For(&v1alpha1.AutomatedMerge{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
