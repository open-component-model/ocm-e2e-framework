// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package assess

import (
	"context"
	"fmt"
	"testing"

	"code.gitea.io/sdk/gitea"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/open-component-model/ocm-e2e-framework/shared"
)

// CheckIfPullRequestExists adds a check to verify that a repository exists.
func CheckIfPullRequestExists(repoName string, number int) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()

		gclient, err := gitea.NewClient(shared.BaseURL, gitea.SetToken(shared.TestUserToken))
		if err != nil {
			t.Fatal(fmt.Errorf("failed to create gitea client: %w", err))
		}

		_, _, err = gclient.GetRepo(shared.Owner, repoName)
		if err != nil {
			t.Fatal(fmt.Errorf("failed to find expected repository: %w", err))
		}

		_, _, err = gclient.GetPullRequest(shared.Owner, repoName, int64(number))
		if err != nil {
			t.Fatal(fmt.Errorf("pull request with number %d not found for repo %s: %w", number, repoName, err))
		}

		return ctx
	}
}
