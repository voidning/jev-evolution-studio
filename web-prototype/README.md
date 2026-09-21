# Forge — Jev UI Compiler

一个用于现场演示的实时 UI 编译器：自然语言输入会被分解为并行的类型化设计决策，编译成三个可交互网页候选，并支持继续用自然语言局部修改。

## 运行

需要 Node.js 22.13 或更新版本。

```bash
npm install
npm run dev
```

打开终端输出的本地地址。

## 接入 Jev

复制环境变量示例并填写 TypeSafe API Key：

```bash
cp .env.example .env.local
```

```env
TYPESAFE_API_KEY=your_key_here
```

重启开发服务后，顶部状态会从 `Local decision engine` 变为 `Jev 1.13`。没有密钥时，应用自动使用本地语义决策引擎，完整交互仍然可演示。

## 主要文件

- `app/page.tsx`：编辑器、页面渲染器、候选切换与 WebMCP 工具
- `app/api/design/route.ts`：Jev 并行问题、概率结果到 PageSpec 的编译、本地降级
- `app/globals.css`：工作台与生成页面的完整视觉系统

## 验证

```bash
npm run build
```
