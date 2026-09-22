import "/@vite/client";

const clientID = sessionStorage.getItem("jev-client") || crypto.randomUUID();
sessionStorage.setItem("jev-client", clientID);
let sourceMap = new Map();
async function refreshSources(){const r=await fetch('/__jev/meta');if(!r.ok)throw Error('Source mapping unavailable');const m=await r.json();sourceMap=new Map(m.targets.map(t=>[t.id,t]));}

const styleProperties = [
  "display",
  "font-size",
  "font-weight",
  "padding-top",
  "padding-right",
  "padding-bottom",
  "padding-left",
  "gap",
  "text-align",
  "grid-template-columns",
  "flex-direction",
  "width",
  "max-width",
  "border-radius",
  "background-color",
  "color",
];
let session = "",
  selecting = false,
  selected = null,
  commandDone = "",
  busy = false;
const box = document.createElement("div");
const label = document.createElement("div");
const status = document.createElement("button");
box.setAttribute("aria-hidden", "true");
box.style.cssText =
  "position:fixed;pointer-events:none;z-index:2147483646;border:2px solid #007f91;outline:2px solid white;display:none;box-sizing:border-box";
label.style.cssText =
  "position:fixed;pointer-events:none;z-index:2147483647;background:#123844;color:white;padding:5px 9px;font:12px/1.5 system-ui;max-width:calc(100vw - 20px);overflow-wrap:anywhere;display:none;border-radius:4px";
status.textContent = "Jev · 连接中";
status.style.cssText =
  "position:fixed;bottom:14px;right:14px;z-index:2147483647;border:1px solid #80a7b0;border-radius:6px;background:#123844;color:white;padding:8px 12px;font:12px system-ui;cursor:pointer";
status.setAttribute("aria-label", "切换 Jev 元素选择模式");
document.body.append(box, label, status);
function patchStyle() {
  return [...document.querySelectorAll("style[data-vite-dev-id]")].find((e) =>
    e.dataset.viteDevId.replaceAll("\\", "/").endsWith("/src/jev-edits.css"),
  );
}
function ensureLast() {
  const s = patchStyle();
  if (s && s !== document.head.lastElementChild) document.head.append(s);
}
const observer = new MutationObserver(ensureLast);
observer.observe(document.head, { childList: true });
ensureLast();
function hide() {
  box.style.display = "none";
  label.style.display = "none";
}
function outline(el) {
  if (!el?.isConnected || getComputedStyle(el).display === "none") {
    hide();
    return;
  }
  const r = el.getBoundingClientRect();
  box.style.display = "block";
  Object.assign(box.style, {
    left: r.left + "px",
    top: r.top + "px",
    width: r.width + "px",
    height: r.height + "px",
  });
  label.textContent = el.tagName.toLowerCase() + " · " + el.dataset.jevId;
  label.style.display = "block";
  label.style.left = Math.max(8, Math.min(r.left, innerWidth - 240)) + "px";
  label.style.top = Math.max(4, r.top - 32) + "px";
}
function snapshot(el) {
  const s = getComputedStyle(el),
    id = el.dataset.jevId;
  const parent = el.parentElement?.closest("[data-jev-id]");
  const important = [];
  for (const p of el.style) {
    if (el.style.getPropertyPriority(p) === "important") {
      important.push(p);
      if (p === "padding")
        important.push(
          "padding-top",
          "padding-right",
          "padding-bottom",
          "padding-left",
        );
      if (p === "background") important.push("background-color");
      if (p === "font") important.push("font-size", "font-weight");
    }
  }
  const source=sourceMap.get(id)||{};
  return {
    ...source,
    revision:globalThis.__JEV_REVISIONS__?.[source.source]||source.revision,
    id,
    tag: el.tagName.toLowerCase(),
    source:source.source||"",
    line:source.line||0,
    parent: parent?.dataset.jevId || "",
    styles: Object.fromEntries(
      styleProperties.map((p) => [p, s.getPropertyValue(p)]),
    ),
    important,
    count: document.querySelectorAll(`[data-jev-id="${CSS.escape(id)}"]`)
      .length,
    width: innerWidth,
  };
}
async function post(data) {
  const res = await fetch("/__jev/api/report", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      session,
      clientId: clientID,
      project: projectRoot,
      ...data,
    }),
  });
  if (!res.ok) throw Error("Jev session disconnected");
}
function choose(e) {
  if (!selecting || e.target === status) return;
  const el = e.target.closest?.("[data-jev-id]");
  e.preventDefault();
  e.stopImmediatePropagation();
  if (!el) return;
  selected = el;
  outline(el);
  void refreshSources().then(()=>post({selected:snapshot(el)})).catch(()=>{});
}
document.addEventListener("click", choose, true);
document.addEventListener(
  "pointerdown",
  (e) => {
    if (selecting && e.target !== status) {
      e.preventDefault();
      e.stopImmediatePropagation();
    }
  },
  true,
);
document.addEventListener(
  "mouseover",
  (e) => {
    if (selecting && e.target !== status) {
      const el = e.target.closest?.("[data-jev-id]");
      if (el) outline(el);
      else hide();
    }
  },
  true,
);
document.addEventListener("keydown", (e) => {
  if (e.key === "Escape") {
    selecting = false;
    hide();
    void post({ commandId: "escape" }).catch(() => {});
  }
});
status.onclick = () =>
  fetch("/__jev/api/select", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ on: !selecting }),
  }).catch(() => {});
window.addEventListener(
  "scroll",
  () => {
    if (selecting && selected) outline(selected);
  },
  true,
);
window.addEventListener("resize", () => {
  if (selecting && selected) outline(selected);
});
async function hash(text) {
  const b = await crypto.subtle.digest(
    "SHA-256",
    new TextEncoder().encode(text),
  );
  return [...new Uint8Array(b)]
    .map((v) => v.toString(16).padStart(2, "0"))
    .join("");
}
function verifyApplied(expectations = []) {
  const merged = new Map();
  for (const check of expectations) {
    if (
      (check.breakpoint === "mobile" && innerWidth >= 768) ||
      (check.breakpoint === "desktop" && innerWidth < 768)
    )
      continue;
    merged.set(check.id, { ...merged.get(check.id), ...check.properties });
  }
  for (const [id, props] of merged) {
    const matches = document.querySelectorAll(
      `[data-jev-id="${CSS.escape(id)}"]`,
    );
    if (matches.length !== 1) return "目标在刷新期间消失或重复";
    const el = matches[0],
      before = getComputedStyle(el),
      keys = Object.keys(props);
    const actual = Object.fromEntries(
      keys.map((p) => [p, before.getPropertyValue(p)]),
    );
    const original = el.getAttribute("style");
    try {
      // Synchronous cascade probe, restored before paint. Persisted edits always come from the file.
      for (const [p, v] of Object.entries(props))
        el.style.setProperty(p, v, "important");
      const expected = getComputedStyle(el);
      for (const p of keys)
        if (expected.getPropertyValue(p) !== actual[p])
          return "不支持：原样式阻止 " + p + " 生效，修改已取消";
    } finally {
      if (original === null) el.removeAttribute("style");
      else el.setAttribute("style", original);
    }
  }
  return "";
}
async function poll() {
  if (busy) return;
  busy = true;
  try {
    const res = await fetch("/__jev/api/bridge");
    if (!res.ok) throw Error("Disconnected");
    const state = await res.json();
    if (state.project !== projectRoot) throw Error("Wrong project");
    if (session && session !== state.session) {
      location.reload();
      return;
    }
    if (session !== state.session) {
      session = state.session;
      selected = null;
      commandDone = "";
    }
    selecting = state.selecting;
    status.textContent = selecting
      ? "Jev · 选择模式 · Esc 退出"
      : "Jev · 点击开始选择";
    status.setAttribute("aria-pressed", String(selecting));
    if (!selecting) hide();
    const cmd = state.command;
    if (cmd && cmd.clientId === clientID && cmd.id !== commandDone) {
      if (cmd.kind === "inspect") {
        await refreshSources();
        await post({
          commandId: cmd.id,
          targets: [...document.querySelectorAll("[data-jev-id]")].map(
            snapshot,
          ),
        });
        commandDone = cmd.id;
      }
      if(cmd.kind==='source-refresh') {
        const loaded=Object.entries(cmd.revisions).every(([file,revision])=>globalThis.__JEV_REVISIONS__?.[file]===revision);
        if(loaded){
          await new Promise(requestAnimationFrame);await new Promise(requestAnimationFrame);await refreshSources();
          const css=[...document.querySelectorAll('style')].map(s=>s.textContent).join('\n');
          let ready=true;
          for(const check of cmd.checks||[]){const nodes=document.querySelectorAll(`[data-jev-id="${CSS.escape(check.id)}"]`);if(nodes.length!==1){ready=false;break;}const el=nodes[0];
            if(check.label && ![...el.querySelectorAll('button')].some(b=>b.textContent===check.label)){ready=false;break;}
            if(check.className&&['literal','template','missing'].includes(check.classKind)){for(const token of check.className.split(/\s+/).filter(Boolean)){if(!el.classList.contains(token)||!css.includes(CSS.escape(token))){ready=false;break;}}}
          }
          if(ready){if(!selected?.isConnected && cmd.targets?.[0])selected=document.querySelector(`[data-jev-id="${CSS.escape(cmd.targets[0])}"]`);await post({commandId:cmd.id,selected:selected?.isConnected?snapshot(selected):undefined});commandDone=cmd.id;}
        }
      }
      if (cmd.kind === "refresh") {
        ensureLast();
        const s = patchStyle();
        const loaded = s ? await hash(s.textContent) : "";
        if (loaded === cmd.hash) {
          await new Promise(requestAnimationFrame);
          await new Promise(requestAnimationFrame);
          const error = verifyApplied(cmd.expectations);
          await post({
            commandId: cmd.id,
            hash: loaded,
            error,
            selected: selected?.isConnected ? snapshot(selected) : undefined,
          });
          commandDone = cmd.id;
          if (selecting && selected) outline(selected);
        }
      }
    }
    await post({});
  } catch {
    status.textContent = "Jev · 未连接编辑器";
    selecting = false;
    hide();
  } finally {
    busy = false;
  }
}
setInterval(poll, 180);
void poll();
if (import.meta.hot) {
  import.meta.hot.on("vite:afterUpdate", () => {
    ensureLast();
    void poll();
  });
  import.meta.hot.dispose(() => observer.disconnect());
}
