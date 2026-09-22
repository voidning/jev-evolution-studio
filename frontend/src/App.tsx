import { useEffect, useRef, useState } from "react";
import { call, isNative, type State } from "./api";
import ProjectPanel from "./ProjectPanel";
import DiffPanel from "./DiffPanel";

export default function App() {
  const [state, setState] = useState<State | null>(null),
    [prompt, setPrompt] = useState(""),
    [error, setError] = useState(""),
    [working, setWorking] = useState(false);
  const composing = useRef(false);
  useEffect(() => {
    let active = true;
    let busy = false;
    const poll = async () => {
      if (busy) return;
      busy = true;
      try {
        const s = await call("state");
        if (active) setState(s);
      } catch (e) {
        if (active) setError(String(e));
      } finally {
        busy = false;
      }
    };
    void poll();
    const timer = setInterval(() => void poll(), 300);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, []);
  async function action(name: string, payload: Record<string, unknown> = {}) {
    setWorking(true);
    setError("");
    try {
      const s = await call(name, payload);
      setState(s);
    } catch (e) {
      setError(String(e));
    } finally {
      setWorking(false);
    }
  }
  async function preview() {
    setWorking(true);
    setError("");
    try {
      const s = await call("start");
      setState(s);
      if (isNative()) await call<string>("stage");
      else window.open(s.previewURL, "jev-preview");
    } catch (e) {
      setError(String(e));
    } finally {
      setWorking(false);
    }
  }
  const busy = working || !!state?.busy;
  const apply = () => {
    if (!busy && state?.selected && prompt.trim())
      void action("apply", {
        prompt,
        session: state.session,
        targetId: state.selected.id,
      });
  };
  const selected = state?.selected;
  return (
    <main className="editor">
      <header className="app-header">
        <div>
          <h1 className="wordmark">
            <span className="logo">j.</span>Jev{" "}
            <span className="product">网页编辑器</span>
          </h1>
          <p>点击网页上的元素，说出修改，代码立即同步。</p>
        </div>
        <span className="mode">
          {state?.hasAPIKey ? "JEV 意图判断" : "离线规则模式"}
        </span>
      </header>
      <ProjectPanel
        state={state}
        busy={busy}
        action={action}
        preview={preview}
      />
      <div className="workspace">
        <div className="edit-column">
          <section className="panel">
            <div className="section-title">
              <h2>
                01 <span>选择元素</span>
              </h2>
              <button
                className={state?.selecting ? "active small" : "small"}
                disabled={!state?.connected}
                onClick={() =>
                  void call("select", { on: !state?.selecting })
                    .then(setState)
                    .catch((e) => setError(String(e)))
                }
              >
                {state?.selecting ? "选择模式开启" : "开启选择模式"}
              </button>
            </div>
            {selected ? (
              <div className="selection">
                <strong>〈{selected.tag}〉</strong>
                <code>{selected.id}</code>
                <span>
                  {selected.source}:{selected.line} · {selected.classKind || "CSS"}
                </span>
                <div className="style-summary">
                  {["font-size", "display", "gap", "padding-top"].map((p) => (
                    <span key={p}>
                      {p} <b>{selected.styles[p] || "—"}</b>
                    </span>
                  ))}
                </div>
              </div>
            ) : (
              <div className="empty">
                在预览中点击标题、卡片组或按钮。
                <small>选择时会阻止页面原点击行为，按 Esc 退出。</small>
              </div>
            )}
          </section>
          <section className="panel">
            <div className="section-title">
              <h2>
                02 <span>描述修改</span>
              </h2>
              <span className="hint">Enter 应用</span>
            </div>
            <label className="sr-only" htmlFor="edit-prompt">
              自然语言修改
            </label>
            <textarea
              id="edit-prompt"
              value={prompt}
              placeholder="例如：这个标题再大一点"
              onChange={(e) => setPrompt(e.target.value)}
              onCompositionStart={() => (composing.current = true)}
              onCompositionEnd={() => (composing.current = false)}
              onKeyDown={(e) => {
                if (
                  e.key === "Enter" &&
                  !e.shiftKey &&
                  !e.nativeEvent.isComposing &&
                  !composing.current &&
                  e.keyCode !== 229
                ) {
                  e.preventDefault();
                  apply();
                }
              }}
            />
            <div className="command-examples">
              {[
                "这个标题再大一点",
                "卡片在桌面端改成三列，手机端保持一列",
                "这里更紧凑",
                "让这个按钮更突出",
              ].map((text) => (
                <button key={text} onClick={() => setPrompt(text)}>
                  {text}
                </button>
              ))}
            </div>
            <div className="apply-row">
              <span>Shift + Enter 换行</span>
              <button
                className="primary"
                disabled={
                  busy ||
                  !selected ||
                  !prompt.trim() ||
                  !!state?.pending ||
                  !state?.browserConnected
                }
                onClick={apply}
              >
                {busy ? "处理中…" : "生成 Diff"}
              </button>
            </div>
            {(error || state?.error) && (
              <p className="error" role="alert">
                {error || state?.error}
              </p>
            )}
            {state?.pending && (
              <p className="notice">
                Diff 已就绪，尚未写入文件。Accept 后应用，Reject 放弃。
              </p>
            )}
            <ol className="stages" aria-live="polite">
              {state?.stages?.map((s) => (
                <li key={s.name}>
                  <span>
                    {s === state.stages?.at(-1) && state.error
                      ? "!"
                      : state.busy && s === state.stages?.at(-1)
                        ? "…"
                        : "✓"}{" "}
                    {s.name}
                  </span>
                  <small>{s.ms} ms</small>
                </li>
              ))}
            </ol>
          </section>
          <details className="panel intent">
            <summary>
              最近一次 EditIntent <span>{state?.mode}</span>
            </summary>
            <pre>
              {state?.intents?.length
                ? JSON.stringify(state.intents, null, 2)
                : "等待修改指令"}
            </pre>
          </details>
        </div>
        <DiffPanel state={state} busy={busy} action={action} />
      </div>
      <footer>
        本地修改 · 可提交 Git · 单步撤销<span>React / Vite / Tailwind + CSS</span>
      </footer>
    </main>
  );
}
