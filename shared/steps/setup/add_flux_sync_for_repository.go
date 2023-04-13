// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package setup

import (
	"context"
	"fmt"
	"testing"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/open-component-model/ocm-e2e-framework/shared"
)

func AddFluxSyncForRepo(name, path, giteaNamespace string) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()

		r, err := resources.New(config.Client().RESTConfig())
		if err != nil {
			t.Fail()
		}

		tokenSecret := corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: "flux-system",
			},
			StringData: map[string]string{
				"bearerToken": shared.TestUserToken,
			},
		}

		if err := r.Create(ctx, &tokenSecret); err != nil {
			t.Error(err)

			return ctx
		}

		t.Logf("Created token secret flux-system/%s", name)

		gitRepo := sourcev1.GitRepository{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: "flux-system",
			},
			Spec: sourcev1.GitRepositorySpec{
				URL: fmt.Sprintf("http://gitea.%s:3000/%s/%s", giteaNamespace, shared.Owner, name),
				SecretRef: &meta.LocalObjectReference{
					Name: tokenSecret.GetName(),
				},
				Interval: v1.Duration{
					Duration: time.Hour,
				},
				Reference: &sourcev1.GitRepositoryRef{
					Branch: "main",
				},
			},
		}

		if err := r.Create(ctx, &gitRepo); err != nil {
			t.Error(err)

			return ctx
		}

		t.Logf("Created git repository flux-system/%s", name)

		kust := kustomizev1.Kustomization{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: "flux-system",
			},
			Spec: kustomizev1.KustomizationSpec{
				Interval: v1.Duration{
					Duration: time.Hour,
				},
				Path:  path,
				Prune: true,
				SourceRef: kustomizev1.CrossNamespaceSourceReference{
					Kind:      "GitRepository",
					Name:      name,
					Namespace: "flux-system",
				},
			},
		}

		if err := r.Create(ctx, &kust); err != nil {
			t.Error(err)

			return ctx
		}

		t.Logf("Created kustomization flux-system/%s", name)

		return ctx
	}
}
