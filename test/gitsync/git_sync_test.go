// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package gitsync

import (
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/open-component-model/git-sync-controller/api/v1alpha1"

	"github.com/open-component-model/ocm-e2e-framework/shared"
	"github.com/open-component-model/ocm-e2e-framework/shared/steps/setup"
)

func TestGitSyncApply(t *testing.T) {
	t.Log("running git sync apply")

	feature := features.New("Custom GitSync").
		Setup(setup.AddSchemeAndNamespace(v1alpha1.AddToScheme, namespace)).
		Setup(setup.AddComponentVersion(shared.Component{
			Name:    "github.com/acme/podinfo",
			Version: "v6.0.0",
		}, "ocm-podinfo", shared.Resource{
			Name: "deployment",
			Data: "this is my deployment",
		})).
		Setup(setup.AddGitRepository("test")).
		Setup(setup.ApplyTestData(namespace, "*")).
		Teardown(setup.DeleteGitRepository("test")).Feature()

	testEnv.Test(t, feature)
}
