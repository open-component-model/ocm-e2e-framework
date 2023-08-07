// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	notifv1 "github.com/fluxcd/notification-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"testing"
	"time"

	"github.com/open-component-model/ocm-e2e-framework/shared"
)

type PayloadHook struct {
	Active              bool              `json:"active"`
	AuthorizationHeader string            `json:"authorization_header"`
	BranchFilter        string            `json:"branch_filter"`
	Config              map[string]string `json:"config"`
	Events              []string          `json:"events"`
	Type                string            `json:"type"`
}

type ConfigHook struct {
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
	HttpMethod  string `json:"http_method"`
	Secret      string `json:"secret"`
}

// CreateWebhookAPI creates a webhook for a gitea git repository
func CreateWebhookAPI(repoName, token string) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()
		client, err := config.NewClient()
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("checking if Reciever %s exists...", repoName)

		receiver := &notifv1.Receiver{ObjectMeta: metav1.ObjectMeta{
			Name:      repoName,
			Namespace: repoName,
		}}
		err = wait.For(conditions.New(client.Resources()).ResourceMatch(receiver, func(object k8s.Object) bool {
			_, ok := object.(*notifv1.Receiver)
			if !ok {
				return false
			}
			return true
		}), wait.WithTimeout(time.Minute*1))
		if err != nil {
			t.Fatal(err)
		}

		fluxHook := receiver.Status.WebhookPath
		t.Logf("CreateWebhookSDK fluxHook %s ..", fluxHook)

		configFormat := map[string]string{}
		configH, err := json.Marshal(ConfigHook{
			ContentType: "json",
			URL:         "http://webhook-receiver.flux-system" + fluxHook,
			HttpMethod:  "post",
			Secret:      token,
		})
		if err != nil {
			t.Fatal(err)
		}
		err = json.Unmarshal(configH, &configFormat)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("CreateWebhookAPI config %s", configFormat)

		data := PayloadHook{
			Active:       true,
			BranchFilter: "main",
			Config:       configFormat,
			Events:       []string{"push", "ping"},
			Type:         "gitea",
		}
		payloadBytes, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		body := bytes.NewReader(payloadBytes)
		url := fmt.Sprintf("%s/api/v1/repos/%s/%s/hooks", shared.BaseURL, shared.Owner, repoName)
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

		if resp.StatusCode != 201 {
			t.Fatal(fmt.Errorf("failed to Create Hook for Repository %s with error: %s", repoName, readResponse(resp.Body)))
		}

		return ctx
	}
}