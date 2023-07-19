# syntax=docker/dockerfile:experimental
# Copyright 2018 Google, Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM centos:7 as build

RUN yum install -y rpm-build make
# This arg is passed by docker buildx & contains the target CPU architecture (e.g., amd64, arm64, etc.)
ARG TARGETARCH
ENV GOPROXY https://goproxy.cn
ENV GOARCH=$TARGETARCH
ENV CGO_ENABLED=0
ENV GOLANG_VERSION 1.19
ENV sslverify=false

RUN curl -sSL https://dl.google.com/go/go${GOLANG_VERSION}.linux-amd64.tar.gz \
    | tar -C /usr/local -xz
ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN echo $TARGETARCH
WORKDIR /src

# Add .docker config dir
RUN mkdir -p /vgpuconfig
COPY . .
RUN \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    make GOARCH=$TARGETARCH

FROM centos:7
COPY --from=0 /src/start /vgpuconfig/start
ENV HOME /root
ENV USER root
ENV PATH $PATH:/vgpuconfig
WORKDIR /workspace
# todo ENTRYPOINT 一定会被执行  但是cmd 会被docker run 后面的覆盖
ENTRYPOINT ["/vgpuconfig/start"]
