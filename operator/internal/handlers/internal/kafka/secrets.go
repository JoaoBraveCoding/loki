package kafka

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	lokiv1 "github.com/grafana/loki/operator/api/loki/v1"
	"github.com/grafana/loki/operator/internal/external/k8s"
	"github.com/grafana/loki/operator/internal/manifests"
	"github.com/grafana/loki/operator/internal/status"
)

const (
	KeyReaderAddress = "readerAddress"
	KeyWriterAddress = "writerAddress"
	KeySASLMechanism = "saslMechanism"
	KeyUsername       = "username"
	KeyPassword       = "password"
)

var (
	errSecretMissingField      = errors.New("missing secret field")
	errSecretInvalidSASLConfig = errors.New("saslMechanism requires both username and password")
)

// BuildOptions validates the Kafka secret and returns KafkaOptions for config generation.
// Returns a status.DegradedError if the secret is missing or invalid.
func BuildOptions(ctx context.Context, k k8s.Client, stack *lokiv1.LokiStack) (*manifests.KafkaOptions, error) {
	secret, err := getSecret(ctx, k, stack)
	if err != nil {
		return nil, err
	}

	opts, err := extractSecrets(stack.Spec.IngestStorage.Kafka, secret)
	if err != nil {
		return nil, &status.DegradedError{
			Message: fmt.Sprintf("Invalid MSK/Kafka secret contents: %s", err),
			Reason:  lokiv1.ReasonInvalidKafkaSecret,
			Requeue: false,
		}
	}

	return opts, nil
}

func getSecret(ctx context.Context, k k8s.Client, stack *lokiv1.LokiStack) (*corev1.Secret, error) {
	var secret corev1.Secret

	key := client.ObjectKey{Name: stack.Spec.IngestStorage.Kafka.Secret.Name, Namespace: stack.Namespace}
	if err := k.Get(ctx, key, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, &status.DegradedError{
				Message: "Missing MSK/Kafka secret",
				Reason:  lokiv1.ReasonMissingKafkaSecret,
				Requeue: false,
			}
		}
		return nil, fmt.Errorf("failed to lookup MSK/Kafka secret: %w", err)
	}

	return &secret, nil
}

func extractSecrets(spec lokiv1.KafkaSpec, secret *corev1.Secret) (*manifests.KafkaOptions, error) {
	readerAddr := stringFromSecret(secret, KeyReaderAddress)
	if readerAddr == "" {
		return nil, fmt.Errorf("%w: %s", errSecretMissingField, KeyReaderAddress)
	}

	writerAddr := stringFromSecret(secret, KeyWriterAddress)
	if writerAddr == "" {
		return nil, fmt.Errorf("%w: %s", errSecretMissingField, KeyWriterAddress)
	}

	saslMechanism := stringFromSecret(secret, KeySASLMechanism)
	username := stringFromSecret(secret, KeyUsername)
	password := stringFromSecret(secret, KeyPassword)

	if saslMechanism != "" && (username == "" || password == "") {
		return nil, errSecretInvalidSASLConfig
	}

	topic := spec.Topic
	if topic == "" {
		topic = "loki"
	}

	return &manifests.KafkaOptions{
		ReaderAddress: readerAddr,
		WriterAddress: writerAddr,
		Topic:         topic,
		SASL:          saslMechanism != "",
	}, nil
}

func stringFromSecret(secret *corev1.Secret, key string) string {
	data, ok := secret.Data[key]
	if !ok || len(data) == 0 {
		return ""
	}
	return string(data)
}
