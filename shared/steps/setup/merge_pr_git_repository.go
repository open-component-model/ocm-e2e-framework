// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"testing"

	"github.com/open-component-model/ocm-e2e-framework/shared"
)

type Payload struct {
	Do                     string `json:"Do"`
	MergeMessageField      string `json:"MergeMessageField"`
	MergeTitleField        string `json:"MergeTitleField"`
	MergeWhenChecksSucceed bool   `json:"merge_when_checks_succeed"`
	ForceMerge             bool   `json:"force_merge"`
}

// MergePullRequest merges a PR
func MergePullRequest(repoName string, prNumber int) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()

		data := Payload{
			Do:                     "squash",
			MergeMessageField:      "PR Merge",
			MergeTitleField:        "Auto Merge PR with after successful Validation Check",
			MergeWhenChecksSucceed: false,
			ForceMerge:             true,
		}
		payloadBytes, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}
		body := bytes.NewReader(payloadBytes)
		url := fmt.Sprintf("%s/api/v1/repos/%s/%s/pulls/%d/merge", shared.BaseURL, shared.Owner, repoName, prNumber)
		req, err := http.NewRequest("POST", url, body)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("token %s", shared.TestUserToken))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		t.Log(fmt.Sprintf("Request Status: %d %s ", resp.StatusCode, readResponse(resp.Body)))

		if resp.StatusCode != 200 {
			t.Fatal(fmt.Errorf("failed to Merge expected PR %d for Repository %s with error: %s", prNumber, repoName, readResponse(resp.Body)))
		}

		return ctx
	}
}

func readResponse(response io.Reader) string {
	content, err := io.ReadAll(response)
	if err != nil {
		println(err)
	}
	return string(content)
}