#! /bin/bash

echo
echo ">>> Sanity checks"
echo
if [ -z "${ERIS_VERSION}" -o -z "${ERIS_RELEASE}" ]
then
    echo "The ERIS_VERSION or ERIS_RELEASE environment variables are not set, aborting"
    echo
    echo "Please start this container from the 'release.sh' script"
    exit 1
fi

export ERIS_BRANCH=master
if [ ! -z "$1" ]; then
  ERIS_BRANCH="$1"
fi

: ${CROSSPKG_GH_ACCOUNT:="eris-ltd"}
: ${CROSSPKG_ARCH:="amd64"}
: ${CROSSPKG_GOOS:="linux"}
: ${CROSSPKG_GOARCH:="amd64"}

export GOREPO=${GOPATH}/src/github.com/eris-ltd/eris-cli
git clone https://github.com/${CROSSPKG_GH_ACCOUNT}/eris-cli ${GOREPO}
pushd ${GOREPO}/cmd/eris
git fetch origin ${ERIS_BRANCH}
git checkout ${ERIS_BRANCH}
echo
echo ">>> Building the Eris ${CROSSPKG_ARCH} binary"
echo
GOOS=${CROSSPKG_GOOS} GOARCH=${CROSSPKG_GOARCH} go get
GOOS=${CROSSPKG_GOOS} GOARCH=${CROSSPKG_GOARCH} go build
popd

echo
echo ">>> Building the Debian package (#${ERIS_BRANCH})"
echo
mkdir -p deb/usr/bin deb/usr/share/doc/eris deb/DEBIAN
cp ${GOREPO}/cmd/eris/eris deb/usr/bin
cat > deb/DEBIAN/control <<EOF
Package: eris
Version: ${ERIS_VERSION}-${ERIS_RELEASE}
Section: devel
Architecture: ${CROSSPKG_ARCH}
Priority: standard
Homepage: https://docs.erisindustries.com
Maintainer: Eris Industries <support@erisindustries.com>
Build-Depends: debhelper (>= 9.1.0), golang-go (>= 1.6)
Standards-Version: 3.9.4
Description: platform for building, testing, maintaining, and operating
  distributed applications with a blockchain backend. Eris makes it easy
  and simple to wrangle the dragons of smart contract blockchains.
EOF
# TODO: manual page addition is pending the issue
# https://github.com/eris-ltd/eris-cli/issues/712.
cp ${GOREPO}/README.md deb/usr/share/doc/eris/README
cat > deb/usr/share/doc/eris/copyright <<EOF
Files: *
Copyright: $(date +%Y) Eris Industries, Ltd. <support@erisindustries.com>
License: GPL-3
EOF
dpkg-deb --build deb
PACKAGE=eris_${ERIS_VERSION}-${ERIS_RELEASE}_${CROSSPKG_ARCH}.deb
mv deb.deb ${PACKAGE}

