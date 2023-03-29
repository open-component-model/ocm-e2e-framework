// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

// internal values.
var (
	//go:embed private_git_key/id_25519
	privateTestKey string
	//go:embed gitea/gitea_deployment.yaml
	giteaDeployment string
	timeout         = time.Minute * 5

	// TestUserToken is the token generated for API access on the created test user.
	TestUserToken = "91efc1d52e9d6069729f373c2cad057da974f11e" //nolint:gosec // this is a test key
	BaseURL       = "http://127.0.0.1:3000"
	Owner         = "e2e-tester"
)

// StartGitServer installs a Gitea Git server into the cluster using the deployment configuration files provided
// under ./gitea folder.
func StartGitServer(namespace string) env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		r, err := resources.New(c.Client().RESTConfig())
		if err != nil {
			return ctx, fmt.Errorf("failed to create rest client: %w", err)
		}

		location, err := createLocalizedDeployment(namespace)
		if err != nil {
			return ctx, fmt.Errorf("failed to create localized deployment: %w", err)
		}

		defer os.RemoveAll(location)

		if err := decoder.DecodeEachFile(
			ctx, os.DirFS(location), "*",
			decoder.CreateHandler(r),
			decoder.MutateNamespace(namespace),
		); err != nil {
			return ctx, fmt.Errorf("failed to apply gitea configuration files: %w", err)
		}

		client, err := c.NewClient()
		if err != nil {
			return ctx, fmt.Errorf("failed to create new client: %w", err)
		}

		deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "gitea", Namespace: namespace}}

		if err = wait.For(
			conditions.New(client.Resources()).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, corev1.ConditionTrue),
			wait.WithTimeout(timeout),
		); err != nil {
			return ctx, fmt.Errorf("gitea deployment didn't become ready: %w", err)
		}

		return ctx, nil
	}
}

// RemoveGitServer removes the previously installed Gitea server.
func RemoveGitServer(namespace string) env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		r, err := resources.New(c.Client().RESTConfig())
		if err != nil {
			return ctx, fmt.Errorf("failed to create rest client: %w", err)
		}

		location, err := createLocalizedDeployment(namespace)
		if err != nil {
			return ctx, fmt.Errorf("failed to create localized deployment: %w", err)
		}

		defer os.RemoveAll(location)

		if err := decoder.DecodeEachFile(
			ctx, os.DirFS(location), "*",
			decoder.DeleteHandler(r),
			decoder.MutateNamespace(namespace),
		); err != nil {
			return ctx, fmt.Errorf("failed to apply gitea configuration files: %w", err)
		}

		return ctx, nil
	}
}

// createLocalizedDeployment takes the generated namespace and updates the deployment, configmap and service.
func createLocalizedDeployment(namespace string) (string, error) {
	dir, err := os.MkdirTemp("", "localized-gitea")
	if err != nil {
		return "", fmt.Errorf("failed to create localized deployment: %w", err)
	}

	deployment := strings.ReplaceAll(giteaDeployment, "<NAMESPACE>", namespace)

	var filePermission os.FileMode = 0o600
	if err := os.WriteFile(filepath.Join(dir, "deployment.yaml"), []byte(deployment), filePermission); err != nil {
		return "", fmt.Errorf("failed to write out deployment file: %w", err)
	}

	return dir, nil
}
