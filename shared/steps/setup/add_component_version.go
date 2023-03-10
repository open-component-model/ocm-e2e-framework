package setup

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/open-component-model/ocm-e2e-framework/shared"
)

// AddComponentVersion defines a setup step for tests to use.
func AddComponentVersion(component shared.Component, repository string, resources ...shared.Resource) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()

		if err := shared.AddComponentVersionToRepository(component, repository, resources...); err != nil {
			t.Fatal(err)
		}

		return ctx
	}
}
