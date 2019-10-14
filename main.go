package main

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/client/pkg/kn/commands"
	servinglib "knative.dev/client/pkg/serving"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
)

func main() {
	// get a service
	p := commands.KnParams{}
	p.Initialize()
	client, _ := p.NewClient("default")
	service, _ := client.GetService("helloworld-go")
	fmt.Println(service.GetName())

	// list services
	serviceList, _ := client.ListServices()
	for _, v := range serviceList.Items {
		fmt.Println(v.GetName())
	}

	// create a service
	var svcInstance = &v1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guotuo-sdk-test3",
			Namespace: "default",
		},
	}

	svcInstance.Spec.Template = &v1alpha1.RevisionTemplateSpec{
		Spec: v1alpha1.RevisionSpec{},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				servinglib.UserImageAnnotationKey: "",
			},
		},
	}

	// svcInstance.Spec.Template.Spec.PodSpec.Containers = []corev1.Container{{
	// 	Image: "docker.io/lijiawang/helloworld-go:v1",
	// 	Name:  "hwg",
	// }}

	svcInstance.Spec.Template.Spec.Containers = []corev1.Container{{Image: "guotuo1024/knative-web-demo:version-1.0.0"}}

	// servinglib.UpdateImage(svcInstance.Spec.Template, "docker.io/guotuo1024/knative-web-demo:v1")

	err := client.CreateService(svcInstance)
	if err != nil {
		fmt.Println(err)
	}
	// Update
	targetService, _ := client.GetService("guotuo-sdk-test3")
	fmt.Println("Will update service " + targetService.GetName())
	servinglib.UpdateImage(targetService.Spec.Template, "guotuo1024/knative-web-demo:version-2.0.0")
	client.UpdateService(targetService)
}
