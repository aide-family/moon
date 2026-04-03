#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:?version required}"
ROOT="${2:?jade_tree app dir required (e.g. moon/app/jade_tree)}"
OUTDIR="${3:?output dir required}"

NAME="jade-tree-${VERSION}"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

MAGICBOX_SRC="${ROOT}/../../magicbox"
if [[ ! -d "${MAGICBOX_SRC}" ]]; then
	echo "error: magicbox not found at ${MAGICBOX_SRC} (need monorepo moon/magicbox next to moon/app/jade_tree)" >&2
	exit 1
fi

mkdir -p "${TMP}/${NAME}"
if command -v rsync >/dev/null 2>&1; then
	rsync -a \
		--exclude='.git/' \
		--exclude='rpmbuild/' \
		--exclude='bin/' \
		--exclude='*.rpm' \
		"${ROOT}/" "${TMP}/${NAME}/"
	rsync -a \
		--exclude='.git/' \
		"${MAGICBOX_SRC}/" "${TMP}/${NAME}/magicbox/"
else
	(cd "${ROOT}" && tar -cf - \
		--exclude='./.git' \
		--exclude='./rpmbuild' \
		--exclude='./bin' \
		.) | tar -xf - -C "${TMP}/${NAME}"
	mkdir -p "${TMP}/${NAME}/magicbox"
	(cd "${MAGICBOX_SRC}" && tar -cf - --exclude='./.git' .) | tar -xf - -C "${TMP}/${NAME}/magicbox"
fi

# RPM extracts a flat tree; ../../magicbox from go.mod would point outside BUILD. Bundle magicbox and point replace here.
GOMOD="${TMP}/${NAME}/go.mod"
if grep -q 'replace github.com/aide-family/magicbox => ../../magicbox' "${GOMOD}"; then
	if sed --version >/dev/null 2>&1; then
		sed -i 's|replace github.com/aide-family/magicbox => ../../magicbox|replace github.com/aide-family/magicbox => ./magicbox|' "${GOMOD}"
	else
		sed -i '' 's|replace github.com/aide-family/magicbox => ../../magicbox|replace github.com/aide-family/magicbox => ./magicbox|' "${GOMOD}"
	fi
else
	echo "error: expected go.mod replace for magicbox (../../magicbox); update mk-source-tarball.sh if layout changed)" >&2
	exit 1
fi

mkdir -p "${OUTDIR}"
tar -czf "${OUTDIR}/${NAME}.tar.gz" -C "${TMP}" "${NAME}"
