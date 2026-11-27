#!/usr/bin/env bash
# Membuat sertifikat development khusus jalur mandat IAM dan HRM.
set -euo pipefail
umask 077

if [[ $# -ne 1 || -z "$1" ]]; then
  echo "Penggunaan: $0 <direktori-baru-dan-diabaikan-git>" >&2
  exit 2
fi
output="$1"
if [[ -e "$output" ]]; then
  echo "Direktori tujuan sudah ada." >&2
  exit 2
fi
mkdir -p -- "$(dirname -- "$output")"
stage="$(mktemp -d "$(dirname -- "$output")/.mandate-certs.XXXXXX")"
trap 'rm -rf -- "$stage"' EXIT

openssl req -new -x509 -newkey ec -pkeyopt ec_paramgen_curve:P-256 -nodes   -days 30 -subj '/CN=IntegriOne requisition mandate development CA'   -addext 'basicConstraints=critical,CA:TRUE,pathlen:0'   -addext 'keyUsage=critical,keyCertSign,cRLSign'   -keyout "$stage/ca.key" -out "$stage/ca.crt" 2>/dev/null

issue() {
  local name="$1" usage="$2" san="$3" serial="$4"
  printf 'basicConstraints=critical,CA:FALSE\nkeyUsage=critical,digitalSignature\nextendedKeyUsage=%s\nsubjectAltName=%s\n'     "$usage" "$san" >"$stage/leaf.ext"
  openssl req -new -newkey ec -pkeyopt ec_paramgen_curve:P-256 -nodes     -subj "/CN=$name" -keyout "$stage/$name.key" -out "$stage/leaf.csr" 2>/dev/null
  openssl x509 -req -in "$stage/leaf.csr" -CA "$stage/ca.crt" -CAkey "$stage/ca.key"     -set_serial "$serial" -days 30 -sha256 -extfile "$stage/leaf.ext" -out "$stage/$name.crt" 2>/dev/null
}

issue iam-mandate serverAuth DNS:iam 1
issue hrm-mandate clientAuth URI:spiffe://integrione/human-resource 2
rm -- "$stage/ca.key" "$stage/leaf.csr" "$stage/leaf.ext"
# Direktori host tetap privat. Berkas leaf di-bind satu per satu ke kontainer non-root.
chmod 0644 "$stage/"*.crt "$stage/"*.key
mv -T -- "$stage" "$output"
