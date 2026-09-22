# dwz-server 架构图 / Architecture Diagrams

使用 [archify](https://github.com/tt-a1i/archify) 生成的中英文双语交互式架构图，按语言分目录存放。每张图均为独立 HTML（内置主题切换、缩放、搜索、语义视图与导出），同名 `.json` 为其规格文件。
Bilingual interactive diagrams generated with [archify](https://github.com/tt-a1i/archify), organized by language. Each diagram is a standalone HTML (theme toggle, pan/zoom, search, semantic views, export); the matching `.json` is its specification.

| 图 / Diagram | 中文 (`zh-CN/`) | English (`en/`) |
|---|---|---|
| 系统总体架构 / System architecture | [zh-CN/system-architecture.html](zh-CN/system-architecture.html) | [en/system-architecture.html](en/system-architecture.html) |
| 启动装配流程 / Startup assembly (workflow) | [zh-CN/startup-assembly.html](zh-CN/startup-assembly.html) | [en/startup-assembly.html](en/startup-assembly.html) |
| 短码跳转时序 / Redirect sequence | [zh-CN/redirect-flow.html](zh-CN/redirect-flow.html) | [en/redirect-flow.html](en/redirect-flow.html) |
| 点击统计数据流 / Click statistics dataflow | [zh-CN/click-statistics.html](zh-CN/click-statistics.html) | [en/click-statistics.html](en/click-statistics.html) |

## 事实依据 / Evidence

- 全部节点与消息均以仓库代码为据（装配链 `config/app.go`、短码分发 `config/autoload/short_code_dispatch.go`、跳转解析 `app/service/shortlink_service.go`、goose 迁移 `pkg/service/migration` 等）。
  All nodes and messages are grounded in repository code (assembly chain, short-code dispatch, redirect resolution, goose migrations, etc.).
- 架构图两份规格带有 Git 校验的源码锚点（`meta.repository`，revision `f80ca31`，local-only 链接模式，共 17 处引用）。
  The two architecture specs carry Git-verified source anchors (`meta.repository`, revision `f80ca31`, local-only link mode, 17 references).
- 每份 HTML 均通过 archify showcase 校验（9/9 检查、0 错误、0 警告）与浏览器 visual-check（1440×900 / 2048×1320，明暗双主题，无溢出）；各语言目录内 `*.visual-check.*` 为交付证据文件。
  Every HTML passed archify showcase validation (9/9 checks, 0 errors, 0 warnings) and browser visual-check (1440×900 / 2048×1320, light and dark, no overflow); the `*.visual-check.*` files inside each language folder are delivery evidence.

## 重新生成 / Regenerate

```bash
cd ~/.claude/skills/archify   # archify skill 目录 / skill directory
node bin/archify.mjs deliver <type> <spec.json> <output.html> --quality showcase --json
# 架构图需附加 / architecture specs also need:
#   --repo-root /path/to/dwz-server
```

图类型 / diagram types: `architecture` · `workflow` · `sequence` · `dataflow`
