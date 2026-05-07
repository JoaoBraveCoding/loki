package manifests

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	envKafkaSASLUsername = "KAFKA_SASL_USERNAME"
	envKafkaSASLPassword = "KAFKA_SASL_PASSWORD"

	kafkaKeyUsername = "username"
	kafkaKeyPassword = "password"
)

func configureDeploymentForKafka(d *appsv1.Deployment, secretName string) {
	envVars := kafkaEnvVars(secretName)
	for i := range d.Spec.Template.Spec.Containers {
		d.Spec.Template.Spec.Containers[i].Env = append(d.Spec.Template.Spec.Containers[i].Env, envVars...)
	}
}

func configureStatefulSetForKafka(s *appsv1.StatefulSet, secretName string) {
	envVars := kafkaEnvVars(secretName)
	for i := range s.Spec.Template.Spec.Containers {
		s.Spec.Template.Spec.Containers[i].Env = append(s.Spec.Template.Spec.Containers[i].Env, envVars...)
	}
}

func kafkaEnvVars(secretName string) []corev1.EnvVar {
	return []corev1.EnvVar{
		envVarFromKafkaSecret(envKafkaSASLUsername, secretName, kafkaKeyUsername),
		envVarFromKafkaSecret(envKafkaSASLPassword, secretName, kafkaKeyPassword),
	}
}

func envVarFromKafkaSecret(envName, secretName, secretKey string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: envName,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
				Key: secretKey,
			},
		},
	}
}
