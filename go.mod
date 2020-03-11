module github.com/eclipse-iofog/iofog-kubelet/v2

go 1.12

require (
	cloud.google.com/go v0.37.4 // indirect
	contrib.go.opencensus.io/exporter/ocagent v0.4.12
	git.apache.org/thrift.git v0.13.0 // indirect
	github.com/Azure/go-autorest v11.1.2+incompatible // indirect
	github.com/Sirupsen/logrus v1.0.6
	github.com/census-instrumentation/opencensus-proto v0.2.1 // indirect
	github.com/cpuguy83/strongerrors v0.2.1
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/eclipse-iofog/iofog-go-sdk/v2 v2.0.0-beta
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/gophercloud/gophercloud v0.0.0-20190126172459-c818fa66e4c8 // indirect
	github.com/gorilla/mux v1.7.2
	github.com/grpc-ecosystem/grpc-gateway v1.9.4 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.4
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0 // indirect
	go.opencensus.io v0.20.2
	golang.org/x/lint v0.0.0-20190409202823-959b441ac422 // indirect
	google.golang.org/api v0.3.2 // indirect
	google.golang.org/appengine v1.6.0 // indirect
	google.golang.org/genproto v0.0.0-20190716160619-c506a9f90610 // indirect
	google.golang.org/grpc v1.22.0 // indirect
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-00010101000000-000000000000 // indirect
	k8s.io/kubernetes v1.14.2
)

// Pinned to kubernetes-1.13.4
replace (
	github.com/openshift/api => github.com/openshift/api v0.0.0-20180801171038-322a19404e37
	k8s.io/api => k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190228180357-d002e88f6236
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190228174905-79427f02047f
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190228180923-a9e421a79326
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190228174230-b40b2a5939e4
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20181117043124-c2090bec4d9b
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190228175259-3e0149950b0e
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20180711000925-0cf8f7e6ed1d
	k8s.io/kubernetes => k8s.io/kubernetes v1.13.4
)
