package teardown

import (
	"context"
	"fmt"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/open-component-model/ocm-e2e-framework/shared"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// DumpRepositoryContent dumps all the files in a repository.
func DumpRepositoryContent(owner, repo string) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()

		gclient, err := gitea.NewClient(shared.BaseURL, gitea.SetToken(shared.TestUserToken))
		if err != nil {
			t.Fatal(fmt.Errorf("failed to create gitea client: %w", err))
		}

		r, _, err := gclient.GetTrees(owner, repo, gitea.ListTreeOptions{
			ListOptions: gitea.ListOptions{
				PageSize: 100,
				Page:     0,
			},
			Ref:       "main",
			Recursive: true,
		})
		if err != nil {
			t.Fatal(fmt.Errorf("failed to find repo for %s/%s: %w", owner, repo, err))
		}

		for _, entry := range r.Entries {
			t.Logf("Type: %s | Path: %s", entry.Type, entry.Path)
		}

		return ctx
	}
}
