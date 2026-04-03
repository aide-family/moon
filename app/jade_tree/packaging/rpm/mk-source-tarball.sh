#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:?version required}"
ROOT="${2:?repo root required}"
OUTDIR="${3:?output dir required}"

NAME="jade-tree-${VERSION}"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

mkdir -p "${TMP}/${NAME}"
if command -v rsync >/dev/null 2>&1; then
	rsync -a \
		--exclude='.git/' \
		--exclude='rpmbuild/' \
		--exclude='bin/' \
		--exclude='*.rpm' \
		"${ROOT}/" "${TMP}/${NAME}/"
else
	(cd "${ROOT}" && tar -cf - \
		--exclude='./.git' \
		--exclude='./rpmbuild' \
		--exclude='./bin' \
		.) | tar -xf - -C "${TMP}/${NAME}"
fi

mkdir -p "${OUTDIR}"
tar -czf "${OUTDIR}/${NAME}.tar.gz" -C "${TMP}" "${NAME}"
