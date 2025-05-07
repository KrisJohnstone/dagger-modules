// Based on th work done by purpleclay here:
// https://github.com/purpleclay/daggerverse/blob/16106a67856f0716aad4c1ceaa033f0e7193abca/helm-oci/main.go
package main

import (
	"context"
	"dagger/helm/internal/dagger"
	"fmt"
	"helm.sh/helm/v3/pkg/chart"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
)

const (
	helmVersion = "3.17"
	workDir     = "/tmp/workdir"
)

type Helm struct {
	// +private
	Base *dagger.Container
}

// Initializes the Helm dagger module
func New(
	ctx context.Context,
// a custom base image containing an installation of helm
// +optional
	base *dagger.Container,
) (*Helm, error) {
	var err error
	if base == nil {
		base, err = defaultImage()
	} else {
		if _, err = base.WithExec([]string{"helm", "version"}).Sync(ctx); err != nil {
			return nil, err
		}
	}

	base = base.WithUser("root").
		WithoutEnvVariable("HELM_HOME").
		WithoutEnvVariable("HELM_REGISTRY_CONFIG")

	return &Helm{Base: base}, err
}

func defaultImage() (*dagger.Container, error) {
	return dag.Container().
		From(fmt.Sprintf("alpine/helm:%s", helmVersion)), nil
}

func resolveChartMetadata(ctx context.Context, dir *dagger.Directory) (*chart.Metadata, error) {
	manifest, err := dir.File("Chart.yaml").Contents(ctx)
	if err != nil {
		return nil, err
	}

	metadata := &chart.Metadata{}
	if err := yaml.Unmarshal([]byte(manifest), metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

func (m *Helm) Template(
	ctx context.Context,
// a path to the directory containing the Chart.yaml file and all templates
// +required
	dir *dagger.Directory,
// specify values in a YAML file bundled within the chart directory (can specify multiple)
// +optional
	values []string,
// specify values in external YAML files loaded from the file system (can specify multiple).
// These have a higher precedence over other values files
// +optional
	valuesExt []*dagger.File) (*dagger.File, error) {
	chart, err := resolveChartMetadata(ctx, dir)
	if err != nil {
		return nil, err
	}

	cmd := []string{"helm", "template", "."}

	cmd = append(cmd, toFlags("--values", values)...)

	ctr := m.Base.
		WithMountedDirectory(workDir, dir).
		WithWorkdir(workDir)

	// Ensure values files loaded externally from the chart have higher precedence
	for i, ext := range valuesExt {
		tmpValues := filepath.Join(os.TempDir(), fmt.Sprintf("values-%d.yaml", i+1))
		ctr = ctr.WithFile(tmpValues, ext)
		cmd = append(cmd, "--values", tmpValues)
	}

	template := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s.yaml", strings.ToLower(chart.Name), chart.Version))

	return ctr.
		WithExec([]string{"helm", "dependency", "build"}).
		WithExec(cmd, dagger.ContainerWithExecOpts{RedirectStdout: template}).
		File(template), nil
}

func toFlags(flag string, values []string) []string {
	var flags []string
	for _, v := range values {
		flags = append(flags, flag, v)
	}
	return flags
}
