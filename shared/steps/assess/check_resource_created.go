// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package assess

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// Object contains information about the object to check existence on.
type Object struct {
	Name      string
	Namespace string
	Obj       k8s.Object
}

// ResourceWasCreated is an assess step to check if a given resource was created.
func ResourceWasCreated(objs ...Object) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		t.Helper()
		t.Log("check if resources are created")

		r, err := resources.New(c.Client().RESTConfig())
		if err != nil {
			t.Fail()
		}

		for _, obj := range objs {
			if err := r.Get(ctx, obj.Name, obj.Namespace, obj.Obj); err != nil {
				t.Fail()
			}

		}

		t.Log("resources successfully created")

		return ctx
	}
}
