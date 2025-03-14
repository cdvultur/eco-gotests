package setup

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/assisted"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/hive"
	"github.com/openshift-kni/eco-goinfra/pkg/namespace"
	"github.com/openshift-kni/eco-goinfra/pkg/schemes/assisted/api/hiveextension/v1beta1"
	"github.com/openshift-kni/eco-goinfra/pkg/secret"
	. "github.com/openshift-kni/eco-gotests/tests/assisted/ztp/internal/ztpinittools"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

// SpokeClusterResources contains necessary resources for creating a spoke cluster.
type SpokeClusterResources struct {
	Name                string
	apiClient           *clients.Settings
	err                 error
	Namespace           *namespace.Builder
	PullSecret          *secret.Builder
	ClusterDeployment   *hive.ClusterDeploymentBuilder
	AgentClusterInstall *assisted.AgentClusterInstallBuilder
	InfraEnv            *assisted.InfraEnvBuilder
}

// NewSpokeCluster creates a new instance of SpokeClusterResources.
func NewSpokeCluster(apiClient *clients.Settings) *SpokeClusterResources {
	return &SpokeClusterResources{apiClient: apiClient}
}

// WithName sets an explicit name for the spoke cluster.
func (spoke *SpokeClusterResources) WithName(name string) *SpokeClusterResources {
	if name == "" {
		spoke.err = fmt.Errorf("spoke name cannot be empty")
	}

	spoke.Name = name

	return spoke
}

// WithAutoGeneratedName generates a random name for the spoke cluster.
func (spoke *SpokeClusterResources) WithAutoGeneratedName() *SpokeClusterResources {
	spoke.Name = generateName(12)

	return spoke
}

// WithDefaultNamespace creates a default namespace for the spoke cluster.
func (spoke *SpokeClusterResources) WithDefaultNamespace() *SpokeClusterResources {
	spoke.Namespace = namespace.NewBuilder(spoke.apiClient, spoke.Name)

	return spoke
}

// WithDefaultPullSecret creates a default pull-secret for the spoke cluster.
func (spoke *SpokeClusterResources) WithDefaultPullSecret() *SpokeClusterResources {
	spoke.PullSecret = secret.NewBuilder(
		spoke.apiClient,
		fmt.Sprintf("%s-pull-secret", spoke.Name),
		spoke.Name,
		corev1.SecretTypeDockerConfigJson).WithData(ZTPConfig.HubPullSecret.Object.Data)

	return spoke
}

// WithDefaultClusterDeployment creates a default clusterdeployment for the spoke cluster.
func (spoke *SpokeClusterResources) WithDefaultClusterDeployment() *SpokeClusterResources {
	spoke.ClusterDeployment = hive.NewABMClusterDeploymentBuilder(
		spoke.apiClient,
		spoke.Name,
		spoke.Name,
		spoke.Name,
		"assisted.test.com",
		spoke.Name,
		metav1.LabelSelector{
			MatchLabels: map[string]string{
				"dummy": "label",
			},
		}).WithPullSecret(fmt.Sprintf("%s-pull-secret", spoke.Name))

	return spoke
}

// WithDefaultIPv4AgentClusterInstall creates a default agentclusterinstall with IPv4 networking for the spoke cluster.
func (spoke *SpokeClusterResources) WithDefaultIPv4AgentClusterInstall() *SpokeClusterResources {
	spoke.AgentClusterInstall = assisted.NewAgentClusterInstallBuilder(
		spoke.apiClient,
		spoke.Name,
		spoke.Name,
		spoke.Name,
		3,
		2,
		v1beta1.Networking{
			ClusterNetwork: []v1beta1.ClusterNetworkEntry{{
				CIDR:       "10.128.0.0/14",
				HostPrefix: 23,
			}},
			ServiceNetwork: []string{"172.30.0.0/16"},
		}).WithImageSet(ZTPConfig.HubOCPXYVersion).WithAPIVip("192.168.254.5").WithIngressVip("192.168.254.10")

	return spoke
}

// WithDefaultIPv6AgentClusterInstall creates a default agentclusterinstall with IPv6 networking for the spoke cluster.
func (spoke *SpokeClusterResources) WithDefaultIPv6AgentClusterInstall() *SpokeClusterResources {
	spoke.AgentClusterInstall = assisted.NewAgentClusterInstallBuilder(
		spoke.apiClient,
		spoke.Name,
		spoke.Name,
		spoke.Name,
		3,
		2,
		v1beta1.Networking{
			ClusterNetwork: []v1beta1.ClusterNetworkEntry{{
				CIDR:       "fd01::/48",
				HostPrefix: 64,
			}},
			ServiceNetwork: []string{"fd02::/112"},
		}).WithImageSet(ZTPConfig.HubOCPXYVersion).WithAPIVip("fd2e:6f44:5dd8:1::5").WithIngressVip("fd2e:6f44:5dd8:1::10")

	return spoke
}

// WithDefaultDualStackAgentClusterInstall creates a default agentclusterinstall
// with dual-stack networking for the spoke cluster.
func (spoke *SpokeClusterResources) WithDefaultDualStackAgentClusterInstall() *SpokeClusterResources {
	spoke.AgentClusterInstall = assisted.NewAgentClusterInstallBuilder(
		spoke.apiClient,
		spoke.Name,
		spoke.Name,
		spoke.Name,
		3,
		2,
		v1beta1.Networking{
			ClusterNetwork: []v1beta1.ClusterNetworkEntry{
				{
					CIDR:       "10.128.0.0/14",
					HostPrefix: 23,
				},
				{
					CIDR:       "fd01::/48",
					HostPrefix: 64,
				},
			},
			ServiceNetwork: []string{"172.30.0.0/16", "fd02::/112"},
		}).WithImageSet(ZTPConfig.HubOCPXYVersion).WithAPIVip("192.168.254.5").WithIngressVip("192.168.254.10")

	return spoke
}

// WithDefaultInfraEnv creates a default infraenv for the spoke cluster.
func (spoke *SpokeClusterResources) WithDefaultInfraEnv() *SpokeClusterResources {
	spoke.InfraEnv = assisted.NewInfraEnvBuilder(
		spoke.apiClient,
		spoke.Name,
		spoke.Name,
		fmt.Sprintf("%s-pull-secret", spoke.Name))

	return spoke
}

// Create creates the instantiated spoke cluster resources.
func (spoke *SpokeClusterResources) Create() (*SpokeClusterResources, error) {
	if spoke.Namespace != nil && spoke.err == nil {
		spoke.Namespace, spoke.err = spoke.Namespace.Create()
	}

	if spoke.PullSecret != nil && spoke.err == nil {
		spoke.PullSecret, spoke.err = spoke.PullSecret.Create()
	}

	if spoke.ClusterDeployment != nil && spoke.err == nil {
		spoke.ClusterDeployment, spoke.err = spoke.ClusterDeployment.Create()
	}

	if spoke.AgentClusterInstall != nil && spoke.err == nil {
		spoke.AgentClusterInstall, spoke.err = spoke.AgentClusterInstall.Create()
	}

	if spoke.InfraEnv != nil && spoke.err == nil {
		spoke.InfraEnv, spoke.err = spoke.InfraEnv.Create()
	}

	return spoke, spoke.err
}

// Delete removes all instantiated spoke cluster resources.
func (spoke *SpokeClusterResources) Delete() error {
	if spoke.InfraEnv != nil {
		spoke.err = spoke.InfraEnv.Delete()
	}

	if spoke.AgentClusterInstall != nil {
		spoke.err = spoke.AgentClusterInstall.Delete()
	}

	if spoke.ClusterDeployment != nil {
		spoke.err = spoke.ClusterDeployment.Delete()
	}

	if spoke.PullSecret != nil {
		spoke.err = spoke.PullSecret.Delete()
	}

	if spoke.Namespace != nil {
		spoke.err = spoke.Namespace.DeleteAndWait(time.Second * 120)
	}

	return spoke.err
}

// generateName generates a random string matching the length supplied.
func generateName(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}
