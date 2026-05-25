#!/bin/sh
# Install pinned Varnish 9.x from packages.varnish-software.com on Debian trixie.
# Usage: install-varnish-9.sh [minimal|tools]
#   minimal — varnishd image: varnish + varnish-modules
#   tools   — controller/exporter: varnish CLI/libs only (no vmods)
set -ex

MODE="${1:-minimal}"
VARNISH_VERSION_NUMBER="${VARNISH_VERSION_NUMBER:-9.0.3-1}"
REPO_FINGERPRINT="${REPO_FINGERPRINT:-694566269779DFAC975ED9BDD0525EAE838B3344}"

. /etc/os-release
VARNISH_VERSION="${VARNISH_VERSION_NUMBER}~${VERSION_CODENAME}"

export DEBIAN_FRONTEND=noninteractive
export DEBCONF_NONINTERACTIVE_SEEN=true

apt-get update
apt-get install -y --no-install-recommends curl ca-certificates gnupg

mkdir -p /etc/apt/keyrings
gpg --batch --keyserver hkps://keys.openpgp.org --recv-keys "${REPO_FINGERPRINT}"
gpg --batch --armor --export "${REPO_FINGERPRINT}" > /etc/apt/keyrings/varnish.gpg
echo "deb [signed-by=/etc/apt/keyrings/varnish.gpg] https://packages.varnish-software.com/varnish/${ID} ${VERSION_CODENAME} main" \
  > /etc/apt/sources.list.d/varnish.list

apt-get update

# Match official varnish/docker-varnish UID layout (not Debian stock 997).
adduser --uid 1000 --quiet --system --no-create-home --home /nonexistent --group varnish
adduser --uid 1001 --quiet --system --no-create-home --home /nonexistent --ingroup varnish vcache
adduser --uid 1002 --quiet --system --no-create-home --home /nonexistent --ingroup varnish varnishlog

if [ "${MODE}" = "minimal" ]; then
  PACKAGES="varnish=${VARNISH_VERSION} varnish-modules=${VARNISH_VERSION}"
else
  PACKAGES="varnish=${VARNISH_VERSION}"
fi

apt-get install -y --no-install-recommends ${PACKAGES}

apt-mark hold varnish
rm -rf /var/lib/apt/lists/* /etc/varnish/* ~/.gnupg
mkdir -p /etc/varnish /var/lib/varnish
chown -R varnish:varnish /etc/varnish /var/lib/varnish
mkdir -p -m 1777 /var/lib/varnish/varnishd
chown varnish /var/lib/varnish/varnishd
