// Copyright © 2017 The virtual-kubelet authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vkubelet

import (
	"context"
	"time"

	"go.opencensus.io/trace"
	corev1 "k8s.io/api/core/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/eclipse-iofog/iofog-kubelet/manager"
	"github.com/eclipse-iofog/iofog-kubelet/providers"
)

const (
	podStatusReasonProviderFailed = "ProviderFailed"
)

// Server masquarades itself as a kubelet and allows for the virtual node to be backed by non-vm/node providers.
type Server struct {
	nodeName        string
	namespace       string
	Client          *kubernetes.Clientset
	taint           *corev1.Taint
	provider        providers.Provider
	resourceManager *manager.ResourceManager
	podSyncWorkers  int
	podInformer     corev1informers.PodInformer
}

// Config is used to configure a new server.
type Config struct {
	Client          *kubernetes.Clientset
	Namespace       string
	NodeName        string
	Provider        providers.Provider
	ResourceManager *manager.ResourceManager
	Taint           *corev1.Taint
	PodSyncWorkers  int
	PodInformer     corev1informers.PodInformer
}

// New creates a new iofog-kubelet server.
// This is the entrypoint to this package.
//
// This creates but does not start the server.
// You must call `Run` on the returned object to start the server.
func New(cfg Config) *Server {
	return &Server{
		namespace:       cfg.Namespace,
		nodeName:        cfg.NodeName,
		taint:           cfg.Taint,
		Client:          cfg.Client,
		resourceManager: cfg.ResourceManager,
		provider:        cfg.Provider,
		podSyncWorkers:  cfg.PodSyncWorkers,
		podInformer:     cfg.PodInformer,
	}
}

// Run creates and starts an instance of the pod controller, blocking until it stops.
//
// Note that this does not setup the HTTP routes that are used to expose pod
// info to the Kubernetes API Server, such as logs, metrics, exec, etc.
// See `AttachPodRoutes` and `AttachMetricsRoutes` to set these up.
func (s *Server) Run(ctx context.Context) error {
	if err := s.registerNode(ctx); err != nil {
		return err
	}

	go s.providerSyncLoop(ctx)

	return NewPodController(s).Run(ctx, s.podSyncWorkers)
}

// providerSyncLoop syncronizes pod states from the provider back to kubernetes
func (s *Server) providerSyncLoop(ctx context.Context) {
	const sleepTime = 1 * time.Second

	t := time.NewTimer(sleepTime)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			t.Stop()

			ctx, span := trace.StartSpan(ctx, "syncActualState")
			s.updateNode(ctx)
			s.updatePodStatuses(ctx)
			span.End()

			// restart the timer
			t.Reset(sleepTime)
		}
	}
}
