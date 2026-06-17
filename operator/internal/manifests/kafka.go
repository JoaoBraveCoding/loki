package manifests

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	envKafkaSASLUsername = "KAFKA_SASL_USERNAME"
	envKafkaSASLPassword = "KAFKA_SASL_PASSWORD"
)

func configureDeploymentForKafka(d *appsv1.Deployment, opts *KafkaOptions) {
	envVars := kafkaEnvVars(opts)
	for i := range d.Spec.Template.Spec.Containers {
		d.Spec.Template.Spec.Containers[i].Env = append(d.Spec.Template.Spec.Containers[i].Env, envVars...)
	}
}

func configureStatefulSetForKafka(s *appsv1.StatefulSet, opts *KafkaOptions) {
	envVars := kafkaEnvVars(opts)
	for i := range s.Spec.Template.Spec.Containers {
		s.Spec.Template.Spec.Containers[i].Env = append(s.Spec.Template.Spec.Containers[i].Env, envVars...)
	}
}

func kafkaEnvVars(opts *KafkaOptions) []corev1.EnvVar {
	return []corev1.EnvVar{
		envVarFromSecretRef(envKafkaSASLUsername, opts.SASLUsername),
		envVarFromSecretRef(envKafkaSASLPassword, opts.SASLPassword),
	}
}

func envVarFromSecretRef(envName string, ref SecretRef) corev1.EnvVar {
	return corev1.EnvVar{
		Name: envName,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: ref.SecretName,
				},
				Key: ref.Key,
			},
		},
	}
}
