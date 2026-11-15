#!/usr/bin/env bash
# Generator sertifikat TLS CA lokal, Caddy (erp.localhost), dan Redis TLS untuk pengembangan.
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
CERT_DIR="${DEPLOY_DIR}/certs"
FORCE="${1:-}"

mkdir -p "${CERT_DIR}"

if [[ "${FORCE}" != "--force" ]] \
  && [[ -f "${CERT_DIR}/ca.crt" ]] \
  && [[ -f "${CERT_DIR}/erp.localhost.crt" ]] \
  && [[ -f "${CERT_DIR}/erp.localhost.key" ]] \
  && [[ -f "${CERT_DIR}/redis.crt" ]] \
  && [[ -f "${CERT_DIR}/redis.key" ]]; then
  echo "Sertifikat TLS pengembangan sudah tersedia di ${CERT_DIR} (gunakan --force untuk regenerasi)."
  exit 0
fi

# 1. Buat CA root pengembangan lokal bila belum ada atau dipaksa ulang
if [[ "${FORCE}" == "--force" ]] || [[ ! -f "${CERT_DIR}/ca.key" ]] || [[ ! -f "${CERT_DIR}/ca.crt" ]]; then
  openssl genrsa -out "${CERT_DIR}/ca.key" 4096
  openssl req -x509 -new -nodes \
    -key "${CERT_DIR}/ca.key" \
    -sha256 -days 3650 \
    -subj "/C=ID/O=IntegriOne Dev/CN=IntegriOne Dev Root CA" \
    -addext "basicConstraints=critical,CA:TRUE,pathlen:0" \
    -addext "keyUsage=critical,keyCertSign,cRLSign" \
    -out "${CERT_DIR}/ca.crt"
fi

generate_leaf_cert() {
  local name="$1"
  local cn="$2"
  local san="$3"

  local ext_file="${CERT_DIR}/${name}.ext.cnf"
  cat > "${ext_file}" <<EOF
basicConstraints=CA:FALSE
keyUsage=critical,digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=${san}
EOF

  openssl genrsa -out "${CERT_DIR}/${name}.key" 2048
  openssl req -new \
    -key "${CERT_DIR}/${name}.key" \
    -subj "/C=ID/O=IntegriOne Dev/CN=${cn}" \
    -out "${CERT_DIR}/${name}.csr"

  openssl x509 -req \
    -in "${CERT_DIR}/${name}.csr" \
    -CA "${CERT_DIR}/ca.crt" \
    -CAkey "${CERT_DIR}/ca.key" \
    -CAcreateserial \
    -out "${CERT_DIR}/${name}.crt" \
    -days 825 \
    -sha256 \
    -extfile "${ext_file}"

  rm -f "${CERT_DIR}/${name}.csr" "${ext_file}"
}

# 2. Sertifikat server HTTPS lokal untuk Caddy (erp.localhost)
generate_leaf_cert \
  "erp.localhost" \
  "erp.localhost" \
  "DNS:erp.localhost,DNS:localhost,DNS:caddy,IP:127.0.0.1,IP:::1"

# 3. Sertifikat server TLS internal untuk Redis keamanan (IAM & Gateway)
generate_leaf_cert \
  "redis" \
  "redis" \
  "DNS:redis,DNS:localhost,DNS:iam-redis,IP:127.0.0.1"

# Izinkan container non-root membaca leaf cert/key pada environment dev lokal
chmod 600 "${CERT_DIR}/ca.key"
chmod 644 "${CERT_DIR}/ca.crt" \
  "${CERT_DIR}/erp.localhost.crt" "${CERT_DIR}/erp.localhost.key" \
  "${CERT_DIR}/redis.crt" "${CERT_DIR}/redis.key"
rm -f "${CERT_DIR}/ca.srl"

echo "Sertifikat TLS pengembangan berhasil dibuat di ${CERT_DIR}:"
openssl x509 -in "${CERT_DIR}/erp.localhost.crt" -noout -subject -ext subjectAltName
openssl x509 -in "${CERT_DIR}/redis.crt" -noout -subject -ext subjectAltName
