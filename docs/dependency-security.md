# 依赖安全升级记录

检查日期：2026-09-21。适用范围：dwz-server、dwz-ee 及两者各自的 admin-webui 子模块。

## 更新结果

| 项目 | 升级后版本 / 结果 |
| --- | --- |
| Go 最低版本与构建工具链 | 1.26.8 |
| Node.js 构建基线 | 22.23.2 |
| pnpm | 10.10.0 → 10.34.5 |
| go-web | 1.9.0 |
| EE 引用的 dwz-server | 2.21.0 → 2.24.0 |
| golang.org/x/crypto / x/net / x/text | 0.57.0 / 0.59.0 / 0.42.0 |
| Axios / Vite / Vitest / ECharts | 1.20.0 / 7.3.6 / 4.1.11 / 6.1.0 |
| 前端完整依赖审计 | 基线 109 条 → 两版均为 0 条 |
| Go 可达漏洞 | CE 基线 11 项 → 两版均为 0 项 |
| Go 已导入包漏洞 | 两版均为 0 项 |

前端基线的严重程度为 critical 3、high 52、moderate 47、low 7。升级后使用 npm 官方审计接口检查完整锁文件，所有级别均为 0。依赖覆盖规则移入 pnpm-workspace.yaml，合并重复规则并保留必要的兼容版本段；没有配置漏洞忽略列表。

Go 1.26.8 属于当前 1.26 补丁分支，包含此前标准库安全修复，见 [Go 发布历史](https://go.dev/doc/devel/release)。pnpm 保持在受支持的 10.x 分支并升级补丁版本，见 [pnpm 支持政策](https://github.com/pnpm/pnpm/security/policy)。两版 Docker、CI 的 Go / Node / pnpm 构建约束与项目声明同步；企业版 Alpine 从 3.19 更新为 3.24.2，分支状态见 [Alpine 发布记录](https://alpinelinux.org/releases/)。Docker 构建上下文排除本地 go.work、vendor 和 node_modules，避免本地覆盖或旧依赖进入镜像。

## 兼容性调整

- 原 EE 的发布依赖缺少 IsSystemAdmin，原有 go.work 覆盖掩盖了独立构建错误。升级 CE 依赖后使用 GOWORK=off 验证。
- SQLite 测试统一使用 go-web 的 modernc 纯 Go 适配器，避免同时注册两个 sqlite 驱动，并支持 CGO_ENABLED=0。
- Vue 的 class 类型接受原生 HTMLAttributes['class']；defineModel 明确布尔类型，数组默认值使用工厂函数。
- VXE 新版配置类型让 defu 深合并推导触发 TS2589。独立 TypeScript 用例复现，旧 VXE 类型不触发；在已有类型约束的输入边界限定合并推导后，全量类型检查通过，没有禁用检查或添加 ts-ignore。
- ECharts 升到 6.1.0，通用图表使用官方 v5 主题；Vitest 4 使用现有 vitest.config.ts，删除已淘汰的 defineWorkspace 入口。参考 [ECharts 升级说明](https://echarts.apache.org/handbook/en/basics/release-note/v6-upgrade-guide/)。
- vue-tsc 保持兼容的 2.2.10，未为了清除版本提示而迁移模板类型检查方式。

## 验证

- 两版 Go 全量测试通过，包括 CGO_ENABLED=0；原生二进制可运行 --help，Linux/amd64 无 CGO 交叉编译通过。
- 两版前端 pnpm install --frozen-lockfile、check:type、生产构建通过；每版 33 个测试文件、285 个单元测试通过。
- 两版 pnpm audit 完整依赖扫描均为 0 条；govulncheck 均无可达漏洞和已导入包漏洞。
- 已用新前端构建产物刷新本地 static/admin，重新编译后端；Chrome 中管理端登录页可加载且无控制台错误。静态预览没有配套后端 API，因此此项不等于登录接口端到端验证。
- Node、Go 和 Alpine 的固定镜像标签均已通过远端 manifest 校验。完整 Docker 构建及容器 OS 包扫描未完成：本地 Docker daemon 的 info 接口返回 500。

## 剩余模块级告警

[GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932) 标记 golang.org/x/crypto/openpgp 已停止维护。项目使用 x/crypto 的其他包，没有导入 OpenPGP 包，也没有可达调用；因此 govulncheck 在模块层仍报告 1 条，在包与符号层均为 0。未压制此告警。

## 复查命令

使用项目记录的 Go / Node / pnpm 版本，在每个服务仓库分别运行：

```sh
GOWORK=off GOFLAGS=-mod=readonly go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...
GOWORK=off CGO_ENABLED=0 go test ./...
cd admin-webui
pnpm install --frozen-lockfile
pnpm run audit:security
pnpm check:type
pnpm test:unit
pnpm build:antd
```

企业版前端构建使用 `VITE_APP_EDITION=ee pnpm build:antd`。本地 go.work 是开发覆盖，不用于发布依赖验收。

## 2026-11-16：CE 管理端依赖复查

本轮仅更新当前 `dwz-server/admin-webui` 子模块及安全修复所需的兼容代码；没有重新审计 Go、企业版或容器 OS 包。继续使用 Node.js 22.23.2 和 pnpm 10.34.5。依赖解析与审计使用 npm 官方源，保留无关 catalog 项目的原有版本范围与锁定版本。

### 审计结果

| 级别 | 更新前 | 更新后 |
| --- | ---: | ---: |
| critical | 0 | 0 |
| high | 24 | 0 |
| moderate | 33 | 0 |
| low | 15 | 1 |
| 合计 | 72 | 1 |

剩余告警为 [GHSA-gcqx-w48m-w2h4](https://github.com/advisories/GHSA-gcqx-w48m-w2h4)：deepmerge 4.3.1 尚无上游修复版本，本轮使用 pnpm 本地补丁修复。npm 审计按发布版本匹配，无法识别本地补丁，因此 `pnpm run audit:security` 仍返回退出码 1。没有加入忽略列表，也不将这一结果记为“零漏洞”。

### 实际锁定的升级版本

| 依赖 | 本轮锁定版本 |
| --- | --- |
| Vite / Vitest（含内部包） | 7.3.8 / 4.1.12 |
| Happy DOM | 21.0.4 |
| Zod / zod-defaults | 4.6.7 / 0.2.3 |
| vue-i18n / @intlify/core-base / @intlify/shared | 11.4.20 |
| qs / follow-redirects | 6.17.1 / 1.16.3 |
| node-forge / undici | 1.4.3 / 7.30.2 |
| serialize-javascript / source-map-js | 7.1.4 / 1.2.2 |
| fast-uri / postcss-selector-parser 6.x | 3.1.9 / 6.1.6 |
| 受影响的 glob / minimatch | 14.0.6 / 11.0.8 |

版本范围继续使用兼容的 caret 范围，部分依赖解析到计划最低修复版本之后的兼容版本；实际安装结果由锁文件固定。minimatch 的未受本轮告警影响的 3.x、5.x 分支保留原版本。

### 兼容性与本地补丁

- 表单规则迁移到 Zod 4，移除只支持 Zod 3 的 `@vee-validate/zod`，保留 vee-validate 4.15.1。新的内部 TypedSchema 适配器处理异步校验、转换、默认值、嵌套数组及 union 错误路径，遵循 [Zod 4 迁移说明](https://zod.dev/v4/changelog)。
- Vue 深层响应式代理不能代理 Zod 4 的不可配置内部属性。初始化、规则解包、必填判断和校验入口统一通过 `toRaw` 获取原始 schema；响应式及 readonly schema 均有回归测试。原有表单初始化行为保持不变。
- `admin-webui/patches/deepmerge@4.3.1.patch` 修复源码、CommonJS 和 UMD 入口，阻止目标复制和源合并过程中的 `__proto__`、`constructor`、`prototype` 键。测试覆盖空原型对象、原型占位值、嵌套合并、数组及 Symbol 属性。
- minimatch 11 改为具名导出；`admin-webui/patches/depcheck@1.4.7.patch` 调整 depcheck 的文件匹配与 gulp 插件匹配入口。回归测试使用临时项目验证导入和插件检测。
- 两个补丁均登记在 `pnpm-workspace.yaml` 的 `patchedDependencies`，补丁哈希进入锁文件；不直接修改安装目录。移除补丁前，需升级到包含对应修复的上游版本并重新执行回归测试及完整审计。

### 本轮验证

- 升级前 33 个测试文件、285 个单元测试通过；升级后 38 个测试文件、311 个测试通过。
- `pnpm check:type`、`pnpm build:antd`、修改源码的 ESLint 检查通过。
- `pnpm check:dep` 正常完成，未报告缺失或未使用依赖。
- 新增测试验证 archiver 文件匹配和 Workbox manifest 生成，覆盖默认生产构建未启用的可选插件路径。
- Chromium 实测开发服务和生产预览：登录输入框正常渲染，空表单提示正确，填写后错误提示消失，没有意外控制台错误。站点品牌和登录选项 GET 接口使用模拟响应，未发送登录请求；此项不等于后端认证端到端验证。
- `pnpm install --frozen-lockfile` 通过，安装前后的 workspace 配置和锁文件哈希一致，两个补丁的登记与哈希通过锁文件校验。
