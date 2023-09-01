FROM --platform=arm64 gcr.io/bazel-public/bazel

ARG PORT
ARG HOST
ARG CERT_FILE
ARG KEY_FILE
WORKDIR /workspace
COPY WORKSPACE.bazel .bazelrc /workspace/
COPY cc /workspace/cc
COPY proto /workspace/proto
COPY cert /workspace/cert
