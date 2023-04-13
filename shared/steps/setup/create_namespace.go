// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package setup

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// CreateNamespace creates the given namespace in the configured environment.
func CreateNamespace(name string) features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		t.Helper()

		namespace := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}

		client, err := cfg.NewClient()
		if err != nil {
			t.Fail()
		}

		if err := client.Resources().Create(ctx, &namespace); err != nil {
			t.Fail()
		}

		return ctx
	}
}
