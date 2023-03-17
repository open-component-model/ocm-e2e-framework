package setup

import (
	"context"
	"fmt"
	"testing"

	"code.gitea.io/sdk/gitea"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/open-component-model/ocm-e2e-framework/shared"
)

// DeleteGitRepository deletes a git repository for the test user.
func DeleteGitRepository(repoName string) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()

		gclient, err := gitea.NewClient(shared.BaseURL, gitea.SetToken(shared.TestUserToken))
		if err != nil {
			t.Fatal(fmt.Errorf("failed to create gitea client: %w", err))
		}

		if _, err := gclient.DeleteRepo("e2e-tester", repoName); err != nil {
			t.Fatal(fmt.Errorf("failed to create repository: %w", err))
		}

		return ctx
	}
}
