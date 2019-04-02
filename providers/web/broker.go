// Package web provides an implementation of the virtual kubelet provider interface
// by forwarding all calls to a web endpoint. The web endpoint to which requests
// must be forwarded must be specified through an environment variable called
// `WEB_ENDPOINT_URL`. This endpoint must implement the following HTTP APIs:
//  - POST /createPod
//  - PUT /updatePod
//  - DELETE /deletePod
//  - GET /getPod?namespace=[namespace]&name=[pod name]
//  - GE /getContainerLogs?namespace=[namespace]&podName=[pod name]&containerName=[container name]&tail=[tail value]
//  - GET /getPodStatus?namespace=[namespace]&name=[pod name]
//  - GET /getPods
//  - GET /capacity
//  - GET /nodeConditions
//  - GET /nodeAddresses
package web

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eclipse-iofog/iofog-kubelet/controller"
	"io"
	"log"
	"net/url"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/remotecommand"
)

// BrokerProvider implements the iofog-kubelet provider interface by forwarding kubelet calls to a web endpoint.
type BrokerProvider struct {
	operatingSystem    string
	client             *controller.HttpClient
	nodeId             string
	nodeName           string
	daemonEndpointPort int32
}

// NewBrokerProvider creates a new BrokerProvider
func NewBrokerProvider(daemonEndpointPort int32, nodeName, operatingSystem, controllerToken, controllerUrl, nodeId string) (*BrokerProvider, error) {
	var provider BrokerProvider

	provider.nodeName = nodeName
	provider.nodeId = nodeId
	provider.operatingSystem = operatingSystem
	provider.daemonEndpointPort = daemonEndpointPort

	var err error
	provider.client, err = controller.NewHttpClient(controllerToken, controllerUrl)
	if err != nil {
		return nil, err
	}

	return &provider, nil
}

// CreatePod accepts a Pod definition and forwards the call to the web endpoint
func (p *BrokerProvider) CreatePod(ctx context.Context, pod *v1.Pod) error {
	urlPathStr := fmt.Sprintf(
		"/api/v3/k8s/createPod?nodeName=%s",
		url.QueryEscape(p.nodeId))
	return p.createUpdatePod(pod, "POST", urlPathStr)
}

// UpdatePod accepts a Pod definition and forwards the call to the web endpoint
func (p *BrokerProvider) UpdatePod(ctx context.Context, pod *v1.Pod) error {
	urlPathStr := fmt.Sprintf(
		"/api/v3/k8s/updatePod?nodeName=%s",
		url.QueryEscape(p.nodeId))
	return p.createUpdatePod(pod, "PUT", urlPathStr)
}

// DeletePod accepts a Pod definition and forwards the call to the web endpoint
func (p *BrokerProvider) DeletePod(ctx context.Context, pod *v1.Pod) error {
	urlPathStr := fmt.Sprintf(
		"/api/v3/k8s/deletePod?nodeName=%s",
		url.QueryEscape(p.nodeId))
	urlPath, err := url.Parse(urlPathStr)
	if err != nil {
		return err
	}

	// encode pod definition as JSON and post request
	podJSON, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	_, err = p.client.DoRequest("DELETE", urlPath, podJSON, false)
	return err
}

// GetPod returns a pod by name that is being managed by the web server
func (p *BrokerProvider) GetPod(ctx context.Context, namespace, name string) (*v1.Pod, error) {
	urlPathStr := fmt.Sprintf(
		"/api/v3/k8s/getPod?namespace=%s&name=%s&nodeName=%s",
		url.QueryEscape(namespace),
		url.QueryEscape(name),
		url.QueryEscape(p.nodeId))

	var pod v1.Pod
	err := p.client.DoGetRequest(urlPathStr, &pod)

	// if we get a "404 Not Found" then we return nil to indicate that no pod
	// with this name was found
	if err != nil {
		if err.Error() == "404 Not Found" {
			return nil, nil
		}
		return nil, err
	}

	return &pod, err
}

// GetContainerLogs returns the logs of a container running in a pod by name.
func (p *BrokerProvider) GetContainerLogs(ctx context.Context, namespace, podName, containerName string, tail int) (string, error) {
	urlPathStr := fmt.Sprintf(
		"/api/v3/k8s/getContainerLogs?namespace=%s&podName=%s&containerName=%s&tail=%d&nodeName=%s",
		url.QueryEscape(namespace),
		url.QueryEscape(podName),
		url.QueryEscape(containerName),
		tail,
		url.QueryEscape(p.nodeId))

	response, err := p.client.DoGetRequestBytes(urlPathStr)
	if err != nil {
		return "", err
	}

	return string(response), nil
}

// Get full pod name as defined in the provider context
// TODO: Implementation
func (p *BrokerProvider) GetPodFullName(namespace string, pod string) string {
	return pod
}

// ExecInContainer executes a command in a container in the pod, copying data
// between in/out/err and the container's stdin/stdout/stderr.
// TODO: Implementation
func (p *BrokerProvider) ExecInContainer(name string, uid types.UID, container string, cmd []string, in io.Reader, out, err io.WriteCloser, tty bool, resize <-chan remotecommand.TerminalSize, timeout time.Duration) error {
	log.Printf("receive ExecInContainer %q\n", container)
	return nil
}

// GetPodStatus retrieves the status of a given pod by name.
func (p *BrokerProvider) GetPodStatus(ctx context.Context, namespace, name string) (*v1.PodStatus, error) {
	urlPathStr := fmt.Sprintf(
		"/api/v3/k8s/getPodStatus?namespace=%s&name=%s&nodeName=%s",
		url.QueryEscape(namespace),
		url.QueryEscape(name),
		url.QueryEscape(p.nodeId))

	var podStatus v1.PodStatus
	err := p.client.DoGetRequest(urlPathStr, &podStatus)

	// if we get a "404 Not Found" then we return nil to indicate that no pod
	// with this name was found
	if err != nil && err.Error() == "404 Not Found" {
		return nil, nil
	}

	return &podStatus, err
}

// GetPods retrieves a list of all pods scheduled to run.
func (p *BrokerProvider) GetPods(ctx context.Context) ([]*v1.Pod, error) {
	urlPathStr := fmt.Sprintf(
		"/api/v3/k8s/getPods?nodeName=%s",
		url.QueryEscape(p.nodeId))

	var pods []*v1.Pod
	err := p.client.DoGetRequest(urlPathStr, &pods)

	return pods, err
}

// Capacity returns a resource list containing the capacity limits
func (p *BrokerProvider) Capacity(ctx context.Context) v1.ResourceList {
	urlPathStr := fmt.Sprintf(
		"/api/v3/k8s/capacity?nodeName=%s",
		url.QueryEscape(p.nodeId))

	resourceList := v1.ResourceList{}
	err := p.client.DoGetRequest(urlPathStr, &resourceList)

	if err != nil {
		log.Println("error ", err)
		return nil
	}

	return resourceList
}

// Allocatable returns a resource list containing the allocatable limits
func (p *BrokerProvider) Allocatable(ctx context.Context) v1.ResourceList {
	urlPathStr := fmt.Sprintf(
		"/api/v3/k8s/allocatable?nodeName=%s",
		url.QueryEscape(p.nodeId))

	resourceList := v1.ResourceList{}
	err := p.client.DoGetRequest(urlPathStr, &resourceList)

	if err != nil {
		log.Println("error ", err)
		return nil
	}

	return resourceList
}

// NodeConditions returns a list of conditions (Ready, OutOfDisk, etc), for updates to the node status
func (p *BrokerProvider) NodeConditions(ctx context.Context) []v1.NodeCondition {
	urlPathStr := fmt.Sprintf(
		"/api/v3/k8s/nodeConditions?nodeName=%s",
		url.QueryEscape(p.nodeId))

	var nodeConditions []v1.NodeCondition
	err := p.client.DoGetRequest(urlPathStr, &nodeConditions)

	if err != nil {
		log.Println("error ", err)
		return nil
	}

	return nodeConditions
}

// NodeAddresses returns a list of addresses for the node status
// within Kubernetes.
func (p *BrokerProvider) NodeAddresses(ctx context.Context) []v1.NodeAddress {
	urlPathStr := fmt.Sprintf(
		"/api/v3/k8s/nodeAddresses?nodeName=%s",
		url.QueryEscape(p.nodeId))

	var nodeAddresses []v1.NodeAddress
	err := p.client.DoGetRequest(urlPathStr, &nodeAddresses)

	if err != nil {
		log.Println("error ", err)
		return nil
	}

	return nodeAddresses
}

// NodeDaemonEndpoints returns NodeDaemonEndpoints for the node status
// within Kubernetes.
func (p *BrokerProvider) NodeDaemonEndpoints(ctx context.Context) *v1.NodeDaemonEndpoints {
	return &v1.NodeDaemonEndpoints{
		KubeletEndpoint: v1.DaemonEndpoint{
			Port: p.daemonEndpointPort,
		},
	}
}

// OperatingSystem returns the operating system for this provider.
func (p *BrokerProvider) OperatingSystem() string {
	return p.operatingSystem
}

func (p *BrokerProvider) createUpdatePod(pod *v1.Pod, method, postPath string) error {
	// build the post url
	postPathURL, err := url.Parse(postPath)
	if err != nil {
		return err
	}

	// encode pod definition as JSON and post request
	podJSON, err := json.Marshal(pod)
	if err != nil {
		return err
	}
	_, err = p.client.DoRequest(method, postPathURL, podJSON, false)
	return err
}
