#!/bin/sh
set -eu

password="${ROOT_PASSWORD:-alioth}"
printf 'root:%s\n' "$password" | chpasswd

cat >/etc/ssh/sshd_config.d/99-alioth.conf <<'EOF'
PermitRootLogin yes
PasswordAuthentication yes
EOF

mkdir -p /run/sshd
/usr/sbin/sshd

exec /usr/local/bin/alkaid serve --conf /etc/alioth/alkaid.yaml
