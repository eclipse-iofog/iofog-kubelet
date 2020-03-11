package iofog

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofog-kubelet/vkubelet/api"
	"github.com/eclipse-iofog/iofog-kubelet/log"
	"io"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/remotecommand"
)

// BrokerProvider implements the iofog-kubelet provider interface by forwarding kubelet calls to a iofog endpoint.
type BrokerProvider struct {
	operatingSystem    string
	controller         apps.IofogController
	client             *client.Client
	nodeId             string
	nodeName           string
	daemonEndpointPort int32
	store              *api.KeyValueStore
}

type FlowPod struct {
	FlowInfo *client.FlowInfo
	Pod      *v1.Pod
}

// NewBrokerProvider creates a new BrokerProvider
func NewBrokerProvider(daemonEndpointPort int32, nodeName, operatingSystem string, controller apps.IofogController, controllerClient *client.Client, nodeId string, store *api.KeyValueStore) (*BrokerProvider, error) {
	provider := BrokerProvider{
		nodeName:           nodeName,
		nodeId:             nodeId,
		operatingSystem:    operatingSystem,
		daemonEndpointPort: daemonEndpointPort,
		controller:         controller,
		client:             controllerClient,
		store:              store,
	}

	return &provider, nil
}

// CreatePod accepts a Pod definition and forwards the call to the iofog endpoint
func (p *BrokerProvider) CreatePod(ctx context.Context, pod *v1.Pod) error {
	return p.createUpdatePod(pod)
}

// UpdatePod accepts a Pod definition and forwards the call to the iofog endpoint
func (p *BrokerProvider) UpdatePod(ctx context.Context, pod *v1.Pod) error {
	return p.createUpdatePod(pod)
}

// DeletePod accepts a Pod definition and forwards the call to the iofog endpoint
func (p *BrokerProvider) DeletePod(ctx context.Context, pod *v1.Pod) error {
	if flowPod, err := p.getFlowPod(pod.Name); err != nil {
		return err
	} else {
		if err = p.client.DeleteFlow(flowPod.FlowInfo.ID); err != nil {
			return err
		}
		return p.store.Remove(pod.Name)
	}
}

// GetPod returns a pod by name that is being managed by the iofog server
func (p *BrokerProvider) GetPod(ctx context.Context, namespace, name string) (*v1.Pod, error) {
	if flowPod, err := p.getFlowPod(name); err != nil {
		return nil, err
	} else {
		return flowPod.Pod, nil
	}
}

// GetContainerLogs returns the logs of a container running in a pod by name.
func (p *BrokerProvider) GetContainerLogs(ctx context.Context, namespace, podName, containerName string, tail int) (string, error) {
	return "", nil
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
	return nil
}

// GetPodStatus retrieves the status of a given pod by name.
func (p *BrokerProvider) GetPodStatus(ctx context.Context, namespace, name string) (*v1.PodStatus, error) {
	flowPod, err := p.getFlowPod(name)
	if err != nil {
		return nil, err
	}

	microservices, err := p.client.GetMicroservicesPerFlow(flowPod.FlowInfo.ID)
	if err != nil {
		return nil, err
	}

	containersStatus := []v1.ContainerStatus{}
	podPhase := v1.PodRunning
	podReady := v1.ConditionStatus("True")
	var podStartTime metav1.Time
	for _, microservice := range microservices.Microservices {
		podStartTime = metav1.Time{Time: time.Unix(microservice.Status.StartTimne/1000, -1)}
		var containerState v1.ContainerState
		microserviceReady := true
		if microservice.Status.Status != "RUNNING" {
			podPhase = v1.PodPending
			podReady = v1.ConditionStatus("False")
			containerState = v1.ContainerState{
				Waiting: &v1.ContainerStateWaiting{
					Reason: microservice.Status.Status,
				},
			}
			microserviceReady = false
		} else {
			containerState = v1.ContainerState{
				Running: &v1.ContainerStateRunning{
					StartedAt: podStartTime,
				},
			}
		}
		containersStatus = append(containersStatus, v1.ContainerStatus{
			Name:         microservice.Name,
			ImageID:      microservice.UUID,
			Ready:        microserviceReady,
			RestartCount: 0,
			State:        containerState,
		})
	}

	podStatus := v1.PodStatus{
		Phase:     podPhase,
		StartTime: &podStartTime,
		Conditions: []v1.PodCondition{
			{
				Type:   "PodInitialized",
				Status: "True",
			},
			{
				Type:   "PodReady",
				Status: podReady,
			},
			{
				Type:   "PodScheduled",
				Status: "True",
			},
		},
		ContainerStatuses: containersStatus,
	}
	return &podStatus, nil
}

// GetPods retrieves a list of all pods scheduled to run.
func (p *BrokerProvider) GetPods(ctx context.Context) ([]*v1.Pod, error) {
	podNames := p.store.Keys()
	pods := make([]*v1.Pod, 0, p.store.Size())
	if len(podNames) == 0 {
		return pods, nil
	}

	for _, podName := range podNames {
		flowPod := &FlowPod{}
		if err := p.store.Get(podName, flowPod); err != nil {
			return nil, err
		}
		pods = append(pods, flowPod.Pod)
	}
	return pods, nil
}

// Capacity returns a resource list containing the capacity limits
func (p *BrokerProvider) Capacity(ctx context.Context) v1.ResourceList {
	node, err := p.client.GetAgentByID(p.nodeId)
	if err != nil {
		log.L.Error("Error getting node capacity: ", err)
		return nil
	}

	return v1.ResourceList{
		"cpu":    *resource.NewQuantity(node.CPULimit, resource.DecimalSI),
		"memory": *resource.NewQuantity(node.MemoryLimit, resource.DecimalSI),
		"pods":   *resource.NewQuantity(100, resource.DecimalSI),
	}
}

// Allocatable returns a resource list containing the allocatable limits
func (p *BrokerProvider) Allocatable(ctx context.Context) v1.ResourceList {
	node, err := p.client.GetAgentByID(p.nodeId)
	if err != nil {
		log.L.Error("Error getting node's allocatable resources: ", err)
		return nil
	}

	return v1.ResourceList{
		"cpu":    *resource.NewQuantity(node.CPULimit-int64(node.CPUUsage), resource.DecimalSI),
		"memory": *resource.NewQuantity(node.MemoryLimit-int64(node.MemoryUsage), resource.DecimalSI),
		"pods":   *resource.NewQuantity(100, resource.DecimalSI),
	}
}

// NodeConditions returns a list of conditions (Ready, OutOfDisk, etc), for updates to the node status
func (p *BrokerProvider) NodeConditions(ctx context.Context) []v1.NodeCondition {
	node, err := p.client.GetAgentByID(p.nodeId)

	var nodeRunning v1.ConditionStatus = "False"
	daemonStatus := "UNKNOWN"

	var outOfDisk v1.ConditionStatus = "False"
	var diskPressure v1.ConditionStatus = "False"
	diskUsage := ""

	var memoryPressure v1.ConditionStatus = "False"
	memoryUsage := ""

	var now = metav1.Time{Time: time.Now()}

	if err == nil && node != nil {
		now = metav1.Time{Time: time.Unix(node.LastStatusTimeMsUTC/1000, -1)}

		daemonStatus = node.DaemonStatus
		if node.DaemonStatus != "RUNNING" {
			nodeRunning = "False"
		}

		diskUsage = "Usage: " + fmt.Sprintf("%f.0", node.DiskUsage) + ", Limit: " + string(node.DiskLimit)
		if int64(node.DiskUsage) >= node.DiskLimit {
			outOfDisk = "True"
		}

		if (node.DiskUsage / float64(node.DiskLimit)) >= 0.9 {
			diskPressure = "True"
		}

		memoryUsage = "Usage: " + fmt.Sprintf("%f.0", node.MemoryUsage) + ", Limit: " + string(node.MemoryLimit)
		if (node.MemoryUsage / float64(node.MemoryLimit)) >= 0.9 {
			memoryPressure = "True"
		}
	}

	condition := []v1.NodeCondition{
		{
			Type:               "Ready",
			Status:             nodeRunning,
			LastHeartbeatTime:  now,
			LastTransitionTime: now,
			Reason:             "",
			Message:            daemonStatus,
		},
		{
			Type:               "OutOfDisk",
			Status:             outOfDisk,
			LastHeartbeatTime:  now,
			LastTransitionTime: now,
			Reason:             "",
			Message:            diskUsage,
		},
		{
			Type:               "MemoryPressure",
			Status:             memoryPressure,
			LastHeartbeatTime:  now,
			LastTransitionTime: now,
			Reason:             "",
			Message:            memoryUsage,
		},
		{
			Type:               "DiskPressure",
			Status:             diskPressure,
			LastHeartbeatTime:  now,
			LastTransitionTime: now,
			Reason:             "",
			Message:            diskUsage,
		},
		{
			Type:               "NetworkUnavailable",
			Status:             "False",
			LastHeartbeatTime:  now,
			LastTransitionTime: now,
			Reason:             "",
			Message:            "",
		},
	}

	return condition
}

// NodeAddresses returns a list of addresses for the node status
// within Kubernetes.
func (p *BrokerProvider) NodeAddresses(ctx context.Context) []v1.NodeAddress {
	node, err := p.client.GetAgentByID(p.nodeId)
	if err != nil {
		log.L.Error("Error getting node's IP': ", err)
		return nil
	}

	nodeAddresses := []v1.NodeAddress{
		{
			Type:    "InternalIP",
			Address: node.IPAddress,
		},
		{
			Type:    "ExternalIP",
			Address: node.IPAddressExternal,
		},
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
	return "linux"
}

func (p *BrokerProvider) createUpdatePod(pod *v1.Pod) error {
	var application *apps.Application
	application, err := p.convertAnnotationToApplication(pod)
	if err != nil {
		return err
	}

	if err := apps.DeployApplication(p.controller, *application); err != nil {
		return err
	}

	flow, err := p.client.GetFlowByName(pod.Name)
	if err != nil {
		return err
	}

	return p.storeFlowInfo(flow, pod)
}

func (p *BrokerProvider) convertAnnotationToApplication(pod *v1.Pod) (*apps.Application, error) {
	microservicesString := pod.Annotations["microservices"]
	routesString := pod.Annotations["routes"]

	microservices := []apps.Microservice{}
	routes := []apps.Route{}
	if err := json.Unmarshal([]byte(microservicesString), &microservices); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(routesString), &routes); err != nil {
		return nil, err
	}

	node, _ := p.client.GetAgentByID(p.nodeId)
	i := 0
	s := len(microservices)
	for i < s {
		microservices[i].Agent = apps.MicroserviceAgent{
			Name: node.Name,
		}
		i++
	}

	application := &apps.Application{
		Name:          pod.Name,
		Microservices: microservices,
		Routes:        routes,
	}

	return application, nil
}

func (p *BrokerProvider) getFlowPod(podName string) (*FlowPod, error) {
	flowPod := &FlowPod{}
	if err := p.store.Get(podName, flowPod); err != nil {
		return nil, err
	}

	return flowPod, nil
}

func (p *BrokerProvider) storeFlowInfo(flowInfo *client.FlowInfo, pod *v1.Pod) error {
	flowPod := FlowPod{
		FlowInfo: flowInfo,
		Pod:      pod,
	}
	return p.store.Put(pod.Name, flowPod)
}
