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

// CheckFileInRepoExists adds a check to verify that file(s) exists in a repository
func CheckFileInRepoExists(files ...File) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()

		gclient, err := gitea.NewClient(shared.BaseURL, gitea.SetToken(shared.TestUserToken))
		if err != nil {
			t.Fatal(fmt.Errorf("failed to create gitea client: %w", err))
		}

		for _, file := range files {
			fmt.Println(fmt.Sprintf("shared.Owner %s file.Repository %s file.Path %s", shared.Owner, file.Repository, file.Path))
			_, _, err := gclient.GetFile(shared.Owner, file.Repository, "main", file.Path)
			if err != nil {
				t.Fatal(fmt.Errorf("failed to find expected file %s/%s with error: %w", file.Repository, file.Path, err))
			}
		}
		return ctx
	}
}
