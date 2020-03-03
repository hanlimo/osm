package main

import (
	"fmt"
	"os"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/deislabs/smc/demo/cmd/common"
)

func main() {
	acr := os.Getenv(common.ContainerRegistryEnvVar)
	containerRegistryCredsName := os.Getenv(common.ContainerRegistryCredsEnvVar)
	azureSubscription := os.Getenv(common.AzureSubscription)
	namespace := os.Getenv(common.KubeNamespaceEnvVar)

	labels := map[string]string{
		"app": common.AggregatedDiscoveryServiceName,
	}

	meta := metav1.ObjectMeta{
		Name:      common.AggregatedDiscoveryServiceName,
		Namespace: namespace,
		Labels:    labels,
	}

	if namespace == "" {
		fmt.Println("Empty namespace")
		os.Exit(1)
	}
	clientset := common.GetClient()

	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: meta,
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: fmt.Sprintf("%s-port", common.AggregatedDiscoveryServiceName),
					Port: common.AggregatedDiscoveryServicePort,
					TargetPort: intstr.IntOrString{
						IntVal: common.AggregatedDiscoveryServicePort,
					},
				},
			},
			Selector: map[string]string{
				"app": common.AggregatedDiscoveryServiceName,
			},
			Type: "NodePort",
		},
	}

	_, err := clientset.CoreV1().Services(namespace).Create(svc)
	if err != nil {
		fmt.Println("Error creating service: ", err)
		os.Exit(1)
	}

	args := []string{
		"--kubeconfig", "/kube/config",
		"--subscriptionID", azureSubscription,
		"--verbosity", "17",
		"--namespace", namespace,
		"--certpem", "/etc/ssl/certs/cert.pem",
		"--keypem", "/etc/ssl/certs/key.pem",
		"--rootcertpem",
		"/etc/ssl/certs/root-cert.pem",
		"--rootkeypem",
		"/etc/ssl/certs/root-key.pem",
	}

	if os.Getenv(common.IsGithubEnvVar) != "true" {
		args = append([]string{
			"--azureAuthFile", "/azure/azureAuth.json",
		}, args...)
	}

	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: meta,
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: "kubeconfig",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "kubeconfig",
							},
						},
					},
				},
				{
					Name: "azureconfig",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "azureconfig",
							},
						},
					},
				},
				{
					Name: "ca-certpemstore-ads",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "ca-certpemstore-ads",
							},
						},
					},
				},
				{
					Name: "ca-rootcertpemstore",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "ca-rootcertpemstore",
							},
						},
					},
				},

				{
					Name: "ca-keypemstore-ads",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "ca-keypemstore-ads",
							},
						},
					},
				},
				{
					Name: "ca-rootkeypemstore",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "ca-rootkeypemstore",
							},
						},
					},
				},
			},
			ImagePullSecrets: []v1.LocalObjectReference{
				{
					Name: containerRegistryCredsName,
				},
			},
			InitContainers: nil,
			Containers: []v1.Container{
				{
					Image:           fmt.Sprintf("%s/%s:latest", acr, common.AggregatedDiscoveryServiceName),
					ImagePullPolicy: "Always",
					Name:            common.AggregatedDiscoveryServiceName,
					Ports: []v1.ContainerPort{
						{
							ContainerPort: common.AggregatedDiscoveryServicePort,
							Name:          fmt.Sprintf("%s-port", common.AggregatedDiscoveryServiceName),
						},
					},
					Command: []string{
						"/ads",
					},
					Args: args,
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "kubeconfig",
							MountPath: "/kube",
						},

						{
							Name:      "azureconfig",
							MountPath: "/azure",
						},
						{
							Name:      "ca-certpemstore-ads",
							MountPath: "/etc/ssl/certs/cert.pem",
							SubPath:   "cert.pem",
							ReadOnly:  false,
						},
						{
							Name:      "ca-keypemstore-ads",
							MountPath: "/etc/ssl/certs/key.pem",
							SubPath:   "key.pem",
							ReadOnly:  false,
						},
						{
							Name:      "ca-rootkeypemstore",
							MountPath: "/etc/ssl/certs/root-key.pem",
							SubPath:   "root-key.pem",
							ReadOnly:  false,
						},
						{
							Name:      "ca-rootcertpemstore",
							MountPath: "/etc/ssl/certs/root-cert.pem",
							SubPath:   "root-cert.pem",
							ReadOnly:  false,
						},
					},
					// ReadinessProbe
					// LivenessProbe
				},
			},
		},
	}

	_, err = clientset.CoreV1().Pods(namespace).Create(pod)
	if err != nil {
		fmt.Println("Error creating pod: ", err)
		os.Exit(1)
	}

}