// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package gitsync

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	fconditions "github.com/fluxcd/pkg/runtime/conditions"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/open-component-model/git-controller/apis/delivery/v1alpha1"
	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"

	"github.com/open-component-model/ocm-e2e-framework/shared"
	"github.com/open-component-model/ocm-e2e-framework/shared/steps/assess"
	"github.com/open-component-model/ocm-e2e-framework/shared/steps/setup"
)

func TestSyncApply(t *testing.T) {
	t.Log("running git sync apply")

	resourceContent, err := os.ReadFile(filepath.Join("testdata_shared", "deployment.tar"))
	if err != nil {
		t.Fatal("test file not found")
	}

	setupFeature := features.New("Setup Test System").
		Setup(setup.AddScheme(v1alpha1.AddToScheme, mpasv1alpha1.AddToScheme)).
		Setup(setup.AddComponentVersions(setup.Component{
			Component: shared.Component{
				Name:    "github.com/acme/podinfo",
				Version: "v6.0.0",
			},
			Repository: "podinfo",
			ComponentVersionModifications: []shared.ComponentModification{
				shared.BlobResource(shared.Resource{
					Name: "deployment",
					Data: string(resourceContent),
					Type: resourcetypes.BLOB,
				}),
			},
		})).
		Setup(setup.AddGitRepository("test")).
		Setup(setup.ApplyTestData(namespace, "testdata_shared", "*.yaml")).
		Setup(setup.ApplyTestData(namespace, "testdata_with_normal_flow", "*.yaml")).Feature()

	verifyState := features.New("Verify System State").
		Assess("wait for git sync done condition", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Helper()
			t.Log("waiting for condition ready on the component version")
			client, err := cfg.NewClient()
			if err != nil {
				t.Fail()
			}

			gitSync := &v1alpha1.Sync{
				ObjectMeta: metav1.ObjectMeta{Name: "git-sample", Namespace: cfg.Namespace()},
			}

			// wait for component version to be reconciled
			err = wait.For(conditions.New(client.Resources()).ResourceMatch(gitSync, func(object k8s.Object) bool {
				obj, ok := object.(*v1alpha1.Sync)
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
			if err := r.Get(ctx, "git-sample", namespace, gitSync); err != nil {
				t.Fail()
			}

			t.Logf("got resource status %+v", gitSync.Status)

			return ctx
		}).Assess("check if content exists in repo",
		assess.CheckRepoFileContent(assess.File{
			Repository: "test",
			Path:       "deployment.yaml",
			Content:    "this is my deployment",
		})).Feature()

	teardownFeature := features.New("Cleanup Test System").
		Teardown(setup.DeleteGitRepository("test")).
		Teardown(setup.DeleteTestData(namespace, "testdata_shared", "*.yaml")).
		Teardown(setup.DeleteTestData(namespace, "testdata_with_normal_flow", "*.yaml")).Feature()

	testEnv.Test(t, setupFeature, verifyState, teardownFeature)
}

func TestRepositoryWithMaintainers(t *testing.T) {
	t.Log("running git repository apply")

	setupFeature := features.New("Setup Test System").
		Setup(setup.AddScheme(v1alpha1.AddToScheme, mpasv1alpha1.AddToScheme)).
		Setup(setup.ApplyTestData(namespace, "testdata_repository_only", "*.yaml")).Feature()

	verifyState := features.New("Verify System State").
		Assess("wait for repository done condition", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Helper()
			t.Log("waiting for condition ready on the repository")
			client, err := cfg.NewClient()
			if err != nil {
				t.Fail()
			}

			repository := &mpasv1alpha1.Repository{
				ObjectMeta: metav1.ObjectMeta{Name: "test-3", Namespace: cfg.Namespace()},
			}

			// wait for component version to be reconciled
			err = wait.For(conditions.New(client.Resources()).ResourceMatch(repository, func(object k8s.Object) bool {
				obj, ok := object.(*mpasv1alpha1.Repository)
				if !ok {
					return false
				}

				return fconditions.IsTrue(obj, meta.ReadyCondition)
			}), wait.WithTimeout(time.Minute*1))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("check if files are in the repo",
			assess.CheckRepoFileContent(assess.File{
				Repository: "test-3",
				Path:       "CODEOWNERS",
				Content:    "@e2e-tester",
			}, assess.File{
				Repository: "test-3",
				Path:       "products/.keep",
			}, assess.File{
				Repository: "test-3",
				Path:       "targets/.keep", Content: "",
			}, assess.File{
				Repository: "test-3",
				Path:       "subscriptions/.keep",
			}, assess.File{
				Repository: "test-3",
				Path:       "generators/.keep",
			})).Feature()

	teardownFeature := features.New("Cleanup Test System").
		Teardown(setup.DeleteGitRepository("test-3")).
		Teardown(setup.DeleteTestData(namespace, "testdata_repository_only", "*.yaml")).Feature()

	testEnv.Test(t, setupFeature, verifyState, teardownFeature)
}

func TestSyncApplyWithPullRequest(t *testing.T) {
	t.Log("running git sync apply")

	resourceContent, err := os.ReadFile(filepath.Join("testdata_shared", "deployment.tar"))
	if err != nil {
		t.Fatal("test file not found")
	}

	setupFeature := features.New("Apply Sync with Pull Request").
		Setup(setup.AddScheme(v1alpha1.AddToScheme, mpasv1alpha1.AddToScheme)).
		Setup(setup.AddComponentVersions(setup.Component{
			Component: shared.Component{
				Name:    "github.com/acme/podinfo",
				Version: "v6.0.0",
			},
			Repository: "podinfo",
			ComponentVersionModifications: []shared.ComponentModification{
				shared.BlobResource(shared.Resource{
					Name: "deployment",
					Data: string(resourceContent),
					Type: resourcetypes.BLOB,
				}),
			},
		})).
		Setup(setup.AddGitRepository("test-2")).
		Setup(setup.ApplyTestData(namespace, "testdata_shared", "*.yaml")).
		Setup(setup.ApplyTestData(namespace, "testdata_with_pull_request", "*.yaml")).Feature()

	verifyState := features.New("Verify System State").
		Assess("wait for git sync done condition", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Helper()
			t.Log("waiting for condition ready on the component version")
			client, err := cfg.NewClient()
			if err != nil {
				t.Fail()
			}

			gitSync := &v1alpha1.Sync{
				ObjectMeta: metav1.ObjectMeta{Name: "git-sample-with-pull-request", Namespace: cfg.Namespace()},
			}

			// wait for component version to be reconciled
			err = wait.For(conditions.New(client.Resources()).ResourceMatch(gitSync, func(object k8s.Object) bool {
				obj, ok := object.(*v1alpha1.Sync)
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
			if err := r.Get(ctx, "git-sample-with-pull-request", namespace, gitSync); err != nil {
				t.Fail()
			}

			t.Logf("got resource status %+v", gitSync.Status)

			return ctx
		}).Assess("check if content exists in repo", assess.CheckIfPullRequestExists("test-2", 1)).Feature()

	teardownFeature := features.New("Teardown Test System").
		Teardown(setup.DeleteGitRepository("test-2")).
		Teardown(setup.DeleteTestData(namespace, "testdata_shared", "*.yaml")).
		Teardown(setup.DeleteTestData(namespace, "testdata_with_pull_request", "*.yaml")).Feature()

	testEnv.Test(t, setupFeature, verifyState, teardownFeature)
}
