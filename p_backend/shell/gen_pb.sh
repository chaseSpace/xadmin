#!/usr/bin/env bash
set -e

_protoc_path=""
if [[ $(uname) == "Linux" ]]; then
  echo "running on Linux"
  _protoc_path='./tool/linux/protoc_v24'
  chmod +x $_protoc_path/*
elif [[ $(uname) == "Darwin" ]]; then
  echo "running on macOS"
  _protoc_path='./tool/mac/protoc_v24'
  chmod +x $_protoc_path/*
elif [[ $(uname) == *MINGW* ]]; then
  echo "running on Windows"
  _protoc_path='./tool/win/protoc_v24'
else
  echo "unknown OS"
  exit 1
fi

PATH=$PATH:$_protoc_path
OUTPUT_DIR="./proto/xadminpb"
mkdir -p $OUTPUT_DIR

echo -n "generate proto/defined/*.proto to $OUTPUT_DIR..."
_protoc_exec="$_protoc_path/protoc"
_protoc_args=(
  -I ./proto/defined/
  -I ./proto/include/
  --go_out="$OUTPUT_DIR"
  --go_opt=paths=source_relative
  --validate_out="lang=go:$OUTPUT_DIR" # 新增 validate 插件输出配置
  --validate_opt=paths=source_relative
)


rm -rf $OUTPUT_DIR/*

# 递归查找 defined 目录下的所有 .proto 文件
PROTO_FILES=()
while IFS= read -r -d '' file; do
  PROTO_FILES+=("$file")
done < <(find ./proto/defined/ -name "*.proto" -print0)

if ((${#PROTO_FILES[@]})); then
  "$_protoc_exec" "${_protoc_args[@]}" "${PROTO_FILES[@]}"
fi

echo "done."
