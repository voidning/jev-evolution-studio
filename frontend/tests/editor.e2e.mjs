import { chromium } from "@playwright/test";
import assert from "node:assert/strict";
import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { spawn } from "node:child_process";
import { fileURLToPath } from "node:url";
const root = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
const temp = await fs.mkdtemp(
  path.join(await fs.realpath(os.tmpdir()), "jev-e2e-"),
);
const sentinel = "JEV_TEST_SECRET_SENTINEL_913845";
let backend,
  browser,
  logs = "",
  context;
const evidence = path.join(root, "work/e2e");
await fs.mkdir(evidence, { recursive: true });
function wait(ms) {
  return new Promise((r) => setTimeout(r, ms));
}
async function eventually(fn, message, timeout = 12000) {
  const until = Date.now() + timeout;
  let last;
  while (Date.now() < until) {
    try {
      const value = await fn();
      if (value) return value;
    } catch (e) {
      last = e;
    }
    await wait(80);
  }
  throw Error(message + (last ? " " + last : ""));
}
try {
  await fs.cp(path.join(root, "editable-demo/src"), path.join(temp, "src"), {
    recursive: true,
  });
  for (const name of ["package.json", "index.html"])
    await fs.copyFile(
      path.join(root, "editable-demo", name),
      path.join(temp, name),
    );
  await fs.symlink(
    path.join(root, "editable-demo/node_modules"),
    path.join(temp, "node_modules"),
  );
  await fs.writeFile(
    path.join(temp, "vite.config.mjs"),
    `import {defineConfig} from 'vite';import react from '@vitejs/plugin-react';import jev from ${JSON.stringify(path.join(root, "integration/jev-vite.mjs"))};export default defineConfig({plugins:[react(),jev()]});`,
  );
  const original =
    "/* Jev deterministic edits. Imported after application styles. */\n";
  await fs.writeFile(path.join(temp, "src/jev-edits.css"), original);
  backend = spawn(
    path.join(root, "work/jev-editor"),
    ["-headless", "-project", temp],
    {
      cwd: root,
      env: { ...process.env, JEV_OFFLINE: "1", TYPESAFE_API_KEY: sentinel },
      stdio: ["ignore", "pipe", "pipe"],
    },
  );
  backend.stdout.on("data", (d) => (logs += d));
  backend.stderr.on("data", (d) => (logs += d));
  const consoleURL = await eventually(
    () => logs.match(/Console: (http:\/\/[^\s]+)/)?.[1],
    "backend start",
  );
  const previewURL = logs.match(/Preview: (http:\/\/[^\s]+)/)[1];
  browser = await chromium.launch({ channel: "chrome", headless: true });
  context = await browser.newContext({
    viewport: { width: 1280, height: 900 },
  });
  const errors = [];
  context.on("page", (p) => p.on("pageerror", (e) => errors.push(e.message)));
  const consolePage = await context.newPage();
  await consolePage.goto(consoleURL);
  const preview = await context.newPage();
  await preview.goto(previewURL);
  const state = () =>
    consolePage.evaluate(() => fetch("/__jev/api/state").then((r) => r.json()));
  const css = () => fs.readFile(path.join(temp, "src/jev-edits.css"), "utf8");
  const select = async (id, position) => {
    await eventually(
      async () =>
        (await preview
          .getByRole("button", { name: "切换 Jev 元素选择模式" })
          .getAttribute("aria-pressed")) === "true",
      "browser selection mode",
    );
    await preview
      .locator(`[data-jev-id="${id}"]`)
      .click(position ? { position } : {});
    await eventually(
      async () => (await state()).selected?.id === id,
      "selection " + id,
    );
    await eventually(
      async () =>
        (await consolePage.locator(".selection code").innerText()) === id,
      "selection UI " + id,
    );
  };
  async function apply(text) {
    const before = await css();
    await consolePage.getByLabel("自然语言修改").fill(text);
    await consolePage
      .getByRole("button", { name: "生成 Diff", exact: true })
      .click();
    const s = await eventually(async () => {
      const s = await state();
      if (s.error) throw Error(s.error);
      const alert = await consolePage.getByRole("alert").allTextContents();
      if (alert.length) throw Error(alert.join(" "));
      return !s.busy && s.pending ? s : false;
    }, "apply " + text);
    assert.equal(await css(),before,'draft must not write');
    await consolePage.getByRole('button',{name:'Accept ✓',exact:true}).click();
    await eventually(async()=>{const s=await state();if(s.error)throw Error(s.error);return !s.busy&&!s.pending&&s.canUndo},'accept write');
    const after = await css();
    assert.notEqual(before, after);
    const added =
      s.diff
        .split("\n")
        .filter((l) => l.startsWith("+") && !l.startsWith("+++"))
        .map((l) => l.slice(1))
        .join("\n") + "\n";
    assert.equal(added, after, "diff matches written bytes");
    assert.equal((await state()).stages.at(-1).name,"Vite refreshed");
    assert.equal(s.mode, "offline");
    return { before, after, s };
  }
  async function undo(expected) {
    await consolePage
      .getByRole("button", { name: "↶ Undo", exact: true })
      .click();
    await eventually(
      async () => !(await state()).busy && !(await state()).canUndo,
      "undo",
    );
    assert.equal(await css(), expected);
  }
  async function accept() {
    if(!(await state()).pending)return;
    await consolePage
      .getByRole("button", { name: "Accept ✓", exact: true })
      .click();
    await eventually(async () => !(await state()).pending, "accept");
  }
  const title = preview.locator('[data-jev-id="src-main-title"]');
  const font = () => title.evaluate((e) => getComputedStyle(e).fontSize);
  const otherViewport = await context.newPage();
  await otherViewport.setViewportSize({ width: 390, height: 844 });
  await otherViewport.goto(previewURL);
  await select("src-main-title");
  const oldFont = await font();
  await apply("这个标题再大一点");
  assert.equal(parseFloat(await font()), parseFloat(oldFont) + 4);
  await undo(original);
  assert.equal(await font(), oldFont);
  await otherViewport.close();
  // Select the grid's own gap, rather than a child card.
  await select("src-main-features", { x: 550, y: 15 });
  const grid = preview.locator('[data-jev-id="src-main-features"]');
  const response = await apply("卡片在桌面端改成三列，手机端保持一列");
  assert.equal(
    await grid.evaluate(
      (e) => getComputedStyle(e).gridTemplateColumns.split(" ").length,
    ),
    3,
  );
  await preview.setViewportSize({ width: 390, height: 844 });
  assert.equal(
    await grid.evaluate(
      (e) => getComputedStyle(e).gridTemplateColumns.split(" ").length,
    ),
    1,
  );
  assert.equal(
    await preview.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
    true,
  );
  await preview.screenshot({
    path: path.join(evidence, "demo-390.png"),
    fullPage: true,
  });
  await consolePage.setViewportSize({ width: 390, height: 844 });
  assert.equal(
    await consolePage.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
    true,
  );
  await consolePage.screenshot({
    path: path.join(evidence, "console-390.png"),
    fullPage: true,
  });
  await preview.setViewportSize({ width: 1280, height: 900 });
  await consolePage.setViewportSize({ width: 1280, height: 900 });
  await accept();
  await undo(response.before);
  await select("src-main-card-capture", { x: 8, y: 8 });
  const card = preview.locator('[data-jev-id="src-main-card-capture"]');
  const padding = await card.evaluate((e) => getComputedStyle(e).paddingTop);
  const compact = await apply("这里更紧凑");
  assert.equal(
    parseFloat(await card.evaluate((e) => getComputedStyle(e).paddingTop)),
    parseFloat(padding) - 4,
  );
  await undo(compact.before);
  await select("src-main-cta-button");
  assert.equal(
    await preview.locator("#notice").innerText(),
    "",
    "selection must not run business click",
  );
  const button = preview.locator('[data-jev-id="src-main-cta-button"]');
  const oldBG = await button.evaluate(
    (e) => getComputedStyle(e).backgroundColor,
  );
  const strong = await apply("让这个按钮更突出");
  assert.notEqual(
    await button.evaluate((e) => getComputedStyle(e).backgroundColor),
    oldBG,
  );
  assert.equal(
    await preview.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
    true,
  );
  assert.equal(
    await consolePage.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
    true,
  );
  await preview.screenshot({
    path: path.join(evidence, "demo-1280.png"),
    fullPage: true,
  });
  await consolePage.screenshot({
    path: path.join(evidence, "console-1280.png"),
    fullPage: true,
  });
  await undo(strong.before);
  assert.equal(
    await button.evaluate((e) => getComputedStyle(e).backgroundColor),
    oldBG,
  );
  // Hidden targets remain selected and can be restored without clicking them.
  await apply("隐藏元素");
  assert.equal(await button.isVisible(), false);
  await accept();
  const restored = await apply("恢复元素");
  assert.equal(await button.isVisible(), true);
  await undo(restored.before);
  // Undo above restores the hidden version. Restore once more and accept.
  await apply("恢复元素");
  await accept();
  const safe = await css();
  await consolePage
    .getByLabel("自然语言修改")
    .fill("这个标题再大一点并删除业务代码");
  await consolePage
    .getByRole("button", { name: "生成 Diff", exact: true })
    .click();
  await eventually(
    async () => (await state()).error.includes("不支持"),
    "unsupported",
  );
  assert.equal(await css(), safe);
  await preview.keyboard.press("Escape");
  await eventually(async () => !(await state()).selecting, "escape");
  await button.click();
  assert.equal(
    await preview.locator("#notice").innerText(),
    "已准备好，一起开始。",
  );
  // A composing Enter must not submit. Shift+Enter inserts a newline.
  const input = consolePage.getByLabel("自然语言修改");
  await input.fill("圆角小一点");
  await input.dispatchEvent("compositionstart");
  await input.press("Enter");
  await input.dispatchEvent("compositionend");
  assert.equal(await css(), safe);
  await input.fill("第一行");
  await input.press("Shift+Enter");
  await input.press("A");
  assert.equal((await input.inputValue()).includes("\n"), true);
  // A more-specific important author rule must be rejected and rolled back.
  const authorFile = path.join(temp, "src/style.css");
  const authorOriginal = await fs.readFile(authorFile, "utf8");
  const authorDirty =
    authorOriginal +
    '\nhtml body [data-jev-id="src-main-title"] { font-size:55px !important; }\n';
  await fs.writeFile(authorFile, authorDirty);
  await eventually(
    async () => (await font()) === "55px",
    "author stylesheet HMR",
  );
  await preview.getByRole("button", { name: "切换 Jev 元素选择模式" }).click();
  await eventually(async () => (await state()).selecting, "selection enabled");
  await select("src-main-title");
  await input.fill("这个标题再大一点");
  await consolePage
    .getByRole("button", { name: "生成 Diff", exact: true })
    .click();
  await eventually(async()=>(await state()).pending,'conflict draft');
  await consolePage.getByRole('button',{name:'Accept ✓',exact:true}).click();
  await eventually(
    async () => (await state()).error.includes("原样式阻止"),
    "cascade conflict rollback",
  );
  assert.equal(await css(), safe);
  assert.equal(
    await fs.readFile(authorFile, "utf8"),
    authorDirty,
    "unrelated dirty styles unchanged",
  );
  await eventually(
    async () => (await font()) === "55px",
    "rollback visual restore",
  );
  // Request credentials and API keys must never appear in authored code or assets.
  assert.equal(logs.includes(sentinel), false);
  assert.equal((await css()).includes(sentinel), false);
  assert.equal(consolePage.url().includes(sentinel), false);
  assert.equal(preview.url().includes(sentinel), false);
  for (const name of await fs.readdir(
    path.join(root, "frontend/dist/assets"),
  )) {
    const text = await fs.readFile(
      path.join(root, "frontend/dist/assets", name),
      "utf8",
    );
    assert.equal(text.includes(sentinel), false);
  }
  assert.deepEqual(errors, []);
  const report = {
    result: "PASS",
    checks: [
      "offline examples",
      "real browser selection",
      "real CSS HMR",
      "byte-identical diff",
      "Accept then Undo",
      "390px and 1280px overflow",
      "hidden target restore",
      "Escape restores business click",
      "unsupported input no writes",
      "IME Enter and Shift+Enter",
      "secret sentinel not exposed",
      "cascade conflict rollback",
      "unrelated dirty files preserved",
      "selection bound to initiating browser tab",
    ],
    screenshots: evidence,
  };
  await fs.writeFile(
    path.join(evidence, "report.json"),
    JSON.stringify(report, null, 2),
  );
  console.log(JSON.stringify(report, null, 2));
} catch (error) {
  if (context) {
    for (const [index, page] of context.pages().entries()) {
      await page
        .screenshot({
          path: path.join(evidence, "failure-" + index + ".png"),
          fullPage: true,
        })
        .catch(() => {});
    }
    const first = context.pages()[0];
    if (first) console.error(await first.locator("body").innerText());
  }
  throw error;
} finally {
  if (browser) await browser.close();
  if (backend) backend.kill("SIGTERM");
  await wait(500);
  await fs.rm(temp, { recursive: true, force: true });
}
