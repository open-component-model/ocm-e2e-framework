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

// CheckRepoFileContent adds a check to verify that content of a pushed file has the expected content.
func CheckRepoFileContent(repoName, filename, expectedContent string) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()

		gclient, err := gitea.NewClient(shared.BaseURL, gitea.SetToken(shared.TestUserToken))
		if err != nil {
			t.Fatal(fmt.Errorf("failed to create gitea client: %w", err))
		}

		content, _, err := gclient.GetFile(shared.Owner, repoName, "main", filename)
		if err != nil {
			t.Fatal(fmt.Errorf("failed to find expected file: %w", err))
		}

		if expectedContent != string(content) {
			t.Fatalf("expected content did not equal actual: %s", string(content))
		}

		return ctx
	}
}
