// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package setup

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// AddScheme provides a setup function to add the scheme to the client.
// Consider renaming this to create a client and pass it over via the context.
func AddScheme(addSchemeFuncs ...func(scheme *runtime.Scheme) error) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()

		r, err := resources.New(config.Client().RESTConfig())
		if err != nil {
			t.Fail()
		}

		for _, f := range addSchemeFuncs {
			if err := f(r.GetScheme()); err != nil {
				t.Fail()
			}
		}

		return ctx
	}
}
