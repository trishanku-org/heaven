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
	"fmt"

	"golang.org/x/oauth2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v50/github"
	v1alpha1 "github.com/trishanku/heaven/api/v1alpha1"
)

// GitHubRepositoryReconciler reconciles a GitHubRepository object
type GitHubRepositoryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=controllers.trishanku.org.trishanku.org,resources=githubrepositories,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=controllers.trishanku.org.trishanku.org,resources=githubrepositories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=controllers.trishanku.org.trishanku.org,resources=githubrepositories/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GitHubRepository object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *GitHubRepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	var (
		l    = log.FromContext(ctx)
		repo = &v1alpha1.GitHubRepository{}
		c    *github.Client
	)

	l.Info("Reconciling", "key", req.String())
	defer func() {
		l.Info("Reconciled", "key", req.String(), "res", res, "err", err)
	}()

	if err = r.Get(ctx, req.NamespacedName, repo); err != nil {
		if apierrors.IsNotFound(err) {
			res.Requeue = false
		}
		return
	}

	if repo.DeletionTimestamp != nil {
		// TODO safely delete GitHub repos.
		return
	}

	// TODO check if controller and owner references are handled by controller-runtime.

	if c, err = r.getGitHubClientFor(ctx, repo); err != nil {
		return
	}

	err = r.createRepoIfNotExists(ctx, c, repo, l)

	return
}

func (r *GitHubRepositoryReconciler) getGitHubClientFor(ctx context.Context, repo *v1alpha1.GitHubRepository) (c *github.Client, err error) {
	var (
		s = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      repo.Spec.CredentialsSecretName,
				Namespace: repo.Namespace,
			},
		}

		token []byte
		ok    bool
	)

	if err = r.Get(ctx, client.ObjectKeyFromObject(s), s); err != nil {
		return
	}

	if token, ok = s.Data[repo.Spec.CredentialsSecretTokenKey]; !ok {
		err = fmt.Errorf("key %q not found in the credentials", repo.Spec.CredentialsSecretTokenKey)
		return
	}

	c = github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(token)})))

	return
}

func (r *GitHubRepositoryReconciler) createRepoIfNotExists(
	ctx context.Context,
	c *github.Client,
	repo *v1alpha1.GitHubRepository,
	l logr.Logger) (err error) {
	var (
		user  *github.User
		owner string
	)

	if user, _, err = c.Users.Get(ctx, ""); err != nil {
		return
	}

	owner = getRepoOwner(repo, *user.Login)

	if _, _, err = c.Repositories.Get(ctx, owner, repo.Name); err == nil {
		l.Info("GitHub Repository already exists", "owner", owner, "repo", repo)
		return
	} else {
		// TODO exit for all errors except not found error.
		l.Error(err, "Error getting project", "owner", owner, "repo", repo.Name)
	}

	_, _, err = c.Repositories.Create(ctx, owner, &github.Repository{
		Name:        &repo.Name,
		Description: &repo.Spec.Description,
		Private:     &repo.Spec.Private,
	})

	return
}

func getRepoOwner(repo *v1alpha1.GitHubRepository, authUser string) string {
	if len(repo.Spec.Organization) > 0 {
		return repo.Spec.Organization
	}

	return authUser
}

// SetupWithManager sets up the controller with the Manager.
func (r *GitHubRepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("githubrepository").
		For(&v1alpha1.GitHubRepository{}).
		Complete(r)
}
