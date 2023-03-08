package component_version //nolint:stylecheck

import (
	"context"
	"os"
	"testing"

	"k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/open-component-model/ocm-controller/api/v1alpha1"
)

func TestComponentVersionApply(t *testing.T) {
	feature := features.New("Custom ComponentVersion").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
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
			return ctx
		}).
		Assess("Check If Resource created", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fail()
			}
			r.WithNamespace(namespace)
			if err := v1alpha1.AddToScheme(r.GetScheme()); err != nil {
				t.Fail()
			}
			ct := &v1alpha1.ComponentVersion{}
			err = r.Get(ctx, "podinfo", namespace, ct)
			if err != nil {
				t.Fail()
			}
			klog.InfoS("CR Details", "cr", ct)
			return ctx
		}).Feature()

	testEnv.Test(t, feature)
}
