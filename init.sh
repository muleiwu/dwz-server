#!/bin/bash

# 项目模板初始化脚本
# 用于将模板项目中的模块路径替换为实际的远程仓库地址

set -e

echo "🚀 开始初始化项目..."

# 获取当前目录名作为备用
CURRENT_DIR=$(basename "$PWD")

# 尝试获取git远程仓库地址
if git remote -v >/dev/null 2>&1; then
    # 获取origin远程仓库地址
    REMOTE_URL=$(git remote get-url origin 2>/dev/null || echo "")
    
    if [ -n "$REMOTE_URL" ]; then
        echo "📡 检测到远程仓库: $REMOTE_URL"
        
        # 处理不同格式的git地址
        if [[ "$REMOTE_URL" == git@* ]]; then
            # SSH格式: git@github.com:user/repo.git -> github.com/user/repo
            NEW_MODULE_PATH=$(echo "$REMOTE_URL" | sed -E 's/git@([^:]+):/\1\//' | sed 's/\.git$//')
        elif [[ "$REMOTE_URL" == https://* ]]; then
            # HTTPS格式: https://github.com/user/repo(.git) -> github.com/user/repo
            NEW_MODULE_PATH=$(echo "$REMOTE_URL" | sed -E 's/https:\/\/([^\/]+)\/(.+)/\1\/\2/' | sed 's/\.git$//')
        else
            # 其他格式，尝试直接使用
            NEW_MODULE_PATH=$(echo "$REMOTE_URL" | sed 's/\.git$//')
        fi
    else
        echo "⚠️  未找到远程仓库地址"
        read -p "请输入新的模块路径 (例如: github.com/username/project): " NEW_MODULE_PATH
    fi
else
    echo "⚠️  当前目录不是git仓库"
    read -p "请输入新的模块路径 (例如: github.com/username/project): " NEW_MODULE_PATH
fi

if [ -z "$NEW_MODULE_PATH" ]; then
    echo "❌ 模块路径不能为空"
    exit 1
fi

echo "🔄 将模块路径替换为: $NEW_MODULE_PATH"

# 原始模块路径
OLD_MODULE_PATH="cnb.cool/mliev/examples/go-web"

# 递归查找并替换所有相关文件
echo "🔍 扫描项目文件..."

# 统计需要替换的文件数量
GO_FILES_COUNT=$(find . -name "*.go" -type f | wc -l | tr -d ' ')
MOD_FILES_COUNT=$(find . -name "go.mod" -type f | wc -l | tr -d ' ')
TOTAL_COUNT=$((GO_FILES_COUNT + MOD_FILES_COUNT))

echo "📊 发现文件:"
echo "   Go源文件: $GO_FILES_COUNT 个"
echo "   Go模块文件: $MOD_FILES_COUNT 个"
echo "   总计: $TOTAL_COUNT 个文件需要处理"

# 先检查哪些文件包含需要替换的内容
echo "🔍 检查包含模板路径的文件..."
FILES_WITH_TEMPLATE=$(grep -rl "$OLD_MODULE_PATH" . --include="*.go" --include="go.mod" 2>/dev/null || true)

if [ -z "$FILES_WITH_TEMPLATE" ]; then
    echo "✅ 没有找到需要替换的文件，可能已经初始化过了"
    exit 0
fi

echo "📝 找到以下文件需要更新:"
echo "$FILES_WITH_TEMPLATE" | while read -r file; do
    echo "   - $file"
done

# 批量替换所有 .go 文件
echo "🔄 开始替换 Go 源文件..."
find . -name "*.go" -type f -exec grep -l "$OLD_MODULE_PATH" {} \; | while read -r file; do
    echo "📝 更新文件: $file"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s|$OLD_MODULE_PATH|$NEW_MODULE_PATH|g" "$file"
    else
        # Linux
        sed -i "s|$OLD_MODULE_PATH|$NEW_MODULE_PATH|g" "$file"
    fi
done

# 替换 go.mod 文件
echo "🔄 开始替换 go.mod 文件..."
find . -name "go.mod" -type f -exec grep -l "$OLD_MODULE_PATH" {} \; | while read -r file; do
    echo "📝 更新文件: $file"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s|$OLD_MODULE_PATH|$NEW_MODULE_PATH|g" "$file"
    else
        # Linux
        sed -i "s|$OLD_MODULE_PATH|$NEW_MODULE_PATH|g" "$file"
    fi
done

# 检查替换结果
REMAINING_FILES=$(grep -rl "$OLD_MODULE_PATH" . --include="*.go" --include="go.mod" 2>/dev/null || true)
if [ -n "$REMAINING_FILES" ]; then
    echo "⚠️  以下文件可能没有完全替换成功:"
    echo "$REMAINING_FILES"
else
    echo "✅ 所有文件替换完成"
fi

echo "🧹 清理依赖..."
# 清理go模块缓存，重新下载依赖
if command -v go >/dev/null 2>&1; then
    go mod tidy
    echo "✅ Go依赖已更新"
else
    echo "⚠️  未检测到Go环境，请手动运行 'go mod tidy'"
fi

echo "🎉 项目初始化完成！"
echo "📋 摘要:"
echo "   原模块路径: $OLD_MODULE_PATH"
echo "   新模块路径: $NEW_MODULE_PATH"
echo ""
echo "🚀 你现在可以开始开发了！"
echo "   运行项目: make run 或 go run main.go"
echo "   构建项目: make build" 