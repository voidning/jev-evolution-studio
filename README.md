# Jev 自然语言源码编辑器

选择真实 React/Vite 页面中的元素，输入自然语言，查看类型化意图和源码 Diff，**Accept 后写入文件并等待 Vite 刷新**。Reject 不写文件，Undo 恢复最近一次修改前的完整源码。Tailwind 项目修改 JSX className / AST；普通 CSS 项目继续写 `src/jev-edits.css`。

Jev 只在有限的 Choice / Noul 候选中判断意图，不生成或执行代码。没有 API Key 时，所有明确命令和示例场景都可离线完成。旧候选采样、Critic 和静态生成舞台已移出主程序；`web-prototype` 未改动。

## 产品理念

**点击网页上的元素，说出修改，把可审查的变化写回真实源码。**

自然语言负责表达意图，程序负责保证修改的边界。Jev 是快速、类型化的意图判断器；它只能在注册过的候选中选择，不生成 Tailwind 类、CSS、JSX 或可执行命令。明确命令先由本地规则解析，模糊语义才交给 Jev；没有 API Key 也能完成主要编辑流程。

我们优先追求以下体验：

- **快且可预测**：常见操作走固定规则与执行器，不做多轮生成、候选页面采样或 Critic 进化。
- **精确作用域**：选中的 DOM 元素必须映射到可验证的源码节点；只改指定属性、状态和断点，保留内容与业务行为。
- **源码是结果**：修改落在项目的 JSX 或 CSS 文件中，刷新和重启后仍存在，可以正常提交 Git。
- **先审查，再写入**：先生成内存 Patch 和真实 Diff，Accept 后才写文件；Reject 不写，Undo 恢复原始字节。
- **边界明确**：无法可靠定位、存在结构歧义、动态表达过于复杂或置信度不足时直接说明原因，不猜测性改代码。

第一阶段把普通 CSS 修改集中写入 `src/jev-edits.css`，先跑通选择、写回、HMR 与撤销；当前阶段在统一意图协议上增加 Tailwind 属性组和 JSX AST 执行器。保留兼容路径，逐步扩大可验证的能力，而非一次支持任意项目。

## 当前规模与成熟度

当前是**单用户、本机运行、显式接入项目的可用 MVP**，不是通用自然语言编程平台，也不宣称已有规模化用户或生产部署验证。当前实现保存在 `v2` 分支，仓库中 `web-prototype/` 是保留的早期原型，不属于现行桌面编辑主链路。

| 组成 | 当前规模 |
|---|---|
| 桌面与服务 | 1 个 Wails 应用；Go 后端管理项目、Vite、本机通信和文件事务 |
| 控制台 | 1 套 React + Tailwind UI，桌面与浏览器入口共用 |
| 执行器 | 3 类：CssExecutor、TailwindExecutor、JSXExecutor |
| 项目集成 | 1 套 Vite 插件、元素选择桥与源码 revision/指纹映射 |
| 示例项目 | 2 个真实 React/Vite 项目：普通 CSS 和 Tailwind |
| 编辑范围 | 视觉/布局属性修改，以及静态空容器插入注册按钮 |
| 验证 | Go race 测试、14 项 AST 测试、2 套真实 Chrome E2E、Wails 正式构建 |

本轮快照的手写源码与测试约 **4,762 行、32 个文件**：Go 生产代码 2,492 行、Go 测试 425 行、控制台 560 行、Vite/AST 集成 525 行、浏览器/AST 测试 447 行、两个 demo 源码 313 行。统计为物理行数（含空行/注释），不含依赖、锁文件、构建产物、Wails 自动绑定、配置/文档及历史 `web-prototype`；这是实现体量，不代表性能或产品成熟度。

已实际验证换色、插入按钮、引号文案、歧义/动态类拒绝、Accept/Reject/Undo、刷新持久化、外部修改冲突保护，以及 390px/1280px 布局。已完成无历史独立验收，对比度问题修复后专项复验通过。正式 macOS Wails 包也实际完成了选择、换色、HMR 和撤销。真实 Jev API 调用仍未验证；其他操作系统、任意大型项目、自动 ID 注入、多人协作和云同步不在当前保证范围内。

## 运行

需要 Go 1.25+、Node 20.19+ 或 22.12+、npm、Wails v2；macOS 正式构建需要 Xcode Command Line Tools。

```sh
git clone --branch v2 https://github.com/voidning/jev-evolution-studio.git
cd jev-evolution-studio
npm ci --prefix integration
npm ci --prefix frontend
npm ci --prefix editable-demo
npm ci --prefix tailwind-demo
npm run build --prefix integration
wails dev
```

打开克隆目录中 `tailwind-demo` 的绝对路径（或普通 CSS 的 `editable-demo`），点击“打开 / 启动预览”。在浏览器预览中点击元素，回控制台输入修改。Enter 生成 Diff，Shift+Enter 换行；中文输入法选字不会提交。选择模式捕获点击，Escape 退出。

```sh
wails build
open build/bin/ForgeDesktop.app
```

Wails 构建会先更新内嵌的 AST 执行器，再构建前端。手动 Go 构建须先构建这两项：

```sh
npm run build --prefix integration
npm run build --prefix frontend
mkdir -p work
go build -o work/jev-editor .
JEV_OFFLINE=1 ./work/jev-editor -headless -project "$PWD/tailwind-demo"
```

终端输出本机 Console / Preview 链接。首次认证后 URL 中的临时 token 被清除；它不是 TypeSafe Key。应用只停止自己启动的 Vite，不停止连接的已有服务。不自动安装项目依赖。

## 支持的操作

| 范围 | 示例 |
|---|---|
| 颜色（Tailwind） | 把这个按钮换成红色、文字换成白色、边框换成蓝色、背景换成 #ff3366、背景换成品牌色 |
| 字号 / 字重 / 对齐 | 文字大一点、这个标题再大一点、字重增强、文字居中、右对齐 |
| 间距 | 这里更紧凑、这里更宽松、padding增加、间距小一点、外边距大一点（Tailwind） |
| 布局 | 桌面端三列、卡片在桌面端改成三列，手机端保持一列、横向排列、纵向排列、全宽、宽度narrow |
| 外观 | 圆角小一点、圆角大一点、显示边框、让这个按钮更突出、阴影大一点（Tailwind）、透明度50%（Tailwind） |
| 可见性 | 隐藏这个元素、恢复元素、显示为flex（Tailwind） |
| 结构（Tailwind） | 在这个空框中间加一个按钮、在中间加一个“立即开始”按钮、在框中间加一个叫“保存”的按钮 |

Tailwind 支持 `桌面端` / `手机端`、`sm:` / `md:` / `lg:` 前缀；状态只在明确指定 `悬停时` / `hover:`、`聚焦时` / `focus:` 时编辑。基础、hover、responsive、dark 类互不混改。桌面映射 md，手机映射 max-md；示例使用 Tailwind 默认断点。自定义断点按项目 Tailwind 配置解释。普通 CSS 使用 768px 分界。

可用 `父容器`、`同级元素` 选择已标注的作用域；插入只接受 selected/all。封闭语法完整匹配，未知、混合、不支持的要求整体拒绝。字号、间距采用固定刻度；任意自由数值和业务逻辑不在本版范围。

## className 与结构边界

支持字符串字面量、无插值模板、从 clsx/classnames 导入的简单调用，以及 `condition && "静态类"`。只修改匹配属性和变体的叶节点，保留无关 token。跨多个条件分支的同一属性拒绝。

拒绝 `calculateClass(state)`、`styles[variant]`、CVA、spread 属性、内联 style、第三方组件内部样式和 `.map()` 的单实例编辑。不将共享实例修改伪装成单实例修改。

插入按钮仅限静态空的原生 div/section/main/article/aside/header/footer。已有内容会提示“作为内容居中还是悬浮居中”的歧义，并安全拒绝。包含响应式/状态布局的空容器也拒绝，避免破坏保留约束。AST 添加 grid、单列和居中类，并生成 JSX 节点与必要 import；recast 格式化修改的 AST 节点，其他源码保持原样。

组件注册表优先识别 `src/components/` 中命名导出的、包含原生 button 的 `Button`。唯一候选时添加或复用相对路径 import；多个候选拒绝。没有候选时生成原生 button。复杂组件、路径别名解析、variant 配置暂未支持；插入的组件不自动获得可编辑 ID，后续直接选择编辑需人工接入，或继续编辑其已标注父容器。

## 接入自己的项目

1. 保留整个 `integration/` 目录并安装其依赖。在项目 Vite 配置中导入 `integration/jev-vite.mjs` 的默认插件，将 `jevEditor()` 加到 plugins。
2. 给 `src` 内 JSX/TSX 的原生元素添加唯一的字面量 `data-jev-id="src-view-hero-title"`。ID 为 1–100 个字母、数字、下划线或连字符；不能在重复渲染中复用。
3. Tailwind 从 package.json 的依赖检测。普通 CSS 项目需创建 `src/jev-edits.css`，在应用入口最后导入，保证生产构建也包含修改。
4. 安装项目依赖，在控制台选择项目。连接现有预览时输入 `http://127.0.0.1:端口`，并确保服务已加载插件、绑定本机。

插件记录源文件、节点范围/指纹、className 来源、父节点与 revision；每次 Accept 重新验证源 revision、指纹和浏览器唯一匹配。自动 JSX ID 注入、CSS Modules、Next.js/Vue 和任意项目适配未包含。

## 事务、安全和 Jev

- 生成 Diff 只构造内存 Patch。Accept 保存历史、原子写入、运行项目 TypeScript（存在 tsconfig 时）及 Vite 构建，再等待发起浏览器的 HMR 回执。普通 CSS 同时检查覆盖是否实际生效。
- Diff 从同一组 before/after 字节产生。一次一笔草稿，Accept 或 Reject 后继续；Accept 后仍可单步 Undo。未接受草稿不跨重启保存；已写入历史可重启恢复。
- 使用 Go `os.Root`、同目录临时文件、原子替换和修改前内容比较；拒绝越界/符号链接，不覆盖无关 dirty 文件。多文件变更逐个原子替换，失败尝试回滚，不承诺跨文件/跨进程原子事务锁。
- 构建或 HMR 失败自动恢复；若外部修改阻止恢复，保留快照并报错。历史位于 macOS `~/Library/Application Support/JevEditor/history`，权限 0600。
- Tailwind HMR 验证加载 revision、实际类和样式可用性；不保证任意外部高优先级 CSS 的最终覆盖。因此首版面向显式接入的 Tailwind 项目。
- 本机代理仅绑定 127.0.0.1，随机会话 token、Strict/HTTP-only Cookie、来源和会话校验。API Key 只在 Go 后端；不进入 URL、bundle、项目文件或子进程环境。

明确颜色、引号文案、常见操作由本地解析器直接确定。颜色区分字面值与设计令牌；支持常见颜色词、hex、rgb/hsl/oklch、静态 Tailwind 配置与 CSS @theme 令牌。

在线 Jev 路径仅补充模糊颜色语义：warm/cool/vivid/soft/darker/lighter/brand/danger；单次批量分层 Choice/Noul，候选受目标、类和令牌裁剪。低置信度、非法值、超时、API失败均拒绝模糊请求，不生成代码。模糊文字/背景色需通过本地对比度检查；未知变量色、透明背景或低于 4.5:1 时要求明确颜色，不擅自改变另一通道。此检查使用固定 Tailwind 4 色板；真实服务调用需自己的 Key，未将模拟解码测试视为线上验证。

后端读取 `TYPESAFE_API_KEY`（兼容既有后端配置）。不要设置 `VITE_*` Key。控制台可强制离线，也可设置 `JEV_OFFLINE=1`。

## 检查

```sh
node --test integration/source.test.mjs
npm run build --prefix integration
go test -race ./...
npm run build --prefix frontend
npm run build --prefix editable-demo
npm run build --prefix tailwind-demo
go build -o work/jev-editor .
node frontend/tests/editor.e2e.mjs
node frontend/tests/tailwind.e2e.mjs
wails build
```

浏览器测试用已安装的 Google Chrome，复制临时 demo 后走真实 UI；不修改交付示例。覆盖颜色/插入/文案/歧义/动态类 A–E、Reject、源码冲突、完整 Undo、刷新持久性和 390/1280px 溢出。普通 CSS 回归覆盖原始四条命令、隐藏恢复、Escape、输入法、样式冲突回滚和哨兵 Key。证据位于 Git 忽略的 `work/e2e`、`work/tailwind-e2e`。

## 关键文件

- `app.go`：Wails 接口、项目进程、本机代理、会话与浏览器通信。
- `edit_flow.go`：草稿、Accept/Reject、校验、回滚、Undo。
- `executors.go` / `patch.go`：统一执行器、构建检查、受限原子写入、历史和 diff。
- `source_intent.go` / `intent.go` / `operations.go` / `color_contrast.go`：注册表、离线解析、Jev 类型边界、颜色校验。
- `integration/source-core.mjs`：AST 定位、组件与令牌注册、Tailwind 属性组、JSX Patch。
- `integration/jev-vite.mjs` / `selection-bridge.js`：源码映射、选择、HMR 回执。
- `frontend/src/`：Tailwind 控制台，桌面与浏览器共用。
- `tailwind-demo/` / `editable-demo/`：真实可运行的两种示例。
- `editor_test.go` / `integration/source.test.mjs` / `frontend/tests/`：后端、AST 和真实浏览器测试。
