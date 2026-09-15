#!/bin/sh
set -e

CERT_DIR="${CERT_DIR:-/data}"
CERT_FILE="${TLS_CERT:-$CERT_DIR/cert.pem}"
KEY_FILE="${TLS_KEY:-$CERT_DIR/key.pem}"
PORT="${PORT:-8443}"

mkdir -p "$(dirname "$CERT_FILE")"

if [ ! -f "$CERT_FILE" ] || [ ! -f "$KEY_FILE" ]; then
  echo "Generating self-signed TLS certificate..."
  HOST="${PUBLIC_HOST:-localhost}"

  SAN="DNS:localhost,IP:127.0.0.1"
  case "$HOST" in
    *[0-9]*.*[0-9]*.*[0-9]*.*[0-9]*)
      SAN="${SAN},IP:${HOST}"
      CN="qnotes"
      ;;
    *)
      SAN="${SAN},DNS:${HOST}"
      CN="$HOST"
      ;;
  esac

  openssl req -x509 -newkey rsa:2048 -nodes \
    -keyout "$KEY_FILE" \
    -out "$CERT_FILE" \
    -days 365 \
    -subj "/CN=${CN}" \
    -addext "subjectAltName=${SAN}"
  echo "⚠ Самоподписанный сертификат создан. Браузер покажет предупреждение о безопасности."
fi

export TLS_CERT="$CERT_FILE"
export TLS_KEY="$KEY_FILE"
export PORT

# Drop privileges when started as root
if [ "$(id -u)" = "0" ]; then
  chown -R qnotes:qnotes "$(dirname "$CERT_FILE")" 2>/dev/null || true
  exec su -s /bin/sh qnotes -c "exec /app/qnotes"
fi

exec /app/qnotes
