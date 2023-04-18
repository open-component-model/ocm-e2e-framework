// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package setup

import (
	"context"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// ApplyTestData takes a pattern and applies that from a testdata location.
func ApplyTestData(namespace, folder, pattern string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		t.Helper()
		t.Log("applying test data...")

		r, err := resources.New(c.Client().RESTConfig())
		if err != nil {
			t.Fail()
		}

		if err := decoder.DecodeEachFile(
			ctx, os.DirFS(folder), pattern,
			decoder.CreateHandler(r),
			decoder.MutateNamespace(namespace),
		); err != nil {
			t.Fail()
		}

		t.Log("apply test data complete")

		return ctx
	}
}
