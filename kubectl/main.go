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

// Kube returns a KubectlCLI
func (m *Kubectl) Kube(ctx context.Context, version string) *KubectlCLI {
	if version == "" {
		version = "1.33.0"
	}

	c := dag.Container().
		From(fmt.Sprintf("bitnami/kubectl:%s", version)).
		WithMountedSecret("/root/.kube/config", m.Kubeconfig)

	return &KubectlCLI{
		Container: c.
			WithEntrypoint([]string{"/bin/kubectl"}),
	}
}
