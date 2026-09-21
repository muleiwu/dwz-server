# 公开首页设计

设计来源：[MasterGo「短网址」画板](https://mastergo.com/goto/WfLxXInT?page_id=M&layer_id=56:00383&file=199699779279775)，1920 × 1968，2026-09-21 经 Chrome 导出。

`templates/index.html`、`templates/home.html` 和 `static/assets/home/` 专用于公开首页，不加载或修改其他公开页的共享样式。两版的页面和资源保持一致，只有 `home_footer` 分别沿用开源版署名和企业版可配置版权规则。现有控制器的品牌名称、Logo、版权与备案字段保持不变。

## 素材

- `hero.webp`：原稿「header / 蒙版组 52」的 2 倍 PNG 导出，包含首屏插画和背景叠色，压缩至设计原尺寸 1920 × 586。
- `deployment-background.webp`：SVG 导出中 `master_svg1s_73_5744` 的嵌入原图。
- `metric-*.webp`：原稿四张指标卡中的原始图标，最长边 168px。
- `feature-*.webp`：原稿六张功能卡中的原始图标，最长边 108px。
- `logo-mark.svg`、`icon-*.svg`：从整页 SVG 导出中提取相应图形及其引用的剪裁和渐变定义。

位图使用 WebP；标题、正文、指标、编号、按钮和页脚均为 HTML。首屏插画上的示例短链是装饰性图片内容，不代表当前部署域名或可点击入口。

## 布局与兼容

桌面内容最大宽度 1400px；在 1920px 设计宽度下，首屏高 586px，指标卡向上重叠 53px，页脚高 120px。641–1024px 使用两列卡片；640px 以下功能卡单列，首屏图文上下排列。菜单支持 Enter、Tab 和 Escape；JavaScript 不可用时直接展示导航链接。

静态资源通过 `/assets/home/` 访问，沿用现有 disk / embed 加载机制。首页默认使用设计稿蓝色 M 标志；自定义 Logo 优先。站点名称及页脚文字可换行，未配置备案信息时不生成备案链接。

## 静态路由回归检查

首页资源必须挂载到 `/assets/home/`。默认目录为 `admin`、`assets`；已有仅配置 `admin` 的安装也会补充 `assets`，显式关闭静态服务仍被保留。企业版独立补充此配置，以兼容尚未包含新默认值的 CE 发布依赖。

`GOWORK=off go test . -run TestHomepageStaticResources -count=1` 使用真实应用配置装配和 go-web HTTP 服务验证 19 个资源的 GET、HEAD、字节内容及 MIME，覆盖 embed / disk、默认 / 旧配置 / 已有 assets / 单独 assets/home / 禁用静态服务，以及缺失文件的 404。独立的模板预览不能替代这项检查。
