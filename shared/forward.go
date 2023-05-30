// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	defaultPortForwardReadyWaitTime = 10
	timeoutDuration                 = time.Minute * 2
)

// PortForward forwards the given port for the given pod name.
func PortForward(port int, stopChannel chan struct{}, podName string, ctx context.Context, config *envconf.Config) (context.Context, error) {
	transport, upgrader, err := spdy.RoundTripperFor(config.Client().RESTConfig())
	if err != nil {
		return ctx, fmt.Errorf("failed to process round tripper: %w", err)
	}
	readyChannel := make(chan struct{})

	reqURL, err := url.Parse(
		fmt.Sprintf(
			"%s/api/v1/namespaces/%s/pods/%s/portforward",
			config.Client().RESTConfig().Host,
			config.Namespace(),
			podName,
		),
	)
	if err != nil {
		return ctx, fmt.Errorf("could not build URL for portforward: %w", err)
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", reqURL)

	fw, err := portforward.NewOnAddresses(
		dialer,
		[]string{"127.0.0.1"},
		[]string{fmt.Sprintf("%d:%d", port, port)},
		stopChannel,
		readyChannel,
		os.Stdout,
		os.Stderr,
	)
	if err != nil {
		return ctx, fmt.Errorf("failed to create port forwarder: %w", err)
	}

	go func() {
		if err := fw.ForwardPorts(); err != nil {
			panic(err)
		}
	}()

	tctx, cancel := context.WithTimeout(ctx, defaultPortForwardReadyWaitTime*time.Second)
	defer cancel()

	select {
	case <-readyChannel:
		break
	case <-tctx.Done():
		return ctx, fmt.Errorf("failed to start port forwarder: %w", ctx.Err())
	}

	ports, err := fw.GetPorts()
	if err != nil {
		return ctx, fmt.Errorf("failed to get ports: %w", err)
	}

	if len(ports) != 1 {
		return ctx, fmt.Errorf("failed to get expected ports: %+v", ports)
	}

	return ctx, nil
}

// ForwardPortForAppName port forwards at test setup phase
func ForwardPortForAppName(name string, port int, stopChannel chan struct{}) env.Func {
	return func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		podName, err := getPodNameForApp(ctx, config, name)
		if err != nil || podName == "" {
			return ctx, fmt.Errorf("failed to get pod for the registry: %w", err)
		}
		return PortForward(port, stopChannel, podName, ctx, config)
	}
}

// ForwardPortForAppNameAfterTest port forwards after each test cleanup
func ForwardPortForAppNameAfterTest(name string, port int, stopChannel chan struct{}) env.TestFunc {
	return func(ctx context.Context, config *envconf.Config, t *testing.T) (context.Context, error) {
		podName, err := getPodNameForAppAfterTest(ctx, config, name)
		if err != nil || podName == "" {
			return ctx, fmt.Errorf("failed to get pod for the registry: %w", err)
		}
		t.Log("\nForwarding port for Pod: "+podName, "\n")
		return PortForward(port, stopChannel, podName, ctx, config)
	}
}

// getPodNameForApp returns the name of the pod the registry is running in for port-forwarding requests to.
func getPodNameForApp(ctx context.Context, config *envconf.Config, name string) (string, error) {
	r, err := resources.New(config.Client().RESTConfig())
	if err != nil {
		return "", fmt.Errorf("failed to create resource client: %w", err)
	}

	if err := v1.AddToScheme(r.GetScheme()); err != nil {
		return "", fmt.Errorf("failed to add schema to resource client: %w", err)
	}

	pods := &v1.PodList{}
	if err := r.List(ctx, pods, resources.WithLabelSelector(
		labels.FormatLabels(map[string]string{"app": name})),
	); err != nil {
		return "", fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) != 1 {
		return "", fmt.Errorf("invalid number of pods found for registry %d", len(pods.Items))
	}

	return pods.Items[0].Name, nil
}

// getPodNameForAppAfterTest Waits for new Registry Pod to be Running && Ready to accept traffic
func getPodNameForAppAfterTest(ctx context.Context, config *envconf.Config, name string) (string, error) {
	r, err := resources.New(config.Client().RESTConfig())
	if err != nil {
		return "", fmt.Errorf("failed to create resource client: %w", err)
	}

	if err := v1.AddToScheme(r.GetScheme()); err != nil {
		return "", fmt.Errorf("failed to add schema to resource client: %w", err)
	}

	pods := &v1.PodList{}
	if err := r.List(ctx, pods, resources.WithLabelSelector(labels.FormatLabels(map[string]string{"app": name}))); err != nil {
		return "", fmt.Errorf("failed to list pods: %w", err)
	}
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {

			podObj := v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: pod.Name, Namespace: config.Namespace()},
			}
			err = wait.For(conditions.New(config.Client().Resources()).PodConditionMatch(&podObj, v1.PodConditionType(v1.PodReady), v1.ConditionTrue), wait.WithTimeout(timeoutDuration))
			if err != nil {
				return "", fmt.Errorf(err.Error())
			}
			return pod.Name, nil
		}
	}

	return "", nil
}

// ShutdownPortForward sends a signal to the stop channel.
func ShutdownPortForward(stopChannel chan struct{}) env.Func {
	return func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		stopChannel <- struct{}{}

		return ctx, nil
	}
}

// ShutdownPortForwardAfterTest sends a signal to the stop channel.
func ShutdownPortForwardAfterTest(stopChannel chan struct{}) env.TestFunc {
	return func(ctx context.Context, config *envconf.Config, t *testing.T) (context.Context, error) {
		stopChannel <- struct{}{}

		return ctx, nil
	}
}
