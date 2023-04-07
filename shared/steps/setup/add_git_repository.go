// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

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

// AddGitRepository creates a git repository for the test user.
func AddGitRepository(repoName string) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()

		gclient, err := gitea.NewClient(shared.BaseURL, gitea.SetToken(shared.TestUserToken))
		if err != nil {
			t.Fatal(fmt.Errorf("failed to create gitea client: %w", err))
		}

		repo, _, err := gclient.CreateRepo(gitea.CreateRepoOption{
			AutoInit:      true,
			Name:          repoName,
			DefaultBranch: "main",
		})
		if err != nil {
			t.Fatal(fmt.Errorf("failed to create repository: %w", err))
		}

		t.Logf("successfully created repository at url %s", repo.CloneURL)

		return ctx
	}
}
