#!/usr/bin/env bash
#
# Creates the Kafka secret in the LokiStack namespace from the Kafka CR status.
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
SECRET_NAME="kafka-credentials"

echo "Waiting for Kafka cluster '${KAFKA_CLUSTER}' to be ready..."
until oc get kafka "${KAFKA_CLUSTER}" -n "${LOKISTACK_NS}" -o jsonpath='{.status.listeners}' 2>/dev/null | grep -q bootstrapServers; do
  sleep 2
done

BOOTSTRAP="$( oc get kafka "${KAFKA_CLUSTER}" -n "${LOKISTACK_NS}" \
  -o jsonpath='{.status.listeners[?(@.name=="plain")].bootstrapServers}' )"

echo "Bootstrap servers: ${BOOTSTRAP}"
echo "Creating secret '${SECRET_NAME}' in namespace '${LOKISTACK_NS}'..."

oc create secret generic "${SECRET_NAME}" \
  -n "${LOKISTACK_NS}" \
  --from-literal=readerAddress="${BOOTSTRAP}" \
  --from-literal=writerAddress="${BOOTSTRAP}" \
  --dry-run=client -o yaml | oc apply -f -

echo ""
echo "Done. Use the following in your LokiStack CR:"
echo ""
echo "  ingestStorage:"
echo "    kafka:"
echo "      topic: loki"
echo "      secret:"
echo "        name: ${SECRET_NAME}"
