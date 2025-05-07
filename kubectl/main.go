// Based on the work done here:
// https://github.com/matipan/daggerverse/blob/84cbdbe89185ad94690a9ada1cdfb79f1878ecd7/kubectl/main.go
// `kubectl` that provides the kubectl through many authentication methods.
package main

import (
	"context"
	"dagger/kubectl/internal/dagger"
	"fmt"
)

const (
	Version = "1.33.0"
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

// KubectlEks returns a KubectlCLI with aws-iam-authenticator and AWS credentials
// configured to communicate with an EKS cluster.
func (m *Kubectl) KubectlContainer(_ context.Context) *KubectlCLI {
	c := dag.Container().
		From(fmt.Sprintf("bitnami/kubectl:%s", Version)).
		WithMountedSecret("/root/.kube/config", m.Kubeconfig)

	return &KubectlCLI{
		Container: c.
			WithEntrypoint([]string{"/bin/kubectl"}),
	}
}
