package component_version //nolint:stylecheck

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// runTiltWithTimeout executes tilt using a timeout.
func runTiltWithTimeout() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutSeconds*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "tilt", "ci")

	_, file, _, _ := runtime.Caller(0)
	// executable, err := os.Executable()
	// if err != nil {
	// 	return err
	// }

	cmd.Dir = filepath.Join(file, "..", "..", "..")

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("output from tilt: ", string(output))
		return err
	}
	return nil
}

func TestMain(m *testing.M) {
	cfg, _ := envconf.NewFromFlags()
	testEnv = env.NewWithConfig(cfg)
	kindClusterName = envconf.RandomName("component-version-", 16)
	namespace = envconf.RandomName("ocm-system", 10)

	testEnv.Setup(
		envfuncs.CreateKindCluster(kindClusterName),
		envfuncs.CreateNamespace(namespace),
	)

	testEnv.Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.DestroyKindCluster(kindClusterName),
	)

	if err := runTiltWithTimeout(); err != nil {
		fmt.Println("failed to execute tilt to set up environment: ", err)
		os.Exit(1)
	}

	os.Exit(testEnv.Run(m))
}
