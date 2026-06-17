package kafka

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	lokiv1 "github.com/grafana/loki/operator/api/loki/v1"
	"github.com/grafana/loki/operator/internal/external/k8s"
	"github.com/grafana/loki/operator/internal/manifests"
	"github.com/grafana/loki/operator/internal/status"
)

// BuildOptions validates the Kafka configuration and returns KafkaOptions for config generation.
// Returns a status.DegradedError if a referenced secret is missing or invalid.
func BuildOptions(ctx context.Context, k k8s.Client, stack *lokiv1.LokiStack) (*manifests.KafkaOptions, error) {
	spec := stack.Spec.IngestStorage.Kafka

	topic := spec.Topic
	if topic == "" {
		topic = "loki"
	}

	metadataTopic := spec.MetadataTopic
	if metadataTopic == "" {
		metadataTopic = topic + "-metadata"
	}

	opts := &manifests.KafkaOptions{
		ReaderAddress: spec.ReaderAddress,
		WriterAddress: spec.WriterAddress,
		Topic:         topic,
		MetadataTopic: metadataTopic,
	}

	if spec.Authentication != nil {
		auth := spec.Authentication

		username, err := resolveSecretReference(ctx, k, stack.Namespace, auth.Username)
		if err != nil {
			return nil, err
		}

		password, err := resolveSecretReference(ctx, k, stack.Namespace, auth.Password)
		if err != nil {
			return nil, err
		}

		opts.SASL = true
		opts.SASLMechanism = string(auth.SASLMechanism)
		opts.SASLUsername = manifests.SecretRef{
			SecretName: auth.Username.SecretName,
			Key:        auth.Username.Key,
			Value:      username,
		}
		opts.SASLPassword = manifests.SecretRef{
			SecretName: auth.Password.SecretName,
			Key:        auth.Password.Key,
			Value:      password,
		}
	}

	return opts, nil
}

func resolveSecretReference(ctx context.Context, k k8s.Client, namespace string, ref *lokiv1.SecretReference) (string, error) {
	var secret corev1.Secret
	key := client.ObjectKey{Name: ref.SecretName, Namespace: namespace}

	if err := k.Get(ctx, key, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			return "", &status.DegradedError{
				Message: fmt.Sprintf("Missing Kafka authentication secret: %s", ref.SecretName),
				Reason:  lokiv1.ReasonMissingKafkaSecret,
				Requeue: false,
			}
		}
		return "", fmt.Errorf("failed to lookup Kafka secret %s: %w", ref.SecretName, err)
	}

	data, ok := secret.Data[ref.Key]
	if !ok || len(data) == 0 {
		return "", &status.DegradedError{
			Message: fmt.Sprintf("Kafka secret %s is missing key %s", ref.SecretName, ref.Key),
			Reason:  lokiv1.ReasonInvalidKafkaSecret,
			Requeue: false,
		}
	}

	return string(data), nil
}
