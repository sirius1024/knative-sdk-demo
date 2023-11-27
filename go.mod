module github.com/sirius1024/knative-sdk-demo

go 1.15

require (
	k8s.io/api v0.27.6
	k8s.io/apimachinery v0.27.6
	knative.dev/client v0.17.0
	knative.dev/serving v0.39.0
)

replace (
	k8s.io/api => k8s.io/api v0.17.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
	k8s.io/code-generator => k8s.io/code-generator v0.17.6
)
