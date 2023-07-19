#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset
declare -i t1 t2
t1=$(date +"%s")
cp  /opt/aps/workdir/ttt1.csv /tmp/kk
t2=$(date +"%s")
echo $t2-$t1

ROOT=$(cd $(dirname ${BASH_SOURCE[0]})/.. && pwd -P)

source "${ROOT}/hack/common.sh"

function plugin::build() {
  (
    for arg; do
        case $arg in
        img)
            plugin::generate_img
            ;;
        lib)
            plugin::build_binary
        esac
    done
  )
}


function plugin::build_binary() {
  go build -o "${ROOT}/go/bin/gpu-$arg" -ldflags "$(plugin::version::ldflags) -s -w" ${PACKAGE}/cmd/$arg
}

function plugin::generate_img() {
  readonly local commit=$(git log --no-merges --oneline | wc -l | sed -e 's,^[ \t]*,,')
  readonly local version=$(<"${ROOT}/VERSION")
  readonly local base_img=${BASE_IMG:-"gpu-config:v1"}

  mkdir -p "${ROOT}/go/build"
  tar czf "${ROOT}/go/build/gpu-manager-source.tar.gz" --transform 's,^,/gpu-config-'${version}'/,' $(plugin::source_targets)

  cp -R "${ROOT}/build/"* "${ROOT}/go/build/"

  (
    cd ${ROOT}/go/build
    docker build \
        --network=host \
        --build-arg version=${version} \
        --build-arg commit=${commit} \
        --build-arg base_img=${base_img} \
        -t "${IMAGE_FILE}:${version}" .
  )
}

plugin::build "$@"
