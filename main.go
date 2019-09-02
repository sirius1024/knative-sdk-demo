/*
Copyright 2019 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func main() {
	// consts
	const apiGroup string = "serving.knative.dev"
	const apiVersion string = "v1alpha1"
	const apiResource string = "services"
	const kindService string = "Service"

	// variables
	namespace := "sdk-test"
	serviceName := "knative-web-demo-gt-sdk-test"
	dockerImage := "guotuo1024/knative-web-demo:version-1.0.0"
	dockerImageUpgrade := "guotuo1024/knative-web-demo:version-2.0.0"

	//kubenetes configuration
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// knative serving
	knativeResource := schema.GroupVersionResource{Group: apiGroup, Version: apiVersion, Resource: apiResource}
	knativeService := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiGroup + "/" + apiVersion,
			"kind":       kindService,
			"metadata": map[string]interface{}{
				"name":      serviceName,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"autoscaling.knative.dev/target": "10",
						},
					},
					"spec": map[string]interface{}{
						"containers": []map[string]interface{}{
							{
								"image": dockerImage,
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating knative serving service...")
	result, err := client.Resource(knativeResource).Namespace(namespace).Create(knativeService, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetName())

	// Update Deployment
	prompt()
	fmt.Println("Updating deployment...")
	//    You have two options to Update() this Deployment:
	//
	//    1. Modify the "deployment" variable and call: Update(deployment).
	//       This works like the "kubectl replace" command and it overwrites/loses changes
	//       made by other clients between you Create() and Update() the object.
	//    2. Modify the "result" returned by Get() and retry Update(result) until
	//       you no longer get a conflict error. This way, you can preserve changes made
	//       by other clients between Create() and Update(). This is implemented below
	//			 using the retry utility package included with client-go. (RECOMMENDED)
	//
	// More Info:
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := client.Resource(knativeResource).Namespace(namespace).Get(serviceName, metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("failed to get latest version of Deployment: %v", getErr))
		}

		// // update replicas to 1
		// if err := unstructured.SetNestedField(result.Object, int64(1), "spec", "replicas"); err != nil {
		// 	panic(fmt.Errorf("failed to set replica value: %v", err))
		// }

		// extract spec containers
		containers, found, err := unstructured.NestedSlice(result.Object, "spec", "template", "spec", "containers")
		if err != nil || !found || containers == nil {
			panic(fmt.Errorf("deployment containers not found or error in spec: %v", err))
		}

		// update container[0] image
		if err := unstructured.SetNestedField(containers[0].(map[string]interface{}), dockerImageUpgrade, "image"); err != nil {
			panic(err)
		}
		if err := unstructured.SetNestedField(result.Object, containers, "spec", "template", "spec", "containers"); err != nil {
			panic(err)
		}

		_, updateErr := client.Resource(knativeResource).Namespace(namespace).Update(result, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("update failed: %v", retryErr))
	}
	fmt.Println("Updated knative CRD...")

	// List Deployments
	prompt()
	// fmt.Printf("Listing knative service in namespace %q:\n", apiv1.NamespaceDefault)
	fmt.Printf("Listing knative service in namespace %q:\n", namespace)

	list, err := client.Resource(knativeResource).Namespace(namespace).List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {

		ns, found, err := unstructured.NestedString(d.Object, "metadata", "namespace")
		if err != nil || !found {
			fmt.Printf("Replicas not found for deployment %s: error=%s", d.GetName(), err)
			continue
		}
		fmt.Printf(" * %s (%s)\n", d.GetName(), ns)
	}

	// Delete Deployment
	prompt()
	fmt.Println("Deleting knative service...")
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	if err := client.Resource(knativeResource).Namespace(namespace).Delete(serviceName, deleteOptions); err != nil {
		panic(err)
	}

	fmt.Println("Deleted deployment.")
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
