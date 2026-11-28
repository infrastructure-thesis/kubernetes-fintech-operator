#!/usr/bin/env bash
# Verify a release image digest, then pin it in the chart values.
set -euo pipefail

DIGEST="${1:?usage: $0 sha256:<64 hex chars>}"
REPO="ghcr.io/infrastructure-thesis/kubernetes-fintech-operator"
VALUES="charts/fintech-operator/values.yaml"

if ! [[ "$DIGEST" =~ ^sha256:[0-9a-f]{64}$ ]]; then
  echo "error: not a sha256 digest: $DIGEST" >&2
  exit 1
fi

# Refuse to pin anything that doesn't pass the Day 7 signature check.
cosign verify "$REPO@$DIGEST" \
  --certificate-identity-regexp 'https://github.com/infrastructure-thesis/kubernetes-fintech-operator/.github/workflows/release-image.yml@refs/heads/main' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com >/dev/null

sed -i "s|^  digest:.*|  digest: \"$DIGEST\"|" "$VALUES"

helm lint charts/fintech-operator
helm template fintech-operator charts/fintech-operator | grep "image:"
echo "pinned $DIGEST in $VALUES"
