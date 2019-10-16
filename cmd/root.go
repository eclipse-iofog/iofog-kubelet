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

package cmd

import (
	"context"
	"fmt"
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofog-kubelet/providers/register"
	"github.com/eclipse-iofog/iofog-kubelet/vkubelet"
	"github.com/eclipse-iofog/iofog-kubelet/vkubelet/api"
	"k8s.io/apimachinery/pkg/fields"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/eclipse-iofog/iofog-kubelet/log"
	logruslogger "github.com/eclipse-iofog/iofog-kubelet/log/logrus"
	"github.com/eclipse-iofog/iofog-kubelet/manager"
	"github.com/eclipse-iofog/iofog-kubelet/providers"
	"github.com/eclipse-iofog/iofog-kubelet/trace"
	"github.com/eclipse-iofog/iofog-kubelet/trace/opencensus"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	octrace "go.opencensus.io/trace"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
)

type IOFogKubelet struct {
	KubeletInstance   *vkubelet.Server
	NodeContextCancel context.CancelFunc
	NodeContext       context.Context
}

const (
	defaultDaemonPort = "10250"
	// kubeSharedInformerFactoryDefaultResync is the default resync period used by the shared informer factories for Kubernetes resources.
	// It is set to the same value used by the Kubelet, and can be overridden via the "--full-resync-period" flag.
	// https://github.com/kubernetes/kubernetes/blob/v1.12.2/pkg/kubelet/apis/config/v1beta1/defaults.go#L51
	kubeSharedInformerFactoryDefaultResync = 1 * time.Minute
)

var deleteNodeLock sync.Mutex
var controllerToken string
var controllerUrl string
var controller apps.IofogController
var controllerClient *client.Client
var kubeletConfig string
var kubeConfig string
var kubeNamespace string
var operatingSystem string
var provider string
var logLevel string
var taint *corev1.Taint
var kubeSharedInformerFactoryResync time.Duration
var podSyncWorkers int
var ioFogKubelets map[string]*IOFogKubelet
var configMapName string
var iofogNodes []client.AgentInfo

var userTraceExporters []string
var userTraceConfig = TracingExporterOptions{Tags: make(map[string]string)}
var traceSampler string

// Create a root context to be used by the pod controller and by the shared informer factories.
var rootContext, rootContextCancel = context.WithCancel(context.Background())

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "iofog-kubelet",
	Short: "iofog-kubelet provides a virtual kubelet interface for your kubernetes cluster.",
	Long: `iofog-kubelet implements the Kubelet interface with a pluggable
backend implementation allowing users to create kubernetes nodes without running the kubelet.
This allows users to schedule kubernetes workloads on nodes that aren't running Kubernetes.`,
	Run: func(cmd *cobra.Command, args []string) {
		defer rootContextCancel()

		controllerServer, err := setupControllerServer(rootContext, startKubelet, shutdownKubelet)
		if err != nil {
			log.L.WithError(err).Fatal("Error initializing controller server")
		}

		ioFogKubelets = make(map[string]*IOFogKubelet)

		controller = apps.IofogController{
			Token:    controllerToken,
			Endpoint: controllerUrl,
		}

		controllerClient, _ = client.NewWithToken(controllerUrl, controllerToken)

		iofogNodes = getIOFogNodes()
		for _, iofog := range iofogNodes {
			go startKubelet(iofog.UUID)
		}
		ioFogSyncLoop(rootContext)

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sig
			controllerServer.Close()
			shutdownAll()
			rootContextCancel()
		}()
		select {}
	},
}

func startKubelet(nodeId string) {
	_, ok := ioFogKubelets[nodeId]
	if ok {
		log.L.Warn("Node has already started ", nodeId)
		return
	}

	kubelet := new(IOFogKubelet)
	ioFogKubelets[nodeId] = kubelet

	nodeContext, nodeContextCancel := context.WithCancel(rootContext)
	kubelet.NodeContextCancel = nodeContextCancel
	kubelet.NodeContext = nodeContext

	nodeName := nodeName(nodeId)

	k8sClient, err := newClient(kubeConfig)
	if err != nil {
		log.L.WithError(err).Fatal("Error creating kubernetes client")
	}

	// Create a shared informer factory for Kubernetes pods in the current namespace (if specified) and scheduled to the current node.
	podInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(k8sClient, kubeSharedInformerFactoryResync, kubeinformers.WithNamespace(kubeNamespace), kubeinformers.WithTweakListOptions(func(options *metav1.ListOptions) {
		options.FieldSelector = fields.OneTermEqualSelector("spec.nodeName", nodeName).String()
	}))
	// Create a pod informer so we can pass its lister to the resource manager.
	podInformer := podInformerFactory.Core().V1().Pods()

	// Create another shared informer factory for Kubernetes secrets and configmaps (not subject to any selectors).
	scmInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(k8sClient, kubeSharedInformerFactoryResync)
	// Create a secret informer and a config map informer so we can pass their listers to the resource manager.
	secretInformer := scmInformerFactory.Core().V1().Secrets()
	configMapInformer := scmInformerFactory.Core().V1().ConfigMaps()

	// Create a new instance of the resource manager that uses the listers above for pods, secrets and config maps.
	rm, err := manager.NewResourceManager(podInformer.Lister(), secretInformer.Lister(), configMapInformer.Lister())
	if err != nil {
		log.L.WithError(err).Fatal("Error initializing resource manager")
	}

	// Start the shared informer factory for pods.
	go podInformerFactory.Start(nodeContext.Done())
	// Start the shared informer factory for secrets and configmaps.
	go scmInformerFactory.Start(nodeContext.Done())

	daemonPortEnv := getEnv("KUBELET_PORT", defaultDaemonPort)
	daemonPort, err := strconv.Atoi(daemonPortEnv)
	if err != nil {
		log.L.WithError(err).WithField("value", daemonPortEnv).Fatal("Invalid value from KUBELET_PORT in environment")
	}

	configMap := k8sClient.CoreV1().ConfigMaps(kubeNamespace)
	store, err := api.NewKeyValueStore(configMap, configMapName)

	initConfig := register.InitConfig{
		NodeName:         nodeName,
		OperatingSystem:  operatingSystem,
		ResourceManager:  rm,
		DaemonPort:       int32(daemonPort),
		InternalIP:       os.Getenv("VKUBELET_POD_IP"),
		Controller:       controller,
		ControllerClient: controllerClient,
		NodeId:           nodeId,
		Store:            store,
	}

	providerInstance, err := register.GetProvider(provider, initConfig)
	if err != nil {
		log.L.WithError(err).Fatal("Error initializing provider")
	}

	kubelet.KubeletInstance = vkubelet.New(vkubelet.Config{
		Client:          k8sClient,
		Namespace:       kubeNamespace,
		NodeName:        initConfig.NodeName,
		Taint:           taint,
		Provider:        providerInstance,
		ResourceManager: rm,
		PodSyncWorkers:  podSyncWorkers,
		PodInformer:     podInformer,
	})

	if err := kubelet.KubeletInstance.Run(nodeContext); err != nil && errors.Cause(err) != context.Canceled {
		log.G(nodeContext).Fatal(err)
	}
}

func shutdownKubelet(nodeId string, deleteNode bool) {
	kubelet, ok := ioFogKubelets[nodeId]
	if !ok {
		log.L.Warn("ioFog Kubelet is not running for node ", nodeId)
		return
	}

	deleteNodeLock.Lock()
	if kubelet != nil && kubelet.KubeletInstance != nil {
		if deleteNode {
			_ = kubelet.KubeletInstance.DeleteNode(kubelet.NodeContext)
		}
		kubelet.NodeContextCancel()
		delete(ioFogKubelets, nodeId)
	}
	deleteNodeLock.Unlock()
}

func shutdownAll() {
	for nodeId := range ioFogKubelets {
		shutdownKubelet(nodeId, false)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.GetLogger(context.TODO()).WithError(err).Fatal("Error executing root command")
	}
}

func nodeName(nodeId string) string {
	return "iofog-" + strings.ToLower(nodeId)
}

type mapVar map[string]string

func (mv mapVar) String() string {
	var s string
	for k, v := range mv {
		if s == "" {
			s = fmt.Sprintf("%s=%v", k, v)
		} else {
			s += fmt.Sprintf(", %s=%v", k, v)
		}
	}
	return s
}

func (mv mapVar) Set(s string) error {
	split := strings.SplitN(s, "=", 2)
	if len(split) != 2 {
		return errors.Errorf("invalid format, must be `key=value`: %s", s)
	}

	_, ok := mv[split[0]]
	if ok {
		return errors.Errorf("duplicate key: %s", split[0])
	}
	mv[split[0]] = split[1]
	return nil
}

func (mv mapVar) Type() string {
	return "map"
}

func init() {
	// make sure the default logger/tracer is initialized
	log.L = logruslogger.FromLogrus(logrus.NewEntry(logrus.StandardLogger()))
	trace.T = opencensus.Adapter{}

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	// RootCmd.PersistentFlags().StringVar(&kubeletConfig, "config", "", "config file (default is $HOME/.iofog-kubelet.yaml)")
	RootCmd.PersistentFlags().StringVar(&controllerToken, "iofog-token", "", "ioFog Controller token")
	RootCmd.PersistentFlags().StringVar(&controllerUrl, "iofog-url", "", "ioFog Controller URL")
	RootCmd.PersistentFlags().StringVar(&kubeConfig, "kubeconfig", "", "config file (default is $HOME/.kube/config)")
	RootCmd.PersistentFlags().StringVar(&kubeNamespace, "namespace", "", "kubernetes namespace (default is 'all')")
	RootCmd.PersistentFlags().StringVar(&configMapName, "config-map-name", "iofog-kubelet-store", "Config Map Name")
	RootCmd.PersistentFlags().StringVar(&operatingSystem, "os", "Linux", "Operating System (Linux/Windows)")
	provider = "iofog"

	RootCmd.PersistentFlags().MarkDeprecated("taint", "Taint key should now be configured using the VK_TAINT_KEY environment variable")
	RootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", `set the log level, e.g. "trace", debug", "info", "warn", "error"`)
	RootCmd.PersistentFlags().IntVar(&podSyncWorkers, "pod-sync-workers", 10, `set the number of pod synchronization workers`)

	RootCmd.PersistentFlags().StringSliceVar(&userTraceExporters, "trace-exporter", nil, fmt.Sprintf("sets the tracing exporter to use, available exporters: %s", AvailableTraceExporters()))
	RootCmd.PersistentFlags().StringVar(&userTraceConfig.ServiceName, "trace-service-name", "iofog-kubelet", "sets the name of the service used to register with the trace exporter")
	RootCmd.PersistentFlags().Var(mapVar(userTraceConfig.Tags), "trace-tag", "add tags to include with traces in key=value form")
	RootCmd.PersistentFlags().StringVar(&traceSampler, "trace-sample-rate", "", "set probability of tracing samples")

	RootCmd.PersistentFlags().DurationVar(&kubeSharedInformerFactoryResync, "full-resync-period", kubeSharedInformerFactoryDefaultResync, "how often to perform a full resync of pods between kubernetes and the provider")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if provider == "" {
		log.G(context.TODO()).Fatal("You must supply a cloud provider option: use --provider")
	}

	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		log.G(context.TODO()).WithError(err).Fatal("Error reading homedir")
	}

	if kubeletConfig != "" {
		// Use config file from the flag.
		viper.SetConfigFile(kubeletConfig)
	} else {
		// Search config in home directory with name ".iofog-kubelet" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".iofog-kubelet")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.G(context.TODO()).Debugf("Using config file %s", viper.ConfigFileUsed())
	}

	if kubeConfig == "" {
		kubeConfig = filepath.Join(home, ".kube", "config")

	}

	if kubeNamespace == "" {
		kubeNamespace = corev1.NamespaceAll
	}

	// Validate operating system.
	ok, _ := providers.ValidOperatingSystems[operatingSystem]
	if !ok {
		log.G(context.TODO()).WithField("OperatingSystem", operatingSystem).Fatalf("Operating system not supported. Valid options are: %s", strings.Join(providers.ValidOperatingSystems.Names(), " | "))
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		log.G(context.TODO()).WithField("logLevel", logLevel).Fatal("log level is not supported")
	}

	defaultNodeName := "iofog-kubelet"
	logrus.SetLevel(level)
	logger := logruslogger.FromLogrus(logrus.WithFields(logrus.Fields{
		"provider":        provider,
		"operatingSystem": operatingSystem,
		"node":            defaultNodeName,
		"namespace":       kubeNamespace,
	}))

	rootContext = log.WithLogger(rootContext, logger)

	log.L = logger

	taint, err = getTaint()
	if err != nil {
		logger.WithError(err).Fatal("Error setting up desired kubernetes node taint")
	}

	if podSyncWorkers <= 0 {
		logger.Fatal("The number of pod synchronization workers should not be negative")
	}

	for k := range userTraceConfig.Tags {
		if reservedTagNames[k] {
			logger.WithField("tag", k).Fatal("must not use a reserved tag key")
		}
	}
	userTraceConfig.Tags["operatingSystem"] = operatingSystem
	userTraceConfig.Tags["provider"] = provider
	userTraceConfig.Tags["nodeName"] = defaultNodeName
	for _, e := range userTraceExporters {
		if e == "zpages" {
			go setupZpages()
			continue
		}
		exporter, err := GetTracingExporter(e, userTraceConfig)
		if err != nil {
			log.L.WithError(err).WithField("exporter", e).Fatal("Cannot initialize exporter")
		}
		octrace.RegisterExporter(exporter)
	}
	if len(userTraceExporters) > 0 {
		var s octrace.Sampler
		switch strings.ToLower(traceSampler) {
		case "":
		case "always":
			s = octrace.AlwaysSample()
		case "never":
			s = octrace.NeverSample()
		default:
			rate, err := strconv.Atoi(traceSampler)
			if err != nil {
				logger.WithError(err).WithField("rate", traceSampler).Fatal("unsupported trace sample rate, supported values: always, never, or number 0-100")
			}
			if rate < 0 || rate > 100 {
				logger.WithField("rate", traceSampler).Fatal("trace sample rate must not be less than zero or greater than 100")
			}
			s = octrace.ProbabilitySampler(float64(rate) / 100)
		}

		if s != nil {
			octrace.ApplyConfig(
				octrace.Config{
					DefaultSampler: s,
				},
			)
		}
	}
}

func getIOFogNodes() []client.AgentInfo {
	if agents, err := controllerClient.ListAgents(); err != nil {
		log.G(rootContext).Fatal(err)
		return []client.AgentInfo{}
	} else {
		return agents.Agents
	}
}

func ioFogSyncLoop(ctx context.Context) {
	const sleepTime = 5 * time.Second

	t := time.NewTimer(sleepTime)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			t.Stop()

			nodes := getIOFogNodes()
			if nodes != nil {
				uuids := make(map[string]bool)
				for _, iofog := range nodes {
					uuids[iofog.UUID] = true
					_, ok := ioFogKubelets[iofog.UUID]
					if ok {
						continue
					}
					go startKubelet(iofog.UUID)
				}

				for uuid := range ioFogKubelets {
					_, ok := uuids[uuid]
					if !ok {
						shutdownKubelet(uuid, true)
					}
				}
			}

			// restart the timer
			t.Reset(sleepTime)
		}
	}
}
