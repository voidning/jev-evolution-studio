import type { State } from "./api";
export default function DiffPanel({
  state,
  busy,
  action,
}: {
  state: State | null;
  busy: boolean;
  action: (name: string) => Promise<void>;
}) {
  return (
    <section className="panel diff-panel">
      <div className="section-title">
        <h2>
          03 <span>检查代码</span>
        </h2>
        <span className="hint">真实文件</span>
      </div>
      <div className="file-label">{state?.files?.join(" · ") || "等待源码 Patch"}</div>
      <pre className="diff" aria-label="代码 diff">
        {state?.diff ? (
          state.diff.split("\n").map((line, i) => (
            <div
              className={
                line.startsWith("+")
                  ? "addition"
                  : line.startsWith("-")
                    ? "deletion"
                    : "context"
              }
              key={i}
            >
              {line || " "}
            </div>
          ))
        ) : (
          <span className="diff-empty">
            修改后的代码 diff 会显示在这里。
            <br />
            先检查 Diff；Accept 后写入真实源码。
          </span>
        )}
      </pre>
      <div className="diff-actions">
        <button
          disabled={busy || !state?.canUndo || !!state?.pending}
          onClick={() => void action("undo")}
        >
          ↶ Undo
        </button>
        <button disabled={busy || !state?.pending} onClick={()=>void action("reject")}>Reject</button>
        <button
          className="primary"
          disabled={busy || !state?.pending}
          onClick={() => void action("accept")}
        >
          Accept ✓
        </button>
      </div>
    </section>
  );
}
