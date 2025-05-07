package main

import (
	"context"
	"dagger/kubectl/internal/dagger"
	"fmt"
)

type Kubectl struct{}

func (m *Kubectl) Cli(config *dagger.Secret) *Cli {
	return &Cli{Config: config}
}

type Cli struct {
	Config *dagger.Secret
}

// Exec runs the specified kubectl command.
// NOTE: `kubectl` should be specified as part of the cmd variable.
// For example, to list pods: ["get", "pods", "-n", "namespace"]
func (m *Cli) Exec(ctx context.Context, args []string) (string, error) {
	if k.Config == nil {
		return "", fmt.Errorf("please provide a kubectl config")
	}

	return m.container().WithExec(args).Stdout(ctx)
}

// Container returns a container with the kubectl image and given config. The entrypoint is set to kubectl.
func (m *Cli) Container(_ context.Context) (*dagger.Container, error) {
	if m.Config == nil {
		return nil, fmt.Errorf("please provide a kubectl config")
	}

	return m.container(), nil
}

// Kubectl returns a KubectlCLI
func (m *Cli) container() *dagger.Container {
	return dag.Container().
		From(fmt.Sprintf("bitnami/kubectl:%s", "1.33.0")).
		WithUser("root").
		WithMountedSecret("/root/.kube/config", m.Config)
}
