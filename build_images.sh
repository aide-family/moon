#!/bin/bash
VERSION=$(git describe --tags --always)
echo $VERSION
# 可执行文件列表
executables=("palace" "rabbit" "houyi")

# 遍历每个可执行文件并构建镜像
for exec in "${executables[@]}"; do
  # 设置镜像名称，可以根据需要修改
  image_name="$1/${exec}:$VERSION"
  # 构建 Docker 镜像
  docker build --build-arg CMD_PARAMS=${exec} -t ${image_name} .
  # 打一个latest版本
  docker tag ${image_name} "$1/${exec}:latest"
  # 推送到 Docker 仓库
  docker push ${image_name}
  docker push "$1/${exec}:latest"
  echo "Built ${image_name} for executable ${exec}"
done
