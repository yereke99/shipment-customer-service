#!/usr/bin/env bash
set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-build/docker-compose.yml}"
BASE_URL="${BASE_URL:-http://localhost:8080}"
IDN="${IDN:-990101123456}"
ROUTE="${ROUTE:-ALMATY->ASTANA}"
PRICE="${PRICE:-120000}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

fail() {
  echo "FAIL: $1"
  exit 1
}

pass() {
  echo "PASS: $1"
}

extract_json_value() {
  local file="$1"
  local key="$2"
  sed -n "s/.*\"${key}\":\"\([^\"]*\)\".*/\1/p" "$file" | head -n1
}

payload="{\"route\":\"${ROUTE}\",\"price\":${PRICE},\"customer\":{\"idn\":\"${IDN}\"}}"

post_code="$(curl -s -o "${TMP_DIR}/post.json" -w "%{http_code}" -X POST "${BASE_URL}/api/v1/shipments" -H "Content-Type: application/json" -d "${payload}" || true)"
[[ "${post_code}" == "201" ]] || fail "POST returned ${post_code}"
shipment_id="$(extract_json_value "${TMP_DIR}/post.json" "id")"
customer_id="$(extract_json_value "${TMP_DIR}/post.json" "customerId")"
[[ -n "${shipment_id}" ]] || fail "POST has no shipment id"
[[ -n "${customer_id}" ]] || fail "POST has no customer id"
pass "server start and curl post"

if docker compose -f "${COMPOSE_FILE}" exec -T postgres bash -lc 'cat < /dev/null > /dev/tcp/envoy/9090' >/dev/null 2>&1; then
  pass "grpc port 9090 works inside docker net"
else
  fail "grpc port 9090 is not reachable inside docker net"
fi

if docker compose -f "${COMPOSE_FILE}" port envoy 9090 >/dev/null 2>&1; then
  fail "grpc port 9090 is exposed outside"
else
  pass "grpc port 9090 is not exposed outside"
fi

get_code="$(curl -s -o "${TMP_DIR}/get.json" -w "%{http_code}" "${BASE_URL}/api/v1/shipments/${shipment_id}" || true)"
[[ "${get_code}" == "200" ]] || fail "GET returned ${get_code}"
grep -q "\"id\":\"${shipment_id}\"" "${TMP_DIR}/get.json" || fail "GET response id mismatch"
grep -q "\"customerId\":\"${customer_id}\"" "${TMP_DIR}/get.json" || fail "GET response customer id mismatch"
pass "server start and curl get"

echo "E2E CURL TEST PASSED"
