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

// MergePullRequest merges a PR
func MergePullRequest(repoName string, prNumber int) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()
		gclient, err := gitea.NewClient(shared.BaseURL, gitea.SetToken(shared.TestUserToken))
		if err != nil {
			t.Fatal(fmt.Errorf("failed to create gitea client: %w", err))
		}
		httpResp, response, err := gclient.MergePullRequest(shared.Owner, repoName, int64(prNumber), gitea.MergePullRequestOption{
			Style:   gitea.MergeStyleMerge,
			Title:   "PR",
			Message: "Auto Merge PR with after successful Validation Check",
		})
		fmt.Println("http  ", httpResp)
		if err != nil {
			t.Fatal(fmt.Errorf("failed to Merge expected PR %s for Repository %s with error: %w", prNumber, repoName, err))
		}
		if !httpResp {
			fmt.Println("http  ", response.Body, response.Status)
		}

		prs, resp, err := gclient.ListRepoPullRequests(shared.Owner, repoName, gitea.ListPullRequestsOptions{
			ListOptions: gitea.ListOptions{},
			State:       gitea.StateAll,
		})
		if err != nil {
			t.Fatal(fmt.Errorf("failed to Merge expected PR %s for Repository %s with error: %w", prNumber, repoName, err))
		}
		for _, obj := range prs {
			fmt.Println(
				obj.ID,
				obj.URL,
				obj.Index,
				obj.Title,
				obj.Body,
				obj.IsLocked,
				obj.Comments,
				obj.HTMLURL,
				obj.DiffURL,
				obj.PatchURL,
				obj.Mergeable,
				obj.HasMerged,
				obj.MergeBase,
				resp.Status)
		}
		gclient.SetSudo(shared.Owner)
		st, response2, err := gclient.MergePullRequest(shared.Owner, repoName, int64(prNumber), gitea.MergePullRequestOption{
			Style:   gitea.MergeStyleRebaseMerge,
			Title:   "PR",
			Message: "Auto Merge PR with after successful Validation Check",
		})
		fmt.Println(st, response2.Status)
		st3, response3, err := gclient.MergePullRequest(shared.Owner, repoName, int64(prNumber), gitea.MergePullRequestOption{
			Style:   gitea.MergeStyleSquash,
			Title:   "PR",
			Message: "Auto Merge PR with after successful Validation Check",
		})
		fmt.Println(st3, response3.Status, response3.Body)
		return ctx
	}
}