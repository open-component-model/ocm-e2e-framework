// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package setup

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"code.gitea.io/sdk/gitea"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/open-component-model/ocm-e2e-framework/shared"
)

// AddFileToGitRepository adds a file to a git repository
func AddFileToGitRepository(repoName, sourceFilepath, destFilepath string) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()

		gclient, err := gitea.NewClient(shared.BaseURL, gitea.SetToken(shared.TestUserToken))
		if err != nil {
			t.Fatal(fmt.Errorf("failed to create gitea client: %w", err))
		}

		data, err := os.ReadFile(filepath.Join("./testdata", sourceFilepath))
		if err != nil {
			return nil
		}

		_, _, err = gclient.CreateFile(shared.Owner, repoName, destFilepath, gitea.CreateFileOptions{
			Content: base64.StdEncoding.EncodeToString(data),
		})
		if err != nil {
			t.Fatal(fmt.Errorf("failed to add file to repository: %w", err))
		}

		t.Logf("successfully added %s to repository %s", destFilepath, repoName)

		return ctx
	}
}
