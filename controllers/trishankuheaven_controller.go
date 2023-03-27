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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/trishanku/heaven/api/v1alpha1"
	"github.com/trishanku/heaven/pkg/certs"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientcmdv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

// TrishankuHeavenReconciler reconciles a TrishankuHeaven object
type TrishankuHeavenReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	DefaultEtcdImage      string
	DefaultCfsslImage     string
	DefaultKubectlImage   string
	DefaultGitImage       string
	DefaultGitcdImage     string
	DefaultApiserverImage string
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

	c.init(heaven)

	l.Info("configNames", "configNames", c)

	if !heaven.Spec.Skip.Cerfificates {
		ensureFns = append(ensureFns, r.ensureCertificateSecrets(heaven, c)...)
	}

	if !heaven.Spec.Skip.Entrypoints {
		ensureFns = append(ensureFns, r.ensureEntrypointsConfigMap)
	}

	ensureFns = append(ensureFns, r.ensureDeployment)

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
	VOLUME_HOME                  = "home"
	VOLUME_HOME_PR               = "home-pr"
	VOLUME_TEMP                  = "temp"

	CONTAINER_ETCD       = "etcd"
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

func (c configNames) init(heaven *v1alpha1.TrishankuHeaven) {
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

type ensureFunc func(context.Context, *v1alpha1.TrishankuHeaven, configNames) error

func (r *TrishankuHeavenReconciler) ensureCertificateSecrets(heaven *v1alpha1.TrishankuHeaven, c configNames) (fns []ensureFunc) {
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
			fns = append(fns, func(ctx context.Context, heaven *v1alpha1.TrishankuHeaven, c configNames) (err error) {
				var (
					secretName = c.getConfigurationName(cs.name)
					namespace  = heaven.Namespace

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

				if err = r.Create(ctx, s); err != nil {
					return
				}

				return
			})
		}(cs)
	}

	return
}

func (r *TrishankuHeavenReconciler) ensureEntrypointsConfigMap(ctx context.Context, heaven *v1alpha1.TrishankuHeaven, c configNames) (err error) {
	var cm = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.getConfigurationName(CONFIG_ENTRYPOINTS),
			Namespace: heaven.Namespace,
		},
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, cm, func() error {
		cm.Name = c.getConfigurationName(CONFIG_ENTRYPOINTS)
		cm.Namespace = heaven.Namespace

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
git fetch || exit 1

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
		return nil
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

func getCreateBranchEnvVar(name string, branch *v1alpha1.BranchSpec) corev1.EnvVar {
	var create bool

	if branch != nil {
		create = !branch.NoCreateBranch
	}

	return corev1.EnvVar{Name: name, Value: strconv.FormatBool(create)}
}

func getRemoteBranchData(heaven *v1alpha1.TrishankuHeaven) string {
	if heaven != nil && heaven.Spec.Gitcd.Branches.Remote != nil {
		return heaven.Spec.Gitcd.Branches.Remote.Data
	}

	return ""
}

func getRemoteBranchMetadata(heaven *v1alpha1.TrishankuHeaven) string {
	if heaven != nil && heaven.Spec.Gitcd.Branches.Remote != nil {
		return heaven.Spec.Gitcd.Branches.Remote.Metadata
	}

	return ""
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
func getMergeCommitterName(heaven *v1alpha1.TrishankuHeaven) string {
	if heaven.Spec.Gitcd.PullRequest != nil && len(heaven.Spec.Gitcd.PullRequest.MergeCommitterName) > 0 {
		return heaven.Spec.Gitcd.PullRequest.MergeCommitterName
	}

	return getCommitterName(heaven)
}

func (r *TrishankuHeavenReconciler) generateDeploymentFor(ctx context.Context, heaven *v1alpha1.TrishankuHeaven, c configNames) (d *appsv1.Deployment, err error) {
	var (
		podSpec                   *corev1.PodSpec
		ntic, ntc                 int
		gitcd, apiserver, gitcdPR corev1.Container
		useEventsEtcd             = heaven.Spec.EventsEtcd != nil && heaven.Spec.EventsEtcd.Local != nil
		gitCredsSecretName        string
	)

	if heaven.Spec.Gitcd == nil {
		err = errors.New("no gitcd configuration found")
		return
	}

	gitCredsSecretName = heaven.Spec.Gitcd.CredentialsSecretName

	d = &appsv1.Deployment{
		ObjectMeta: *heaven.Spec.App.PodTemplate.ObjectMeta.DeepCopy(),
		Spec: appsv1.DeploymentSpec{
			Replicas: heaven.Spec.App.Replicas,
			Selector: &metav1.LabelSelector{MatchLabels: heaven.Spec.App.PodTemplate.ObjectMeta.Labels},
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: *heaven.Spec.App.PodTemplate.ObjectMeta.DeepCopy(),
				Spec:       *heaven.Spec.App.PodTemplate.Spec.DeepCopy(),
			},
		},
	}

	d.Name, d.Namespace = heaven.Name, heaven.Namespace

	podSpec = &d.Spec.Template.Spec
	ntic, ntc = len(podSpec.InitContainers), len(podSpec.Containers)

	for _, secret := range []string{SECRET_CERT_KUBERNETES, SECRET_CERT_SERVICE_ACCOUNTS} {
		podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
			Name:         secret,
			VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: c.getConfigurationName(secret)}},
		})
	}

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
		corev1.Volume{
			Name:         VOLUME_TEMP,
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{Medium: corev1.StorageMediumMemory}},
		},
	)

	if useEventsEtcd {
		podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
			Name:         CONTAINER_ETCD,
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		})
	}

	if len(heaven.Spec.Gitcd.ImagePullSecretName) > 0 {
		podSpec.ImagePullSecrets = append(podSpec.ImagePullSecrets, corev1.LocalObjectReference{Name: heaven.Spec.Gitcd.ImagePullSecretName})
	}

	podSpec.InitContainers = append(
		podSpec.InitContainers,
		corev1.Container{
			Name:            CONTAINER_GIT_PRE,
			Image:           getImage(heaven.Spec.Gitcd.GitImage, r.DefaultGitImage),
			ImagePullPolicy: getImagePullPolicy(heaven.Spec.Gitcd.GitImage),
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
				{Name: "GITCD_COMMITTER_NAME", Value: getCommitterName(heaven)},
				{Name: "GITCD_BRANCH_DATA", Value: heaven.Spec.Gitcd.Branches.Local.Data},
				{Name: "GITCD_BRANCH_METADATA", Value: heaven.Spec.Gitcd.Branches.Local.Metadata},
				getCreateBranchEnvVar("GITCD_CREATE_LOCAL_BRANCH", &heaven.Spec.Gitcd.Branches.Local),
				{Name: "GITCD_REMOTE_REPO", Value: heaven.Spec.Gitcd.RemoteRepo},
				{Name: "GITCD_REMOTE_BRANCH_DATA", Value: getRemoteBranchData(heaven)},
				{Name: "GITCD_REMOTE_BRANCH_METADATA", Value: getRemoteBranchMetadata(heaven)},
				getCreateBranchEnvVar("GITCD_CREATE_REMOTE_BRANCH", heaven.Spec.Gitcd.Branches.Remote),
			},
			VolumeMounts: []corev1.VolumeMount{
				{Name: CONFIG_ENTRYPOINTS, MountPath: BASE_PATH_ENTRYPOINTS, ReadOnly: true},
				{Name: VOLUME_HOME, MountPath: BASE_PATH_HOME},
			},
		},
		corev1.Container{
			Name:            CONTAINER_GITCD_INIT,
			Image:           getImage(heaven.Spec.Gitcd.GitcdImage, r.DefaultGitcdImage),
			ImagePullPolicy: getImagePullPolicy(heaven.Spec.Gitcd.GitcdImage),
			Args: []string{
				"init",
				"--repo=/root/repo",
				"--data-reference-names=default=refs/heads/" + heaven.Spec.Gitcd.Branches.Local.Data,
				"--metadata-reference-names=default=refs/heads/" + heaven.Spec.Gitcd.Branches.Local.Metadata,
			},
			VolumeMounts: []corev1.VolumeMount{
				{Name: VOLUME_HOME, MountPath: BASE_PATH_HOME},
			},
		},
	)

	gitcd = corev1.Container{
		Name:            CONTAINER_GITCD,
		Image:           getImage(heaven.Spec.Gitcd.GitcdImage, r.DefaultGitcdImage),
		ImagePullPolicy: getImagePullPolicy(heaven.Spec.Gitcd.GitcdImage),
		Command: []string{
			"/gitcd",
			"serve",
			"--repo=/root/repo",
			"--committer-name=" + getCommitterName(heaven),
			"--data-reference-names=default=refs/heads/" + heaven.Spec.Gitcd.Branches.Local.Data,
			"--metadata-reference-names=default=refs/heads/" + heaven.Spec.Gitcd.Branches.Local.Metadata,
			"--key-prefixes=default=/registry",
			"--pull-ticker-duration=" + heaven.Spec.Gitcd.Pull.TickerDuration.Duration.String(),
			"--remote-names=default=origin",
			"--no-fast-forwards=default=false",
			"--remote-data-reference-names=default=refs/heads/" + getRemoteBranchData(heaven),
			"--remote-meta-reference-names=default=refs/heads/" + getRemoteBranchMetadata(heaven),
			"--listen-urls=default=http://0.0.0.0:2479/",
			"--advertise-client-urls=default=http://127.0.0.1:2479/",
			"--watch-dispatch-channel-size=1",
			"--push-after-merges=default=" + strconv.FormatBool(heaven.Spec.Gitcd.Pull.PushAfterMerge),
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: VOLUME_HOME, MountPath: BASE_PATH_HOME},
		},
	}

	appendMergeFlagsToCommand(
		&gitcd,
		map[string]string{
			"--merge-retention-policies-include": heaven.Spec.Gitcd.Pull.RetentionPolicies.Include,
			"--merge-retention-policies-exclude": heaven.Spec.Gitcd.Pull.RetentionPolicies.Exclude,
			"--merge-conflict-resolutions":       heaven.Spec.Gitcd.Pull.ConflictResolutions,
		},
	)

	podSpec.Containers = append(podSpec.Containers, gitcd)

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

	if heaven.Spec.Gitcd.PullRequest != nil && heaven.Spec.Gitcd.PullRequest.TickerDuration.Duration > 0 {
		podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
			Name:         VOLUME_HOME_PR,
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		})

		podSpec.InitContainers = append(
			podSpec.InitContainers,
			corev1.Container{
				Name:            CONTAINER_GIT_POST,
				Image:           getImage(heaven.Spec.Gitcd.GitImage, r.DefaultGitImage),
				ImagePullPolicy: getImagePullPolicy(heaven.Spec.Gitcd.GitImage),
				Command: []string{
					"git",
					"push",
				},
				VolumeMounts: []corev1.VolumeMount{
					{Name: VOLUME_HOME, MountPath: BASE_PATH_HOME},
				},
				WorkingDir: "/root/repo",
			},
			corev1.Container{
				Name:            CONTAINER_GIT_PRE_PR,
				Image:           getImage(heaven.Spec.Gitcd.GitImage, r.DefaultGitImage),
				ImagePullPolicy: getImagePullPolicy(heaven.Spec.Gitcd.GitImage),
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
					{Name: "GITCD_COMMITTER_NAME", Value: getMergeCommitterName(heaven)},
					{Name: "GITCD_BRANCH_DATA", Value: getRemoteBranchData(heaven)},
					{Name: "GITCD_BRANCH_METADATA", Value: getRemoteBranchMetadata(heaven)},
					getCreateBranchEnvVar("GITCD_CREATE_LOCAL_BRANCH", heaven.Spec.Gitcd.Branches.Remote),
					{Name: "GITCD_REMOTE_REPO", Value: heaven.Spec.Gitcd.RemoteRepo},
					{Name: "GITCD_REMOTE_BRANCH_DATA", Value: heaven.Spec.Gitcd.Branches.Local.Data},
					{Name: "GITCD_REMOTE_BRANCH_METADATA", Value: heaven.Spec.Gitcd.Branches.Local.Metadata},
					getCreateBranchEnvVar("GITCD_CREATE_REMOTE_BRANCH", &heaven.Spec.Gitcd.Branches.Local),
					{Name: "GITCD_PUSH_AFTER_INIT", Value: "x"},
				},
				VolumeMounts: []corev1.VolumeMount{
					{Name: CONFIG_ENTRYPOINTS, MountPath: BASE_PATH_ENTRYPOINTS, ReadOnly: true},
					{Name: VOLUME_HOME_PR, MountPath: BASE_PATH_HOME},
				},
			},
		)

		gitcdPR = corev1.Container{
			Name:            CONTAINER_GITCD_PR,
			Image:           getImage(heaven.Spec.Gitcd.GitcdImage, r.DefaultGitcdImage),
			ImagePullPolicy: getImagePullPolicy(heaven.Spec.Gitcd.GitcdImage),
			Command: []string{
				"/gitcd",
				"pull",
				"--repo=/root/repo",
				"--committer-name=pr-" + getCommitterName(heaven),
				"--data-reference-names=default=refs/heads/" + getRemoteBranchData(heaven),
				"--metadata-reference-names=default=refs/heads/" + getRemoteBranchMetadata(heaven),
				"--no-fast-forwards=default=false",
				"--push-after-merges=default=" + strconv.FormatBool(heaven.Spec.Gitcd.PullRequest.PushAfterMerge),
				"--remote-names=default=origin",
				"--remote-data-reference-names=default=refs/heads/" + heaven.Spec.Gitcd.Branches.Local.Data,
				"--remote-meta-reference-names=default=refs/heads/" + heaven.Spec.Gitcd.Branches.Local.Metadata,
				"--pull-ticker-duration=" + heaven.Spec.Gitcd.PullRequest.TickerDuration.Duration.String(),
			},
			VolumeMounts: []corev1.VolumeMount{
				{Name: VOLUME_HOME_PR, MountPath: BASE_PATH_HOME},
			},
		}

		appendMergeFlagsToCommand(
			&gitcdPR,
			map[string]string{
				"--merge-retention-policies-include": heaven.Spec.Gitcd.Pull.RetentionPolicies.Include,
				"--merge-retention-policies-exclude": heaven.Spec.Gitcd.Pull.RetentionPolicies.Exclude,
				"--merge-conflict-resolutions":       heaven.Spec.Gitcd.PullRequest.ConflictResolutions,
			},
		)

		podSpec.Containers = append(podSpec.Containers, gitcdPR)
	}

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

func (r *TrishankuHeavenReconciler) ensureDeployment(ctx context.Context, heaven *v1alpha1.TrishankuHeaven, c configNames) (err error) {
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

	d = rd.DeepCopy()

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, d, func() (err error) {
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
		return
	})

	return
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
