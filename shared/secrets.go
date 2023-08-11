// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"context"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

// Creates a secret
func CreateSecret(name string, data map[string][]byte, stringData map[string]string, namespace string) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		clientset, err := kubernetes.NewForConfig(config.Client().RESTConfig())
		if err != nil {
			t.Fatal(err)
			return ctx
		}
		if len(namespace) == 0 {
			namespace = config.Namespace()
		}
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Data:       data,
			StringData: stringData,
		}

		_, err = clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			t.Fatal(err)
			return ctx
		}

		newSecret := v1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		}
		err = wait.For(conditions.New(config.Client().Resources()).ResourceMatch(&newSecret, func(object k8s.Object) bool {
			_, ok := object.(*v1.Secret)
			if !ok {
				return false
			}
			return true
		}), wait.WithTimeout(time.Minute*2))

		if err != nil {
			t.Fatal(err)
		}
		return ctx
	}
}

// DeleteSecret Deletes a secret
func DeleteSecret(name string) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		clientset, err := kubernetes.NewForConfig(config.Client().RESTConfig())
		if err != nil {
			t.Fatal(err)
			return ctx
		}

		err = clientset.CoreV1().Secrets(config.Namespace()).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			t.Fatal(err)
			return ctx
		}

		secret := v1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: config.Namespace()},
		}
		err = wait.For(conditions.New(config.Client().Resources()).ResourceDeleted(&secret), wait.WithTimeout(time.Minute*2))
		if err != nil {
			t.Fatal(err)
		}

		return ctx
	}
}