package component_version //nolint:stylecheck

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

	"github.com/open-component-model/ocm-controller/api/v1alpha1"
)

func TestComponentVersionApply(t *testing.T) {
	t.Log("running component version apply")
	feature := features.New("Custom ComponentVersion").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			t.Log("in setup phase")
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fail()
			}
			if err := v1alpha1.AddToScheme(r.GetScheme()); err != nil {
				t.Fail()
			}
			r.WithNamespace(namespace)
			if err := decoder.DecodeEachFile(
				ctx, os.DirFS("./testdata"), "*",
				decoder.CreateHandler(r),
				decoder.MutateNamespace(namespace),
			); err != nil {
				t.Fail()
			}
			t.Log("set up is done, component version should have been applied")

			return ctx
		}).
		Assess("Check If Resource created", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			t.Log("check if resources are created")
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fail()
			}
			r.WithNamespace(namespace)
			if err := v1alpha1.AddToScheme(r.GetScheme()); err != nil {
				t.Fail()
			}

			ct := &v1alpha1.ComponentVersion{}
			if err := r.Get(ctx, "podinfo", namespace, ct); err != nil {
				t.Fail()
			}

			t.Log("resource successfully created")

			return ctx
		}).
		Assess("wait for condition to be successful", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Log("waiting for condition ready on the component version")

			client, err := cfg.NewClient()
			if err != nil {
				t.Fail()
			}

			cv := &v1alpha1.ComponentVersion{
				ObjectMeta: metav1.ObjectMeta{Name: "podinfo", Namespace: cfg.Namespace()},
			}

			// wait for component version to be reconciled
			err = wait.For(conditions.New(client.Resources()).ResourceMatch(cv, func(object k8s.Object) bool {
				cvObj := object.(*v1alpha1.ComponentVersion)
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
			if err := r.Get(ctx, "podinfo", namespace, cv); err != nil {
				t.Fail()
			}

			t.Logf("got resource status %+v", cv.Status)
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
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
