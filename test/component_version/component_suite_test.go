package component_version //nolint:stylecheck

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
)

var (
	testEnv         env.Environment
	kindClusterName string
	namespace       string
)

const (
	defaultTimeoutSeconds = 600
)

// starts from dir and tries finding the controller by stepping outside
// until root is reached.
func lookForController(name string, dir string) (string, error) {
	separatorIndex := strings.LastIndex(dir, "/")
	for separatorIndex > 0 {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return filepath.Join(dir, name), nil
		}
		separatorIndex = strings.LastIndex(dir, string(os.PathSeparator))
		dir = dir[0:separatorIndex]
	}

	return "", fmt.Errorf("failed to find controller %s", name)
}

// runTiltWithTimeout executes tilt using a timeout.
func runTiltWithTimeoutEnvFunc() env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		controllers := []string{"ocm-controller", "replication-controller"}
		tiltFile := ""
		tctx, cancel := context.WithTimeout(ctx, defaultTimeoutSeconds*time.Second)
		defer cancel()

		_, dir, _, _ := runtime.Caller(0)

		for _, controller := range controllers {
			path, err := lookForController(controller, dir)
			if err != nil {
				fmt.Printf("controller with name %q not found", controller)
				return ctx, err
			}

			tiltFile += fmt.Sprintf("include('%s/Tiltfile')\n", path)
		}

		temp, err := os.MkdirTemp("", "tilt-ci")
		if err != nil {
			return ctx, fmt.Errorf("failed to create temp folder: %w", err)
		}

		defer os.RemoveAll(temp)

		if err := os.WriteFile(filepath.Join(temp, "Tiltfile"), []byte(tiltFile), 0o777); err != nil {
			return ctx, fmt.Errorf("failed to create tilt file %w", err)
		}

		cmd := exec.CommandContext(tctx, "tilt", "ci")
		cmd.Dir = temp

		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("output from tilt: ", string(output))
			return ctx, err
		}

		return ctx, nil
	}
}

func TestMain(m *testing.M) {
	cfg, _ := envconf.NewFromFlags()
	testEnv = env.NewWithConfig(cfg)
	kindClusterName = envconf.RandomName("component-version-", 32)
	fmt.Println("using clustername: ", kindClusterName)
	namespace = "ocm-system"

	testEnv.Setup(
		envfuncs.CreateKindCluster(kindClusterName),
		envfuncs.CreateNamespace(namespace),
		runTiltWithTimeoutEnvFunc(),
	)

	testEnv.Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.DestroyKindCluster(kindClusterName),
	)

	os.Exit(testEnv.Run(m))
}
