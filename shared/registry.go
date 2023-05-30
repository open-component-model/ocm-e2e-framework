// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"context"
	"fmt"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

// ResetRegistry Restarts registry deployment, waits for new deployment to exist and for old pod to exit
func ResetRegistry(name string, stopChannel chan struct{}) env.TestFunc {
	return func(ctx context.Context, config *envconf.Config, t *testing.T) (context.Context, error) {

		oldPodName, err := getPodNameForApp(ctx , config , name)
		if err != nil || oldPodName == "" {
			return ctx, fmt.Errorf("failed to get pod for the registry: %w", err)
		}

		t.Log("Old Pod Name: ", oldPodName)

		clientset, err := kubernetes.NewForConfig(config.Client().RESTConfig())
		if err != nil {
			t.Fatal(err)
			return ctx, err
		}

		deploymentsClient := clientset.AppsV1().Deployments(config.Namespace())
		oldDeployment, err := deploymentsClient.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
			return ctx, err
		}

		restartString := fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`, time.Now().Format("20060102150405"))

		_, err = deploymentsClient.Patch(ctx, name, k8sTypes.StrategicMergePatchType, []byte(restartString), metav1.PatchOptions{})
		if err != nil {
			t.Fatal(err)
			return ctx, err
		}

		newDeployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: config.Namespace(), Generation: (oldDeployment.Status.ObservedGeneration + 1)},
		}
		err = wait.For(conditions.New(config.Client().Resources()).ResourceMatch(&newDeployment, func(object k8s.Object) bool {
			_, ok := object.(*appsv1.Deployment)
			if !ok {
				return false
			}
			return true
		}), wait.WithTimeout(time.Minute*2))
		if err != nil {
			t.Fatal(err)
		}

		oldPod := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: oldPodName, Namespace: config.Namespace()},
		}
		wait.For(conditions.New(config.Client().Resources()).PodPhaseMatch(&oldPod, v1.PodSucceeded), wait.WithTimeout(time.Minute*2))

		return ctx, nil
	}
}
