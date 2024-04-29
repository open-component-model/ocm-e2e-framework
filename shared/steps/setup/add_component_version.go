// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package setup

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/open-component-model/ocm-e2e-framework/shared"
)

// Component contains information about a component to add.
type Component struct {
	Component                     shared.Component
	Scheme                        string
	ComponentVersionModifications []shared.ComponentModification
}

// AddComponentVersions defines a list of component versions to add.
func AddComponentVersions(components ...Component) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()

		for _, c := range components {
			scheme := "https"
			if c.Scheme != "" {
				scheme = c.Scheme
			}
			t.Logf("c.Component: %s c.Component.Version %s c.Scheme: %s ", c.Component.Name, c.Component.Version, c.Scheme)

			if err := shared.AddComponentVersionToRepository(c.Component, scheme, c.ComponentVersionModifications...); err != nil {
				t.Fatal(err)
			}
		}

		return ctx
	}
}
