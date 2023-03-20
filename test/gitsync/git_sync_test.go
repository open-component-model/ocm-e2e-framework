// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package gitsync

import (
	"context"
	"testing"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	fconditions "github.com/fluxcd/pkg/runtime/conditions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/open-component-model/git-sync-controller/api/v1alpha1"

	"github.com/open-component-model/ocm-e2e-framework/shared"
	"github.com/open-component-model/ocm-e2e-framework/shared/steps/setup"
)

func TestGitSyncApply(t *testing.T) {
	t.Log("running git sync apply")

	feature := features.New("Custom GitSync").
		Setup(setup.AddSchemeAndNamespace(v1alpha1.AddToScheme, namespace)).
		Setup(setup.AddComponentVersion(shared.Component{
			Name:    "github.com/acme/podinfo",
			Version: "v6.0.0",
		}, "ocm-podinfo", shared.Resource{
			Name: "deployment",
			Data: "this is my deployment",
		})).
		Setup(setup.AddGitRepository("test")).
		Setup(setup.ApplyTestData(namespace, "*")).
		Assess("wait for git sync done condition", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Helper()
			t.Log("waiting for condition ready on the component version")
			client, err := cfg.NewClient()
			if err != nil {
				t.Fail()
			}

			gitSync := &v1alpha1.GitSync{
				ObjectMeta: metav1.ObjectMeta{Name: "git-sync-sample", Namespace: cfg.Namespace()},
			}

			// wait for component version to be reconciled
			err = wait.For(conditions.New(client.Resources()).ResourceMatch(gitSync, func(object k8s.Object) bool {
				obj, ok := object.(*v1alpha1.GitSync)
				if !ok {
					return false
				}

				return fconditions.IsTrue(obj, meta.ReadyCondition)
			}), wait.WithTimeout(time.Minute*1))

			if err != nil {
				t.Fatal(err)
			}

			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				t.Fail()
			}

			r.WithNamespace(namespace)
			if err := r.Get(ctx, "git-sync-sample", namespace, gitSync); err != nil {
				t.Fail()
			}

			t.Logf("got resource status %+v", gitSync.Status)

			return ctx
		}).Teardown(setup.DeleteGitRepository("test")).Feature()

	testEnv.Test(t, feature)
}
