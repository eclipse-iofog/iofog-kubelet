/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package vkubelet

import (
	"context"
	"strings"

	"github.com/cpuguy83/strongerrors/status/ocstatus"
	"github.com/eclipse-iofog/iofog-kubelet/log"
	"github.com/eclipse-iofog/iofog-kubelet/trace"
	"github.com/eclipse-iofog/iofog-kubelet/versions"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// vkVersion is a concatenation of the Kubernetes version the VK is built against, the string "vk" and the VK release version.
	// TODO @pires revisit after VK 1.0 is released as agreed in https://github.com/eclipse-iofog/iofog-kubelet/pull/446#issuecomment-448423176.
	vkVersion = strings.Join([]string{"v1.13.1", "vk", version.Version}, "-")
)

// registerNode registers the virtual node with the Kubernetes API.
func (s *Server) registerNode(ctx context.Context) error {
	ctx, span := trace.StartSpan(ctx, "registerNode")
	defer span.End()

	taints := make([]corev1.Taint, 0)

	if s.taint != nil {
		taints = append(taints, *s.taint)
	}

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.nodeName,
			Labels: map[string]string{
				"type":                   "iofog-kubelet",
				"kubernetes.io/role":     "agent",
				"beta.kubernetes.io/os":  strings.ToLower(s.provider.OperatingSystem()),
				"kubernetes.io/hostname": s.nodeName,
				"alpha.service-controller.kubernetes.io/exclude-balancer": "true",
			},
		},
		Spec: corev1.NodeSpec{
			Taints: taints,
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				OperatingSystem: s.provider.OperatingSystem(),
				Architecture:    "amd64",
				KubeletVersion:  vkVersion,
			},
			Capacity:        s.provider.Capacity(ctx),
			Allocatable:     s.provider.Capacity(ctx),
			Conditions:      s.provider.NodeConditions(ctx),
			Addresses:       s.provider.NodeAddresses(ctx),
			DaemonEndpoints: *s.provider.NodeDaemonEndpoints(ctx),
		},
	}
	ctx = addNodeAttributes(ctx, span, node)
	if _, err := s.Client.CoreV1().Nodes().Create(node); err != nil && !errors.IsAlreadyExists(err) {
		span.SetStatus(ocstatus.FromError(err))
		return err
	}

	log.G(ctx).Info("Registered node")

	return nil
}

// updateNode updates the node status within Kubernetes with updated NodeConditions.
func (s *Server) updateNode(ctx context.Context) {
	ctx, span := trace.StartSpan(ctx, "updateNode")
	defer span.End()

	opts := metav1.GetOptions{}
	n, err := s.Client.CoreV1().Nodes().Get(s.nodeName, opts)
	if err != nil && !errors.IsNotFound(err) {
		log.G(ctx).WithError(err).Error("Failed to retrive node")
		span.SetStatus(ocstatus.FromError(err))
		return
	}

	ctx = addNodeAttributes(ctx, span, n)
	log.G(ctx).Debug("Fetched node details from k8s")

	if errors.IsNotFound(err) {
		if err = s.registerNode(ctx); err != nil {
			log.G(ctx).WithError(err).Error("Failed to register node")
			span.SetStatus(ocstatus.FromError(err))
		} else {
			log.G(ctx).Debug("Registered node in k8s")
		}
		return
	}

	n.ResourceVersion = "" // Blank out resource version to prevent object has been modified error
	n.Status.Conditions = s.provider.NodeConditions(ctx)

	capacity := s.provider.Capacity(ctx)
	allocatable := s.provider.Allocatable(ctx)
	n.Status.Capacity = capacity
	n.Status.Allocatable = allocatable

	n.Status.Addresses = s.provider.NodeAddresses(ctx)

	n, err = s.Client.CoreV1().Nodes().UpdateStatus(n)
	if err != nil {
		log.G(ctx).WithError(err).Error("Failed to update node")
		span.SetStatus(ocstatus.FromError(err))
		return
	}
}

// deleteNode deletes the virtual node with the Kubernetes API.
func (s *Server) DeleteNode(ctx context.Context) error {
	ctx, span := trace.StartSpan(ctx, "deleteNode")
	defer span.End()

	deleteOptions := metav1.DeleteOptions{}
	if err := s.Client.CoreV1().Nodes().Delete(s.nodeName, &deleteOptions); err != nil && !errors.IsNotFound(err) {
		span.SetStatus(ocstatus.FromError(err))
		return err
	}

	log.G(ctx).Info("Deleted node")

	return nil
}

type taintsStringer []corev1.Taint

func (t taintsStringer) String() string {
	var s string
	for _, taint := range t {
		if s == "" {
			s = taint.Key + "=" + taint.Value + ":" + string(taint.Effect)
		} else {
			s += ", " + taint.Key + "=" + taint.Value + ":" + string(taint.Effect)
		}
	}
	return s
}

func addNodeAttributes(ctx context.Context, span trace.Span, n *corev1.Node) context.Context {
	return span.WithFields(ctx, log.Fields{
		"UID":     string(n.UID),
		"name":    n.Name,
		"cluster": n.ClusterName,
		"tains":   taintsStringer(n.Spec.Taints),
	})
}
