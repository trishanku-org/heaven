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
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/trishanku/heaven/api/v1alpha1"
	"github.com/trishanku/heaven/pkg/certs"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientcmdv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

// gitcdReconciler reconciles a deployment running Gitcd.
type gitcdReconciler interface {
	client.Client
	getScheme() *runtime.Scheme
	getDefaultGitImage() string
	getDefaultGitcdImage() string
	ensureEntrypointsConfigMap(context.Context, metav1.Object, configNames) error
	createOrUpdateDeployment(context.Context, metav1.Object, *appsv1.Deployment) error
	newDeployment(metav1.Object, *int32, *v1alpha1.PodTemplateSpec, configNames, string) *appsv1.Deployment
	appendVolumesForGitcd(*corev1.PodSpec, configNames)
	appendImagePullSecrets(*corev1.PodSpec, *v1alpha1.GitcdSpec)
	appendGitcdContainers(*corev1.PodSpec, *v1alpha1.GitcdSpec, string)
	appendGitPostInitContainer(*corev1.PodSpec, *v1alpha1.GitcdSpec)
}

// gitcdReconcilerImpl implements gitcdReconciler.
type gitcdReconcilerImpl struct {
	client.Client
	scheme            *runtime.Scheme
	defaultGitImage   string
	defaultGitcdImage string
}

var _ gitcdReconciler = &gitcdReconcilerImpl{}

func (r *gitcdReconcilerImpl) getScheme() *runtime.Scheme {
	return r.scheme
}

func (r *gitcdReconcilerImpl) getDefaultGitImage() string {
	return r.defaultGitImage
}

func (r *gitcdReconcilerImpl) getDefaultGitcdImage() string {
	return r.defaultGitcdImage
}

// TrishankuHeavenReconciler reconciles a TrishankuHeaven object
type TrishankuHeavenReconciler struct {
	gitcdReconciler

	DefaultEtcdImage      string
	DefaultApiserverImage string
}

func NewTrishankuHeavenReconciler(
	client client.Client,
	scheme *runtime.Scheme,
	defaultEtcdImage,
	defaultGitImage,
	defaultGitcdImage,
	defaultApiserverImage string,
) *TrishankuHeavenReconciler {
	return &TrishankuHeavenReconciler{
		gitcdReconciler: &gitcdReconcilerImpl{
			Client:            client,
			scheme:            scheme,
			defaultGitImage:   defaultGitImage,
			defaultGitcdImage: defaultGitcdImage,
		},
		DefaultEtcdImage:      defaultEtcdImage,
		DefaultApiserverImage: defaultApiserverImage,
	}
}

//+kubebuilder:rbac:groups=controllers.trishanku.org.trishanku.org,resources=trishankuheavens,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=controllers.trishanku.org.trishanku.org,resources=trishankuheavens/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=controllers.trishanku.org.trishanku.org,resources=trishankuheavens/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *TrishankuHeavenReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	var (
		l         = log.FromContext(ctx)
		heaven    = &v1alpha1.TrishankuHeaven{}
		c         = make(configNames)
		ensureFns []ensureFunc
	)

	l.Info("Reconciling", "key", req.String())
	defer func() {
		l.Info("Reconciled", "key", req.String(), "res", res, "err", err)
	}()

	if err = r.Get(ctx, req.NamespacedName, heaven); err != nil {
		if apierrors.IsNotFound(err) {
			res.Requeue = false
		}
		return
	}

	if heaven.DeletionTimestamp != nil {
		// Rely on owner references and cascade deletion.
		return
	}

	// TODO check if controller and owner references are handled by controller-runtime.

	c.initForHeaven(heaven)

	l.Info("configNames", "configNames", c)

	if !heaven.Spec.Skip.Cerfificates {
		ensureFns = append(ensureFns, r.ensureCertificateSecrets(heaven)...)
	}

	if !heaven.Spec.Skip.Entrypoints {
		ensureFns = append(ensureFns, r.ensureEntrypointsConfigMap)
	}

	ensureFns = append(ensureFns, r.ensureDeployment(heaven))

	for _, ensureFn := range ensureFns {
		if err = ensureFn(ctx, heaven, c); err != nil {
			res.Requeue = true
			return
		}
	}

	return
}

const (
	KEY_SECRET_KUBECONFIG = "kubeconfig"
	KEY_SECRET_URL        = "url"

	SECRET_CERT_CA               = "cert-ca"
	SECRET_CERT_KUBERNETES       = "cert-kubernetes"
	SECRET_CERT_SERVICE_ACCOUNTS = "cert-service-account"
	SECRET_KUBECONFIG_ADMIN      = "admin"
	SECRET_KUBECONFIG_CONTROLLER = "controller"
	CONFIG_ENTRYPOINTS           = "entrypoints"
	DEPLOY_HEAVEN                = "heaven"
	DEPLOY_AUTOMERGE             = "automerge"
	VOLUME_HOME                  = "home"
	VOLUME_HOME_PR               = "home-pr"

	CONTAINER_ETCD       = "events-etcd"
	CONTAINER_GIT_PRE    = "git-pre"
	CONTAINER_GITCD_INIT = "gitcd-init"
	CONTAINER_GIT_POST   = "git-post"
	CONTAINER_GIT_PRE_PR = "git-pre-pr"
	CONTAINER_GITCD      = "gitcd"
	CONTAINER_GITCD_PR   = "gitcd-pr"
	CONTAINER_APISERVER  = "apiserver"

	ENTRYPOINT_GIT_PRE = "git-pre.sh"

	BASE_PATH_SECRETS     = "/.trishanku/secrets"
	BASE_PATH_ENTRYPOINTS = "/.trishanku/entrypoints"
	BASE_PATH_HOME        = "/root"
	BASE_PATH_TEMP        = "/.trishanku/temp"

	TRISHANKU_OU           = "Heaven"
	TRISHANKU_CONTEXT_NAME = "default"
	TRISHANKU_CLUSTER_NAME = "trishanku"
)

type configNames map[string]string

func (c configNames) initForHeaven(heaven *v1alpha1.TrishankuHeaven) {
	type configSpec struct {
		name       string
		configName *string
	}

	var (
		configs = []configSpec{
			{name: SECRET_CERT_CA, configName: heaven.Spec.Certificates.CertificateAuthoritySecretName},
			{name: SECRET_CERT_SERVICE_ACCOUNTS, configName: heaven.Spec.Certificates.ServiceAccountsSecretName},
			{name: SECRET_CERT_KUBERNETES, configName: heaven.Spec.Certificates.KubernetesSecretName},
			{name: SECRET_KUBECONFIG_ADMIN, configName: heaven.Spec.Certificates.AdminSecretName},
			{name: DEPLOY_HEAVEN},
		}

		entryPointsConfigMapName *string = nil
	)

	if heaven.Spec.App != nil {
		entryPointsConfigMapName = heaven.Spec.App.EntrypointsConfigMapName
	}

	configs = append(configs, configSpec{name: CONFIG_ENTRYPOINTS, configName: entryPointsConfigMapName})

	if heaven.Spec.Certificates.Controller != nil {
		configs = append(configs, configSpec{
			name:       SECRET_KUBECONFIG_CONTROLLER,
			configName: heaven.Spec.Certificates.Controller.SecretName,
		})
	}

	for _, s := range configs {
		func(s configSpec) {
			if _, ok := c[s.name]; !ok {
				if s.configName != nil {
					c[s.name] = *s.configName
				} else {
					c[s.name] = heaven.Name + "-" + s.name
				}
			}
		}(s)
	}
}

func (c configNames) getConfigurationName(config string) string {
	return c[config]
}

type ensureFunc func(context.Context, metav1.Object, configNames) error

func (r *TrishankuHeavenReconciler) ensureCertificateSecrets(heaven *v1alpha1.TrishankuHeaven) (fns []ensureFunc) {
	type certSpec struct {
		name, cn, o, ou    string
		dnsNames           []string
		isCA               bool
		generateKubeconfig bool
		secretName         *string
	}

	var (
		ca        *certs.Certificate
		certSpecs = []certSpec{
			{
				name:       SECRET_CERT_CA,
				secretName: heaven.Spec.Certificates.CertificateAuthoritySecretName,
				cn:         "CA",
				o:          "Trishanku",
				ou:         "CA",
				isCA:       true,
			},
			{
				name:       SECRET_CERT_SERVICE_ACCOUNTS,
				secretName: heaven.Spec.Certificates.ServiceAccountsSecretName,
				cn:         "service-accounts",
				o:          "Trishanku",
				ou:         TRISHANKU_OU,
			},
			{
				name:               SECRET_KUBECONFIG_ADMIN,
				cn:                 "admin",
				o:                  "system:masters",
				ou:                 TRISHANKU_OU,
				generateKubeconfig: true,
				secretName:         heaven.Spec.Certificates.AdminSecretName,
			},
			{
				name: SECRET_CERT_KUBERNETES,
				cn:   "kubernetes",
				o:    "Trishanku",
				ou:   TRISHANKU_OU,
				dnsNames: []string{
					"localhost",
					"kubernetes",
					"kubernetes.default",
					"kubernetes.default.svc",
					"kubernetes.default.svc.cluster",
					"kubernetes.svc.cluster.local",
				},
				secretName: heaven.Spec.Certificates.KubernetesSecretName,
			},
		}
	)

	if heaven.Spec.Certificates.Controller != nil {
		certSpecs = append(certSpecs, certSpec{
			name:               SECRET_KUBECONFIG_CONTROLLER,
			cn:                 heaven.Spec.Certificates.Controller.UserName,
			o:                  heaven.Spec.Certificates.Controller.GroupName,
			ou:                 TRISHANKU_OU,
			generateKubeconfig: true,
			secretName:         heaven.Spec.Certificates.Controller.SecretName,
		})
	}

	for _, cs := range certSpecs {
		func(cs certSpec) {
			fns = append(fns, func(ctx context.Context, owner metav1.Object, c configNames) (err error) {
				var (
					secretName = c.getConfigurationName(cs.name)
					namespace  = owner.GetNamespace()

					s = &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      secretName,
							Namespace: namespace,
						},
					}

					cert *certs.Certificate
				)

				if err = r.Get(ctx, client.ObjectKeyFromObject(s), s); err == nil {
					// The certificate secret exists re-use it.
					if cs.isCA {
						ca, err = certs.LoadCertificateFromMap(s.Data)
					}

					return

				} else if !apierrors.IsNotFound(err) {
					return // Fail for any error other than not found error.
				}

				defer func() {
					if cs.isCA && err == nil {
						// Save generated CA certificat for later use in generating other certificates.
						ca = cert
					}
				}()

				// Create a fresh certificate.

				s.Name = secretName
				s.Namespace = namespace

				s.Type = corev1.SecretTypeTLS

				if cert, err = (&certs.CertificateConfig{
					CommonName:         cs.cn,
					Organization:       []string{cs.o},
					OrganizationalUnit: []string{cs.ou},
					DNSNames:           cs.dnsNames,
					CA:                 ca,
					IsCA:               cs.isCA,
				}).Generate(); err != nil {
					return
				}

				if s.Data == nil {
					s.Data = make(map[string][]byte)
				}

				cert.WriteToMap(s.Data)

				if cs.generateKubeconfig {
					if err = certs.GenerateKubeconfigAndWriteToMap(
						TRISHANKU_CONTEXT_NAME,
						TRISHANKU_CLUSTER_NAME,
						cs.cn,
						clientcmdv1.Cluster{
							Server:                   "https://localhost:6443",
							CertificateAuthorityData: cert.CA.CertificatePEM,
						},
						clientcmdv1.AuthInfo{
							ClientCertificateData: cert.CertificatePEM,
							ClientKeyData:         cert.PrivateKeyPEM,
						},
						s.Data,
					); err != nil {
						return
					}
				}

				// TODO set owner references for shared secrets.
				if err = controllerutil.SetControllerReference(owner, s, r.getScheme()); err != nil {
					return
				}

				if err = r.Create(ctx, s); err != nil {
					return
				}

				return
			})
		}(cs)
	}

	return
}

func (r *gitcdReconcilerImpl) ensureEntrypointsConfigMap(ctx context.Context, owner metav1.Object, c configNames) (err error) {
	var cm = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.getConfigurationName(CONFIG_ENTRYPOINTS),
			Namespace: owner.GetNamespace(),
		},
	}

	_, err = ctrl.CreateOrUpdate(ctx, r, cm, func() error {
		cm.Name = c.getConfigurationName(CONFIG_ENTRYPOINTS)
		cm.Namespace = owner.GetNamespace()

		if cm.Data == nil {
			cm.Data = make(map[string]string)
		}

		cm.Data[ENTRYPOINT_GIT_PRE] = `#!/bin/bash
if [ ! -f "${HOME}/.git-credentials" ]; then
	for file in "${HOME}/.gitconfig" "${HOME}/.git-credentials"; do
		touch "$file"
	done

	chmod 600 "${HOME}/.git-credentials"

	git config --global credential.helper store

	cat <<CRED_SPEC_EOF | git credential approve
url=$GIT_CRED_URL
username=$GIT_CRED_USERNAME
password=$GIT_CRED_PASSWORD
CRED_SPEC_EOF

	git config --global user.name "$GITCD_COMMITTER_NAME"
	git config --global user.email "trishanku@heaven.com"
fi

REPO="${HOME}/repo"

function set_head {
	local BRANCH="$1"

	echo "Setting branch ${BRANCH} as the HEAD."
	git symbolic-ref HEAD "refs/heads/$BRANCH"
}

function init_data_branch {
	local BRANCH="$1"

	# Create an empty tree as the base commit if the branch is not pointing to a valid commit.
	if git cat-file -e "$BRANCH" 2> /dev/null; then
	 	echo "Reusing existing branch ${BRANCH}."
		set_head "$BRANCH"
	else
		echo "Creating an initial commit for the branch ${BRANCH}."
		git update-ref "refs/heads/$BRANCH" $(git commit-tree -m "init" $(git write-tree)) \
			&& set_head "$BRANCH"
	fi
}

if [ ! -d "$REPO" ]; then
	if [ "$GITCD_REMOTE_REPO" == "" ]; then
		git init --bare --initial-branch "$GITCD_BRANCH_DATA" "$REPO" || exit 1
	else
		git clone --bare "$GITCD_REMOTE_REPO" "$REPO" || exit 1
	fi
fi

cd "$REPO" || exit 1

if [ "$GITCD_REMOTE_REPO" == "" ]; then
	# Local repo.
	init_data_branch "$GITCD_BRANCH_DATA"

	exit 0
fi

# Remote repo.
git show-branch "$GITCD_REMOTE_BRANCH_DATA" /dev/null 2>&1 || [ "$GITCD_CREATE_REMOTE_BRANCH" == "true" ] || git fetch || exit 1

git config core.logAllRefUpdates always
git config --unset-all remote.origin.fetch
git config --unset-all remote.origin.push

function prepare_branch {
	local BRANCH="$1"
	local REMOTE_BRANCH="$2"
	
	git config --add remote.origin.fetch "refs/heads/${REMOTE_BRANCH}:refs/heads/${REMOTE_BRANCH}"
	git config --add remote.origin.push "refs/heads/${BRANCH}:refs/heads/${BRANCH}"
 
	if  git show-branch "$BRANCH" > /dev/null 2>&1; then
	  echo "Branch ${BRANCH} already exists."
	elif [ "$GITCD_CREATE_LOCAL_BRANCH" == "true" ]; then
		if git show-branch "$REMOTE_BRANCH" > /dev/null 2>&1; then
			echo "Creating branch ${BRANCH} from ${REMOTE_BRANCH}."
			git branch "${BRANCH}" "${REMOTE_BRANCH}" --no-track
		elif [ "$GITCD_CREATE_REMOTE_BRANCH" == "true" ]; then
			echo "Remote branch ${REMOTE_BRANCH} does not exist. Creating a fresh branch ${BRANCH}."
		else
			echo "Exiting because ${REMOTE_BRANCH} does not exist."
			exit 1
		fi
	else
		echo "Exiting because ${BRANCH} does not exist."
		exit 1
	fi
}

prepare_branch "$GITCD_BRANCH_DATA" "$GITCD_REMOTE_BRANCH_DATA"
prepare_branch "$GITCD_BRANCH_METADATA" "$GITCD_REMOTE_BRANCH_METADATA"

init_data_branch "$GITCD_BRANCH_DATA"

if [ "$GITCD_PUSH_AFTER_INIT" != "" ]; then
	git push
fi
`
		if existing := metav1.GetControllerOf(cm); existing == nil {
			if err := controllerutil.SetControllerReference(owner, cm, r.getScheme()); err != nil {
				return err
			}
		}

		return nil
	})

	return
}

func (r *gitcdReconcilerImpl) createOrUpdateDeployment(ctx context.Context, owner metav1.Object, rd *appsv1.Deployment) (err error) {
	var d = rd.DeepCopy()

	_, err = ctrl.CreateOrUpdate(ctx, r, d, func() (err error) {
		// Ensure annotations and labels.
		for _, s := range []struct {
			src map[string]string
			dst *map[string]string
		}{
			{dst: &d.Labels, src: rd.Labels},
			{dst: &d.Annotations, src: rd.Annotations},
		} {
			if *s.dst == nil || len(*s.dst) <= 0 {
				*s.dst = s.src
				continue
			}
			for k, v := range s.src {
				(*s.dst)[k] = v
			}
		}

		d.Spec = rd.Spec

		if existing := metav1.GetControllerOf(d); existing == nil {
			err = controllerutil.SetControllerReference(owner, d, r.getScheme())
		}

		return
	})

	return
}

func getImage(spec *v1alpha1.ImageSpec, defaultImage string) string {
	if spec != nil && spec.Image != nil && len(*spec.Image) > 0 {
		return *spec.Image
	}

	return defaultImage
}

func getImagePullPolicy(spec *v1alpha1.ImageSpec) corev1.PullPolicy {
	if spec != nil && spec.ImagePullPolicy != nil {
		return *spec.ImagePullPolicy
	}

	return ""
}

func getCreateBranchEnvVar(name string, create bool) corev1.EnvVar {
	return corev1.EnvVar{Name: name, Value: strconv.FormatBool(create)}
}

func getRemoteBranchData(remote *v1alpha1.RemoteSpec) string {
	if remote != nil && remote.Branches != nil {
		return remote.Branches.Data
	}

	return ""
}

func getRemoteBranchMetadata(remote *v1alpha1.RemoteSpec) string {
	if remote != nil && remote.Branches != nil {
		return remote.Branches.Metadata
	}

	return ""
}

func getRemoteNames(remotes []v1alpha1.RemoteSpec) string {
	var ss []string

	for _, r := range remotes {
		ss = append(ss, r.Name)
	}

	return strings.Join(ss, ":")
}

func getRemoteDataReferenceNames(remotes []v1alpha1.RemoteSpec) string {
	var ss []string

	for _, r := range remotes {
		ss = append(ss, "refs/heads/"+r.Branches.Data)
	}

	return strings.Join(ss, ":")
}

func getRemoteMetadataReferenceNames(remotes []v1alpha1.RemoteSpec) string {
	var ss []string

	for _, r := range remotes {
		ss = append(ss, "refs/heads/"+r.Branches.Metadata)
	}

	return strings.Join(ss, ":")
}

func getMergeRetentionPoliciesInclude(remotes []v1alpha1.RemoteSpec) string {
	var ss []string

	for _, r := range remotes {
		ss = append(ss, r.RetentionPolicies.Include)
	}

	return strings.Join(ss, ":")
}

func getMergeRetentionPoliciesExclude(remotes []v1alpha1.RemoteSpec) string {
	var ss []string

	for _, r := range remotes {
		ss = append(ss, r.RetentionPolicies.Exclude)
	}

	return strings.Join(ss, ":")
}

func getMergeConflictResolutions(remotes []v1alpha1.RemoteSpec) string {
	var ss []string

	for _, r := range remotes {
		ss = append(ss, strconv.Itoa(int(r.ConflictResolution)))
	}

	return strings.Join(ss, ":")
}

func appendMergeFlagsToCommand(c *corev1.Container, mergeFlags map[string]string) {
	for f, v := range mergeFlags {
		if len(v) > 0 {
			c.Command = append(c.Command, fmt.Sprintf("%s=%s", f, v))
		}
	}
}

func getCommitterName(heaven *v1alpha1.TrishankuHeaven) string {
	if heaven.Spec.Certificates.Controller != nil {
		return heaven.Spec.Certificates.Controller.UserName
	}

	return ""
}

func getRemoteName(remote *v1alpha1.RemoteSpec) string {
	if remote != nil {
		return remote.Name
	}

	return ""
}

func getRemoteRepo(remote *v1alpha1.RemoteSpec) string {
	if remote != nil {
		return remote.Repo
	}
	return ""
}

func (r *gitcdReconcilerImpl) newDeployment(owner metav1.Object, replicas *int32, podTemplate *v1alpha1.PodTemplateSpec, c configNames, config string) (d *appsv1.Deployment) {
	d = &appsv1.Deployment{
		ObjectMeta: *podTemplate.ObjectMeta.DeepCopy(),
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{MatchLabels: podTemplate.ObjectMeta.Labels},
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: *podTemplate.ObjectMeta.DeepCopy(),
				Spec:       *podTemplate.Spec.DeepCopy(),
			},
		},
	}

	d.Name, d.Namespace = c.getConfigurationName(config), owner.GetNamespace()
	return
}

func (r *gitcdReconcilerImpl) appendVolumesForGitcd(podSpec *corev1.PodSpec, c configNames) {
	podSpec.Volumes = append(
		podSpec.Volumes,
		corev1.Volume{
			Name: CONFIG_ENTRYPOINTS,
			VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: c.getConfigurationName(CONFIG_ENTRYPOINTS)},
				DefaultMode:          pointer.Int32(0755),
			}},
		},
		corev1.Volume{
			Name:         VOLUME_HOME,
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		},
	)
}

func (r *gitcdReconcilerImpl) appendImagePullSecrets(podSpec *corev1.PodSpec, gitcdSpec *v1alpha1.GitcdSpec) {
	if len(gitcdSpec.ImagePullSecretName) > 0 {
		podSpec.ImagePullSecrets = append(podSpec.ImagePullSecrets, corev1.LocalObjectReference{Name: gitcdSpec.ImagePullSecretName})
	}
}

func getGitPreInitContainerName(remote *v1alpha1.RemoteSpec) string {
	if remote != nil && len(remote.Name) > 0 {
		return CONTAINER_GIT_PRE + "-" + remote.Name
	}

	return CONTAINER_GIT_PRE
}

func (r *gitcdReconcilerImpl) appendGitcdContainers(podSpec *corev1.PodSpec, gitcdSpec *v1alpha1.GitcdSpec, committerName string) {
	var (
		gitCredsSecretName string
		remotes            []*v1alpha1.RemoteSpec
		gitcd              corev1.Container
	)

	gitCredsSecretName = gitcdSpec.CredentialsSecretName

	for _, r := range gitcdSpec.Git.Remotes {
		remotes = append(remotes, r.DeepCopy())
	}

	if len(remotes) <= 0 {
		remotes = append(remotes, nil) // We need to init the repo even if there are no remotes.
	}

	for _, remote := range remotes {
		podSpec.InitContainers = append(podSpec.InitContainers, corev1.Container{
			Name:            getGitPreInitContainerName(remote),
			Image:           getImage(gitcdSpec.GitImage, r.getDefaultGitImage()),
			ImagePullPolicy: getImagePullPolicy(gitcdSpec.GitImage),
			Command:         []string{path.Join(BASE_PATH_ENTRYPOINTS, ENTRYPOINT_GIT_PRE)},
			Env: []corev1.EnvVar{
				{
					Name: "GIT_CRED_USERNAME",
					ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: gitCredsSecretName},
						Key:                  corev1.BasicAuthUsernameKey,
					}},
				},
				{
					Name: "GIT_CRED_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: gitCredsSecretName},
						Key:                  corev1.BasicAuthPasswordKey,
					}},
				},
				{
					Name: "GIT_CRED_URL",
					ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: gitCredsSecretName},
						Key:                  KEY_SECRET_URL,
					}},
				},
				{Name: "GITCD_COMMITTER_NAME", Value: committerName},
				{Name: "GITCD_BRANCH_DATA", Value: gitcdSpec.Git.Branches.Data},
				{Name: "GITCD_BRANCH_METADATA", Value: gitcdSpec.Git.Branches.Metadata},
				getCreateBranchEnvVar("GITCD_CREATE_LOCAL_BRANCH", !gitcdSpec.Git.Branches.NoCreateBranch),
				{Name: "GITCD_REMOTE_NAME", Value: getRemoteName(remote)},
				{Name: "GITCD_REMOTE_REPO", Value: getRemoteRepo(remote)},
				{Name: "GITCD_REMOTE_BRANCH_DATA", Value: getRemoteBranchData(remote)},
				{Name: "GITCD_REMOTE_BRANCH_METADATA", Value: getRemoteBranchMetadata(remote)},
				getCreateBranchEnvVar("GITCD_CREATE_REMOTE_BRANCH", remote != nil && !remote.Branches.NoCreateBranch),
			},
			VolumeMounts: []corev1.VolumeMount{
				{Name: CONFIG_ENTRYPOINTS, MountPath: BASE_PATH_ENTRYPOINTS, ReadOnly: true},
				{Name: VOLUME_HOME, MountPath: BASE_PATH_HOME},
			},
		})
	}

	podSpec.InitContainers = append(podSpec.InitContainers, corev1.Container{
		Name:            CONTAINER_GITCD_INIT,
		Image:           getImage(gitcdSpec.GitcdImage, r.getDefaultGitcdImage()),
		ImagePullPolicy: getImagePullPolicy(gitcdSpec.GitcdImage),
		Args: []string{
			"init",
			"--repo=/root/repo",
			"--data-reference-names=default=refs/heads/" + gitcdSpec.Git.Branches.Data,
			"--metadata-reference-names=default=refs/heads/" + gitcdSpec.Git.Branches.Metadata,
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: VOLUME_HOME, MountPath: BASE_PATH_HOME},
		},
	})

	gitcd = corev1.Container{
		Name:            CONTAINER_GITCD,
		Image:           getImage(gitcdSpec.GitcdImage, r.getDefaultGitcdImage()),
		ImagePullPolicy: getImagePullPolicy(gitcdSpec.GitcdImage),
		Command: []string{
			"/gitcd",
			"serve",
			"--repo=/root/repo",
			"--committer-name=" + committerName,
			"--data-reference-names=default=refs/heads/" + gitcdSpec.Git.Branches.Data,
			"--metadata-reference-names=default=refs/heads/" + gitcdSpec.Git.Branches.Metadata,
			"--key-prefixes=default=/registry",
			"--pull-ticker-duration=" + gitcdSpec.Pull.TickerDuration.Duration.String(),
			"--remote-names=default=" + getRemoteNames(gitcdSpec.Git.Remotes),
			"--no-fast-forwards=default=false",
			"--remote-data-reference-names=default=" + getRemoteDataReferenceNames(gitcdSpec.Git.Remotes),
			"--remote-meta-reference-names=default=" + getRemoteMetadataReferenceNames(gitcdSpec.Git.Remotes),
			"--listen-urls=default=http://0.0.0.0:2479/",
			"--advertise-client-urls=default=http://127.0.0.1:2479/",
			"--watch-dispatch-channel-size=1",
			"--push-after-merges=default=" + strconv.FormatBool(gitcdSpec.Pull.PushAfterMerge),
			"--push-on-pull-failures=default=" + strconv.FormatBool(gitcdSpec.Pull.PushOnPullFailure),
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: VOLUME_HOME, MountPath: BASE_PATH_HOME},
		},
	}

	appendMergeFlagsToCommand(
		&gitcd,
		map[string]string{
			"--merge-retention-policies-include": "default=" + getMergeRetentionPoliciesInclude(gitcdSpec.Git.Remotes),
			"--merge-retention-policies-exclude": "default=" + getMergeRetentionPoliciesExclude(gitcdSpec.Git.Remotes),
			"--merge-conflict-resolutions":       "default=" + getMergeConflictResolutions(gitcdSpec.Git.Remotes),
		},
	)

	podSpec.Containers = append(podSpec.Containers, gitcd)
}

func (r *gitcdReconcilerImpl) appendGitPostInitContainer(podSpec *corev1.PodSpec, gitcdSpec *v1alpha1.GitcdSpec) {
	if len(gitcdSpec.Git.Remotes) > 0 {
		podSpec.InitContainers = append(podSpec.InitContainers, corev1.Container{
			Name:            CONTAINER_GIT_POST,
			Image:           getImage(gitcdSpec.GitImage, r.getDefaultGitImage()),
			ImagePullPolicy: getImagePullPolicy(gitcdSpec.GitImage),
			Command: []string{
				"git",
				"push",
			},
			VolumeMounts: []corev1.VolumeMount{
				{Name: VOLUME_HOME, MountPath: BASE_PATH_HOME},
			},
			WorkingDir: "/root/repo",
		})
	}
}

func (r *TrishankuHeavenReconciler) generateDeploymentFor(ctx context.Context, heaven *v1alpha1.TrishankuHeaven, c configNames) (d *appsv1.Deployment, err error) {
	var (
		podSpec       *corev1.PodSpec
		ntic, ntc     int
		apiserver     corev1.Container
		useEventsEtcd = heaven.Spec.EventsEtcd != nil && heaven.Spec.EventsEtcd.Local != nil
	)

	if heaven.Spec.Gitcd == nil {
		err = errors.New("no gitcd configuration found")
		return
	}

	d = r.newDeployment(heaven, heaven.Spec.App.Replicas, &heaven.Spec.App.PodTemplate, c, DEPLOY_HEAVEN)
	podSpec = &d.Spec.Template.Spec
	ntic, ntc = len(podSpec.InitContainers), len(podSpec.Containers)

	for _, secret := range []string{SECRET_CERT_KUBERNETES, SECRET_CERT_SERVICE_ACCOUNTS} {
		podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
			Name:         secret,
			VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: c.getConfigurationName(secret)}},
		})
	}

	r.appendVolumesForGitcd(podSpec, c)

	if useEventsEtcd {
		podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
			Name:         CONTAINER_ETCD,
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		})
	}

	r.appendImagePullSecrets(podSpec, heaven.Spec.Gitcd)

	r.appendGitcdContainers(podSpec, heaven.Spec.Gitcd, getCommitterName(heaven))

	apiserver = corev1.Container{
		Name:            CONTAINER_APISERVER,
		Image:           getImage((*v1alpha1.ImageSpec)(heaven.Spec.Apiserver), r.DefaultApiserverImage),
		ImagePullPolicy: getImagePullPolicy((*v1alpha1.ImageSpec)(heaven.Spec.Apiserver)),
		Command: []string{
			"kube-apiserver",
			"--advertise-address=$(POD_IP)",
			"--allow-privileged=true",
			"--audit-log-maxage=30",
			"--audit-log-maxbackup=3",
			"--audit-log-maxsize=100",
			"--audit-log-path=/var/log/audit.log",
			"--authorization-mode=Node,RBAC",
			"--bind-address=0.0.0.0",
			"--client-ca-file=/.trishanku/secrets/cert-kubernetes/ca.crt",
			"--enable-admission-plugins=NamespaceLifecycle,NodeRestriction,LimitRanger,ServiceAccount,DefaultStorageClass,ResourceQuota",
			"--enable-garbage-collector=false",
			"--etcd-compaction-interval=0",
			"--etcd-count-metric-poll-period=0",
			"--etcd-db-metric-poll-interval=0",
			"--etcd-healthcheck-timeout=10s",
			"--etcd-servers=http://127.0.0.1:2479",
			"--event-ttl=1h",
			"--kubelet-certificate-authority=/.trishanku/secrets/cert-kubernetes/ca.crt",
			"--kubelet-client-certificate=/.trishanku/secrets/cert-kubernetes/tls.crt",
			"--kubelet-client-key=/.trishanku/secrets/cert-kubernetes/tls.key",
			"--lease-reuse-duration-seconds=120",
			"--service-account-key-file=/.trishanku/secrets/cert-service-account/tls.crt",
			"--service-account-signing-key-file=/.trishanku/secrets/cert-service-account/tls.key",
			"--service-account-issuer=https://$(POD_IP):6443",
			"--service-cluster-ip-range=10.32.0.0/24",
			"--service-node-port-range=30000-32767",
			"--storage-media-type=application/yaml",
			"--tls-cert-file=/.trishanku/secrets/cert-kubernetes/tls.crt",
			"--tls-private-key-file=/.trishanku/secrets/cert-kubernetes/tls.key",
			"--v=2",
			"--watch-cache=false",
		},
		Env: []corev1.EnvVar{
			{Name: "POD_IP", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}}},
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: SECRET_CERT_KUBERNETES, MountPath: path.Join(BASE_PATH_SECRETS, SECRET_CERT_KUBERNETES), ReadOnly: true},
			{Name: SECRET_CERT_SERVICE_ACCOUNTS, MountPath: path.Join(BASE_PATH_SECRETS, SECRET_CERT_SERVICE_ACCOUNTS), ReadOnly: true},
		},
	}

	if useEventsEtcd {
		podSpec.Containers = append(podSpec.Containers, corev1.Container{
			Name:            CONTAINER_ETCD,
			Image:           getImage((*v1alpha1.ImageSpec)(heaven.Spec.EventsEtcd.Local), r.DefaultEtcdImage),
			ImagePullPolicy: getImagePullPolicy((*v1alpha1.ImageSpec)(heaven.Spec.EventsEtcd.Local)),
			Command: []string{
				"etcd",
				"--initial-cluster-token=trishanku",
				"--data-dir=/var/lib/etcd",
			},
			VolumeMounts: []corev1.VolumeMount{
				{Name: CONTAINER_ETCD, MountPath: "/var/lib/etcd"},
			},
		})

		apiserver.Command = append(apiserver.Command, "--etcd-servers-overrides=/events#http://127.0.0.1:2379")
	}

	podSpec.Containers = append(podSpec.Containers, apiserver)

	r.appendGitPostInitContainer(podSpec, heaven.Spec.Gitcd)

	if len(heaven.Spec.App.KubeconfigMountPath) > 0 {
		var baseMountPath, kubeconfigFileName = path.Split(heaven.Spec.App.KubeconfigMountPath)

		setKubeconfigVolume(heaven, podSpec, kubeconfigFileName, c)

		for i := 0; i < ntic; i++ {
			setKubeconfigVolumeMount(&d.Spec.Template.Spec.InitContainers[i], baseMountPath)
		}

		for i := 0; i < ntc; i++ {
			setKubeconfigVolumeMount(&d.Spec.Template.Spec.Containers[i], baseMountPath)
		}
	}

	return
}

func setKubeconfigVolume(heaven *v1alpha1.TrishankuHeaven, podSpec *corev1.PodSpec, filePath string, c configNames) {
	podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
		Name: SECRET_KUBECONFIG_CONTROLLER,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: c.getConfigurationName(SECRET_KUBECONFIG_CONTROLLER),
				Items: []corev1.KeyToPath{
					{
						Key:  certs.KeyKubeconfig,
						Path: filePath,
					},
				},
			},
		},
	})
}

// setKubeconfigVolumeMount sets the volume mount for the controller kubeconfig.
func setKubeconfigVolumeMount(c *corev1.Container, kubeconfigMountPath string) {
	// Add VolumeMount for kubeconfigMountPath
	c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
		Name:      SECRET_KUBECONFIG_CONTROLLER,
		MountPath: kubeconfigMountPath,
	})
}

func (r *TrishankuHeavenReconciler) ensureDeployment(heaven *v1alpha1.TrishankuHeaven) ensureFunc {
	return func(ctx context.Context, owner metav1.Object, c configNames) (err error) {
		var d, rd *appsv1.Deployment

		if heaven.Spec.Skip.App {
			d = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      heaven.Name,
					Namespace: heaven.Namespace,
				},
			}

			if err = r.Delete(ctx, d); apierrors.IsNotFound(err) {
				err = nil // Ignore if deployment does not exist.
			}

			return
		}

		if heaven.Spec.Certificates.Controller == nil {
			return fmt.Errorf("controller certificates not configured for %q", client.ObjectKeyFromObject(heaven).String())
		}

		if heaven.Spec.App == nil {
			return fmt.Errorf("controller app not configured for %q", client.ObjectKeyFromObject(heaven).String())
		}

		if rd, err = r.generateDeploymentFor(ctx, heaven, c); err != nil {
			return
		}

		err = r.createOrUpdateDeployment(ctx, owner, rd)
		return
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *TrishankuHeavenReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("heaven").
		For(&v1alpha1.TrishankuHeaven{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
