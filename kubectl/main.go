// `kubectl` that provides the kubectl through many authentication methods.
// The goal of this module is to be the one stop shop for interacting with a
// kubernetes cluster. The main challenge when doing this is how authentication
// is done against the cluster. Eventually this module should support all most
// used methods.
// Each top level method is in charge of creating a container with all the tools
// and credentials ready to go for kubectl commands to be executed. For an example on how this is implemented you can check out the `KubectlEks` method.
package main

import (
	"context"
	"dagger/kubectl/internal/dagger"
	"fmt"
)

type Kubectl struct {
	Kubeconfig *dagger.Secret
}

// New creates a new instance of the Kubectl module with an already configured
// kubeconfig file. Kubectl is the top level module that provides functions setting
// up the authentication for a specific k8s setup.
func New(kubeconfig *dagger.Secret) *Kubectl {
	return &Kubectl{
		Kubeconfig: kubeconfig,
	}
}

// KubectlCLI is a child module that holds a Container that should already
// be configured to talk to a k8s cluster.
type KubectlCLI struct {
	Container *dagger.Container
}

// Exec runs the specified kubectl command.
// NOTE: `kubectl` should be specified as part of the cmd variable.
// For example, to list pods: ["get", "pods", "-n", "namespace"]
func (k *KubectlCLI) Exec(ctx context.Context, cmd []string) (string, error) {
	return k.Container.WithExec(cmd).Stdout(ctx)
}

// DebugSh is a helper function that developers can use to get a terminal
// into the container where the commands are run and troubleshoot potential
// misconfigurations.
// For example:
// dagger call --kubeconfig kubeconfig.yaml kubectl-eks --aws-creds ~/.aws/credentials --aws-profile "example" --aws-config ~/.aws/config debug-sh terminal
func (k *KubectlCLI) DebugSh() *dagger.Container {
	return k.Container.WithoutEntrypoint()
}

// KubectlEks returns a KubectlCLI with aws-iam-authenticator and AWS credentials
// configured to communicate with an EKS cluster.
func (m *Kubectl) KubectlEks(_ context.Context) *KubectlCLI {
	c := dag.Container().
		From(fmt.Sprintf("bitnami/kubectl:%s", "1.33.0")).
		// WithExec([]string{"apk", "add", "--no-cache", "--update", "ca-certificates", "curl"}).
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "-y", "curl", "gettext-base"}).
		WithMountedSecret("/root/.kube/config", m.Kubeconfig)
	return &KubectlCLI{
		Container: c.
			WithEntrypoint([]string{"/bin/kubectl"}),
	}
}
