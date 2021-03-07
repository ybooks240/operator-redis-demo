package controllers

import (
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"

	devv1 "github.com/ybooks240/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	JamesRedisCommonKey = "ybooks240.github.com/jamesRedis"
	JamesRedisLabelKey  = "jamesRedis"
	RedisDataDirName    = "redisdatadir"
)

func MutateService(jamesRedis *devv1.JamesRedis, svc *corev1.Service) {
	// labels
	svc.Labels = map[string]string{
		JamesRedisCommonKey: "redis",
	}
	svc.Spec = corev1.ServiceSpec{
		//Type:      corev1.ServiceTypeClusterIP,
		//Type: corev1.ServiceTypeNodePort,
		ClusterIP: corev1.ClusterIPNone,
		//Type: corev1.ServiceTypeClusterIP,
		//ClusterIP: "10.99.77.135",
		//Type: corev1.ServiceTypeClusterIP,
		SessionAffinity: corev1.ServiceAffinityNone,
		Selector: map[string]string{
			JamesRedisLabelKey: jamesRedis.Name,
		},
		Ports: []corev1.ServicePort{
			corev1.ServicePort{
				Protocol: "TCP",
				Port:     6379,
				//NodePort:   30008,
				TargetPort: intstr.FromInt(6379),
				Name:       "redis-service",
			},
		},
	}
}

// 根据JamesRedis组装statefulset
func MutateStatefulSet(jamesRedis *devv1.JamesRedis, sts *appsv1.StatefulSet) {
	// labels
	sts.Labels = map[string]string{
		JamesRedisCommonKey: "redis",
	}
	// spec
	// containers
	sts.Spec = appsv1.StatefulSetSpec{
		Replicas:    jamesRedis.Spec.Replicas,
		ServiceName: jamesRedis.Name, // headless svc name
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				JamesRedisCommonKey: jamesRedis.Name,
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					JamesRedisCommonKey: jamesRedis.Name,
					JamesRedisLabelKey:  "redis",
				},
			},
			Spec: corev1.PodSpec{
				Containers: newContainers(jamesRedis),
				Volumes: []corev1.Volume{
					{
						Name: jamesRedis.Spec.ConfigMapName,
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: jamesRedis.Spec.ConfigMapName,
								},
							},
						},
					},
				},
			},
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: RedisDataDirName,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("100Mi"),
						},
					},
				},
			},
		},
	}
}

func newContainers(jamesRedis *devv1.JamesRedis) []corev1.Container {
	return []corev1.Container{
		corev1.Container{
			Name:            "jamesredis",
			Image:           jamesRedis.Spec.Image,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Ports: []corev1.ContainerPort{
				corev1.ContainerPort{
					Name:          "peer",
					ContainerPort: 6379,
				},
			},
			Env: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "INITIAL_CLUSTER_SIZE",
					Value: strconv.Itoa(int(*jamesRedis.Spec.Replicas)),
				},
				corev1.EnvVar{
					Name:  "SET_NAME",
					Value: jamesRedis.Name,
				},
				corev1.EnvVar{
					Name: "MY_NAMESPACE",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.namespace",
						},
					},
				},
				corev1.EnvVar{
					Name: "POD_IP",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "status.podIP",
						},
					},
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				corev1.VolumeMount{
					Name:      RedisDataDirName,
					MountPath: "/data/",
				},
				corev1.VolumeMount{
					Name:      jamesRedis.Spec.ConfigMapName,
					MountPath: "/redis/",
				},
			},
			Command: []string{
				"redis-server", "/redis/redis.conf",
			},
		},
	}
}
