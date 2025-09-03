package orka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	orkav1 "github.com/macstadium/packer-plugin-macstadium-orka/orkaapi/api/v1"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	OrkaJobTypeLabel             = "orka.macstadium.com/job.type"
	OrkaJobTypeRegistryPushValue = "registry-push"
	OCIImageNameAnnotationKey    = "orka.macstadium.com/oci-image"
)

type OrkaClient interface {
	Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error
	Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error
	Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error
	WaitForVm(ctx context.Context, namespace, name string, timeout int) (string, int, error)
	WaitForImage(ctx context.Context, name string) error
	WaitForPush(ctx context.Context, namespace, name string) error
}

type RealOrkaClient struct {
	client.WithWatch
}

// GetOrkaClient returns a runtime client with the on-disk discovery cache enabled
func GetOrkaClient(orkaEndpoint, authToken string) (*RealOrkaClient, error) {
	sch := runtime.NewScheme()
	if err := orkav1.AddToScheme(sch); err != nil {
		log.Fatal("failed to add orkav1 to scheme")
	}
	if err := corev1.AddToScheme(sch); err != nil {
		log.Fatal("failed to add corev1 to scheme")
	}

	endpoint, err := url.JoinPath(orkaEndpoint, "api", "v1", "cluster-info")
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	clusterInfo := struct {
		APIEndpoint string `json:"apiEndpoint"`
		APIDomain   string `json:"apiDomain"`
		CertData    string `json:"certData"`
	}{}
	if err := json.Unmarshal(bodyBytes, &clusterInfo); err != nil {
		return nil, err
	}

	restConfig := &rest.Config{
		Host:        clusterInfo.APIEndpoint,
		BearerToken: authToken,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: []byte(clusterInfo.CertData),
		},
	}

	// Determine if using a public IP address and update config with k8s apiserver name
	if clusterInfo.APIDomain != "" {
		ip := lookupIP(orkaEndpoint)
		if ip != nil && !ip.IsPrivate() {
			restConfig.Host = fmt.Sprintf("https://%s", ip)
			restConfig.TLSClientConfig.ServerName = clusterInfo.APIDomain
		}
	}

	c, err := client.NewWithWatch(restConfig, client.Options{Scheme: sch})
	if err != nil {
		return nil, err
	}

	return &RealOrkaClient{c}, nil
}

func lookupIP(orkaEndpoint string) net.IP {
	u, err := url.Parse(orkaEndpoint)
	if err != nil {
		return nil
	}
	ips, err := net.LookupIP(u.Hostname())
	if err != nil {
		return nil
	}
	return ips[0]
}

func (c *RealOrkaClient) WaitForVm(ctx context.Context, namespace, name string, timeout int) (string, int, error) {
	var host string
	var port int
	err := RetryOnWatcherErrorWithTimeout(ctx, time.Duration(timeout)*time.Minute, func(contextWithTimeout context.Context) error {
		var err error
		host, port, err = c.waitForVm(contextWithTimeout, namespace, name, timeout)
		return err
	}, 1*time.Second)
	return host, port, err
}

func (c *RealOrkaClient) waitForVm(ctx context.Context, namespace, name string, timeout int) (string, int, error) {
	vmiList := &orkav1.VirtualMachineInstanceList{}
	watcher, err := c.Watch(ctx, vmiList, client.InNamespace(namespace), client.MatchingFields{"metadata.name": name})
	if err != nil {
		return "", 0, err
	}

	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", 0, ctx.Err()

		case event, ok := <-watcher.ResultChan():
			if !ok {
				return "", 0, WatcherError{Err: errors.New("watcher closed unexpectedly")}
			}
			vmi := event.Object.(*orkav1.VirtualMachineInstance)

			if vmi.Status.Phase == orkav1.VMRunning {
				return vmi.Status.HostIP, *vmi.Status.SSHPort, nil
			}

			if vmi.Status.Phase == orkav1.VMFailed {
				err := c.Delete(ctx, vmi)
				return "", 0, errors.Join(fmt.Errorf("%s", vmi.Status.ErrorMessage), err)
			}
		}
	}
}

func (c *RealOrkaClient) WaitForImage(ctx context.Context, name string) error {
	return RetryOnWatcherErrorWithTimeout(ctx, 1*time.Hour, func(contextWithTimeout context.Context) error {
		return c.waitForImage(contextWithTimeout, name)
	}, 1*time.Second)
}

func (c *RealOrkaClient) waitForImage(ctx context.Context, name string) error {
	imageList := &orkav1.ImageList{}
	watcher, err := c.Watch(ctx, imageList, client.InNamespace(DefaultOrkaNamespace), client.MatchingFields{"metadata.name": name})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event, ok := <-watcher.ResultChan():
			if !ok {
				return WatcherError{Err: errors.New("watcher closed unexpectedly")}
			}
			image := event.Object.(*orkav1.Image)

			switch image.Status.State {
			case orkav1.Ready:
				return nil
			case orkav1.Failed:
				return errors.New(image.Status.ErrorMessage)
			}
		}
	}
}

// TODO: Add configurable image push timeout
func (c *RealOrkaClient) WaitForPush(ctx context.Context, namespace, name string) error {
	return RetryOnWatcherErrorWithTimeout(ctx, 1*time.Hour, func(contextWithTimeout context.Context) error {
		return c.waitForPush(contextWithTimeout, namespace, name)
	}, 1*time.Second)
}

func (c *RealOrkaClient) waitForPush(ctx context.Context, namespace, name string) error {
	matchLabels := client.MatchingLabels{OrkaJobTypeLabel: OrkaJobTypeRegistryPushValue}
	if len(name) > 0 {
		matchLabels[batchv1.JobNameLabel] = name
	}

	pods := &corev1.PodList{}
	watcher, err := c.Watch(ctx, pods, client.InNamespace(namespace), matchLabels)
	if err != nil {
		return fmt.Errorf("watcher failed to initilize: %s", err)
	}

	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event, ok := <-watcher.ResultChan():
			if !ok {
				return WatcherError{Err: errors.New("watcher closed unexpectedly")}
			}

			if event.Type == watch.Deleted {
				return errors.New("vm push pod has been deleted")
			}

			p := event.Object.(*corev1.Pod)

			switch p.Status.Phase {
			case corev1.PodSucceeded:
				return nil
			case corev1.PodFailed:
				return fmt.Errorf("failed to save image: %s", p.Status.Message)
			}
		}
	}
}
