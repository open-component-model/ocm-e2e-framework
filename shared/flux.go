package shared

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fluxcd/flux2/v2/pkg/manifestgen"
	"github.com/fluxcd/flux2/v2/pkg/manifestgen/install"
	runclient "github.com/fluxcd/pkg/runtime/client"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"

	"github.com/open-component-model/ocm-e2e-framework/internal/utils"
)

const (
	maximumQueriesPerSecond = 50.0
	burst                   = 300
)

// InstallFlux creates a flux installation with a given version.
func InstallFlux(version string) env.Func {
	// add files to cluster
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		// download version
		tmpDir, err := manifestgen.MkdirTempAbs("", "ocm-system")
		if err != nil {
			return ctx, err
		}

		defer os.RemoveAll(tmpDir)

		opts := install.MakeDefaultOptions()
		opts.Version = version

		manifest, err := install.Generate(opts, "")
		if err != nil {
			return ctx, fmt.Errorf("install generate failed: %w", err)
		}

		if _, err := manifest.WriteFile(tmpDir); err != nil {
			return ctx, fmt.Errorf("install write failed: %w", err)
		}

		kubeConfig := cfg.KubeconfigFile()
		kfg := genericclioptions.ConfigFlags{KubeConfig: &kubeConfig}
		runOpts := &runclient.Options{
			QPS:   maximumQueriesPerSecond,
			Burst: burst,
		}

		if _, err = utils.Apply(ctx, &kfg, runOpts, tmpDir, filepath.Join(tmpDir, manifest.Path)); err != nil {
			return ctx, fmt.Errorf("install apply failed: %w", err)
		}

		return ctx, nil
	}
}
