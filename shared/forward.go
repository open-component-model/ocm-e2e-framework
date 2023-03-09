package shared

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var (
	stopChannel = make(chan struct{}, 1)
)

// ForwardRegistry forwards the in cluster oci registry to a local port.
// I need to find the registry pod.
func ForwardRegistry() env.Func {
	return func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		tctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		r, err := resources.New(config.Client().RESTConfig())
		if err != nil {
			return ctx, fmt.Errorf("failed to create resource client: %w", err)
		}
		if err := v1.AddToScheme(r.GetScheme()); err != nil {
			return ctx, fmt.Errorf("failed to add schema to resource client: %w", err)
		}

		pods := &v1.PodList{}
		if err := r.List(ctx, pods, resources.WithLabelSelector(labels.FormatLabels(map[string]string{"app": "registry"}))); err != nil {
			return ctx, fmt.Errorf("failed to list pods: %w", err)
		}
		if len(pods.Items) != 1 {
			return ctx, fmt.Errorf("invalid number of pods found for registry %d", len(pods.Items))
		}
		podName := pods.Items[0].Name
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
			[]string{fmt.Sprintf("%d:%d", 5000, 5000)},
			stopChannel,
			readyChannel,
			os.Stdout,
			os.Stderr,
		)
		if err != nil {
			return ctx, fmt.Errorf("failed to create port forwarder: %w", err)
		}
		go func() {
			fw.ForwardPorts()
		}()
		fmt.Printf("set up port-forwarding for pod with name %q under url %q\n", podName, reqURL)

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
}

func ShutdownPortForward() env.Func {
	return func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		stopChannel <- struct{}{}
		return ctx, nil
	}
}
