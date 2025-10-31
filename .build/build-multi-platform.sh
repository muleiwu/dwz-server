#!/bin/bash

# 定义目标平台列表
platforms=(
    "linux/amd64"
    "linux/amd64/v2"
    "linux/amd64/v3"
    "linux/arm64"
    "linux/riscv64"
    "linux/ppc64le"
    "linux/s390x"
    "linux/386"
    "linux/mips64le"
    "linux/mips64"
    "linux/loong64"
    "linux/arm/v7"
    "linux/arm/v6"
)

# 创建输出目录
output_dir="build"
mkdir -p "$output_dir"

# 遍历所有平台进行编译
for platform in "${platforms[@]}"; do
    # 解析平台参数
    IFS='/' read -ra parts <<< "$platform"
    GOOS="${parts[0]}"
    GOARCH="${parts[1]}"
    GOARM=""
    suffix=""

    # 处理特殊架构参数
    if [[ $GOARCH == arm* ]] && [ -n "${parts[2]}" ]; then
        GOARM="${parts[2]/v/}"  # 提取ARM版本号（去除'v'前缀）
        suffix="-${parts[2]}"   # 文件名后缀
    elif [ -n "${parts[2]}" ]; then
        suffix="-${parts[2]}"   # 非ARM架构的后缀（如v2/v3）
    fi

    # 生成输出文件名
    output_name="main-${GOOS}-${GOARCH}${suffix}"
    output_path="$output_dir/$output_name"

    # 设置编译环境变量
    export GOOS GOARCH CGO_ENABLED=0
    if [ -n "$GOARM" ]; then
        export GOARM
    fi

    # 执行编译
    echo "🛠️  正在编译: $platform → $output_name"
    if ! go build -o "$output_path" main.go 2>/dev/null; then
        echo "❌ 编译失败: $platform (可能不受支持)"
        rm -f "$output_path"  # 删除空输出文件
    fi
done

echo -e "\n✅ 编译完成！输出文件位于: $output_dir/"