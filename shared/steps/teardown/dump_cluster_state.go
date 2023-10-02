package teardown

import (
	"context"
	"fmt"
	"io"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// DumpClusterState dumps the status of pods and logs of given controllers.
func DumpClusterState(controllers ...string) features.Func {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()
		// Dump controller logs
		for _, controller := range controllers {
			if err := dumpLogs(ctx, t, config, controller); err != nil {
				t.Fatalf("failed to dump logs for controller %s, %s", controller, err)
			}
		}

		// Dump list of pods in namespace
		namespaces := &v1.NamespaceList{}

		client, err := config.NewClient()
		if err != nil {
			t.Fatal(err)
		}

		if err := client.Resources().List(ctx, namespaces); err != nil {
			t.Fatal(err)
		}

		for _, ns := range namespaces.Items {
			pods := &v1.PodList{}
			if err := client.Resources(ns.Name).List(ctx, pods); err != nil {
				t.Fatal(fmt.Errorf("failed to list pods in namespace %s: %w", ns.Name, err))
			}

			for _, pod := range pods.Items {
				t.Logf("Name: %s | Namespace: %s | Status: %s", pod.Name, pod.Namespace, pod.Status.String())
			}
		}

		return ctx
	}
}

func dumpLogs(ctx context.Context, t *testing.T, config *envconf.Config, controller string) error {
	t.Helper()

	client, err := config.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	pods := &v1.PodList{}
	if err := client.Resources(config.Namespace()).List(ctx, pods, resources.WithLabelSelector(
		labels.FormatLabels(map[string]string{"app": controller})),
	); err != nil {
		t.Fatal(fmt.Errorf("failed to list pods: %w", err))
	}

	if len(pods.Items) != 1 {
		t.Fatal(fmt.Errorf("invalid number of pods found for registry %d", len(pods.Items)))
	}

	pod := pods.Items[0]

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config.Client().RESTConfig())
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	podReq := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{})

	reader, err := podReq.Stream(ctx)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to fetch pod logs: %w", err))
	}

	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to read log: %w", err))
	}

	t.Logf("Pod: %s | Log: %s", controller, string(content))

	return nil
}
