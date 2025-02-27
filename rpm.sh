#!/usr/bin/env bash
# init env
rm -rf ~/rpmbuild/

# env variable
VERSION=$(head -n 1 VERSION)
BUILD_NUMBER=$(git rev-parse --short HEAD)

# set up directories
pushd ~
mkdir rpmbuild
cd rpmbuild
mkdir BUILD RPMS SOURCES SPECS SRPMS
popd

# tarball files
mkdir ~/source
cp -r ./api ./cmd ./internal ./configs ./init go.mod go.sum LICENSE ~/source
pushd ~
tar -cvzf "cube-cos-api-${VERSION}.tar.gz" source
mv "cube-cos-api-${VERSION}.tar.gz" ~/rpmbuild/SOURCES
rm -r source/
popd

# copy the rpm spec file
cp ./cube-cos-api.spec ~/rpmbuild/SPECS

# build the rpm
rpmbuild -bb --nodeps --define "version $VERSION" --define "build_number $BUILD_NUMBER" ~/rpmbuild/SPECS/cube-cos-api.spec

# check
ls -ahl ~/rpmbuild/RPMS/x86_64/cube-cos-api-$VERSION-1.el9.$BUILD_NUMBER.x86_64.rpm
