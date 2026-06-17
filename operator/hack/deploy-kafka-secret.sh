#!/usr/bin/env bash
#
# Prints the Kafka bootstrap address from the Kafka CR status for use in the
# LokiStack CR spec.ingestStorage.kafka.readerAddress / writerAddress fields.
#
# Prerequisites:
#   1. Streams for Apache Kafka operator installed via OLM
#   2. hack/addons_kafka_ocp.yaml applied and Kafka cluster ready
#
# Usage:
#   ./hack/deploy-kafka-secret.sh [namespace]

set -euo pipefail

LOKISTACK_NS="${1:-openshift-logging}"
KAFKA_CLUSTER="loki-kafka"

echo "Waiting for Kafka cluster '${KAFKA_CLUSTER}' to be ready..."
until oc get kafka "${KAFKA_CLUSTER}" -n "${LOKISTACK_NS}" -o jsonpath='{.status.listeners}' 2>/dev/null | grep -q bootstrapServers; do
  sleep 2
done

BOOTSTRAP="$( oc get kafka "${KAFKA_CLUSTER}" -n "${LOKISTACK_NS}" \
  -o jsonpath='{.status.listeners[?(@.name=="plain")].bootstrapServers}' )"

echo "Bootstrap servers: ${BOOTSTRAP}"
echo ""
echo "Done. Use the following in your LokiStack CR:"
echo ""
echo "  ingestStorage:"
echo "    kafka:"
echo "      topic: loki"
echo "      readerAddress: ${BOOTSTRAP}"
echo "      writerAddress: ${BOOTSTRAP}"
