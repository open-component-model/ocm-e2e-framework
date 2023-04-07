// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package subscription

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	fconditions "github.com/fluxcd/pkg/runtime/conditions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/open-component-model/replication-controller/api/v1alpha1"

	"github.com/open-component-model/ocm-e2e-framework/shared"
	"github.com/open-component-model/ocm-e2e-framework/shared/steps/assess"
	"github.com/open-component-model/ocm-e2e-framework/shared/steps/setup"
)

func TestComponentSubscribeApply(t *testing.T) {
	t.Log("running component subscription apply")

	feature := features.New("Custom ComponentSubscription").
		Setup(setup.AddScheme(v1alpha1.AddToScheme)).
		Setup(setup.AddComponentVersion(shared.Component{
			Name:    "github.com/acme/podinfo",
			Version: "v1.0.0",
		}, "ocm-replication")).
		Setup(setup.ApplyTestData(namespace, "*")).
		Assess("check if resource was created",
			assess.ResourceWasCreated(
				"componentsubscription-sample",
				namespace,
				&v1alpha1.ComponentSubscription{},
			)).
		Assess("wait for condition to be successful", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Helper()
			t.Log("waiting for condition ready on the component version")

			client, err := cfg.NewClient()
			if err != nil {
				t.Fail()
			}

			cv := &v1alpha1.ComponentSubscription{
				ObjectMeta: metav1.ObjectMeta{Name: "componentsubscription-sample", Namespace: cfg.Namespace()},
			}

			// wait for component version to be reconciled
			err = wait.For(conditions.New(client.Resources()).ResourceMatch(cv, func(object k8s.Object) bool {
				cvObj, ok := object.(*v1alpha1.ComponentSubscription)
				if !ok {
					return false
				}

				return fconditions.IsTrue(cvObj, meta.ReadyCondition)
			}), wait.WithTimeout(time.Minute*2))

			if err != nil {
				t.Fatal(err)
			}

			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				t.Fail()
			}

			r.WithNamespace(namespace)
			if err := r.Get(ctx, "componentsubscription-sample", namespace, cv); err != nil {
				t.Fail()
			}

			t.Logf("got resource status %+v", cv.Status)

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Helper()
			t.Log("teardown")

			// remove test resources before exiting
			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}

			if err := decoder.DecodeEachFile(ctx, os.DirFS("./testdata"), "*",
				decoder.DeleteHandler(r),           // try to DELETE objects after decoding
				decoder.MutateNamespace(namespace), // inject a namespace into decoded objects, before calling DeleteHandler
			); err != nil {
				t.Fatal(err)
			}

			t.Log("teardown done")

			return ctx
		}).Feature()

	testEnv.Test(t, feature)
}
