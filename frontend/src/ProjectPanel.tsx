import { useState } from "react";
import { isNative, type State } from "./api";
export default function ProjectPanel({
  state,
  busy,
  action,
  preview,
}: {
  state: State | null;
  busy: boolean;
  action: (name: string, payload?: Record<string, unknown>) => Promise<void>;
  preview: () => Promise<void>;
}) {
  const [path, setPath] = useState(""),
    [url, setURL] = useState("");
  return (
    <section className="panel project">
      <div className="section-title">
        <h2>项目与预览</h2>
        <span className={`connection ${state?.browserConnected ? "live" : ""}`}>
          {state?.browserConnected
            ? "● 浏览器已连接"
            : state?.connected
              ? "○ 等待浏览器"
              : "○ 尚未连接"}
        </span>
      </div>
      <p className="project-path">
        {state?.project || "选择一个已接入 Jev 的 React + Vite 项目"}
      </p>
      <div className="input-row">
        <input
          aria-label="项目路径"
          placeholder="/path/to/react-vite-project"
          value={path}
          onChange={(e) => setPath(e.target.value)}
        />
        <button
          disabled={busy || !!state?.pending}
          onClick={() => void action("open", { path })}
        >
          打开项目
        </button>
        {isNative() && (
          <button
            disabled={busy || !!state?.pending}
            onClick={() => void action("choose")}
          >
            浏览…
          </button>
        )}
      </div>
      <div className="project-actions">
        <button
          className="primary"
          disabled={busy || !state?.project}
          onClick={() => void preview()}
        >
          ↗ 打开 / 启动预览
        </button>
        <label className="check">
          <input
            type="checkbox"
            checked={!state?.hasAPIKey}
            disabled={busy}
            onChange={(e) => void action("offline", { on: e.target.checked })}
          />
          离线规则
        </label>
      </div>
      <details>
        <summary>连接已启动的 Vite</summary>
        <div className="input-row">
          <input
            aria-label="Vite 地址"
            placeholder="http://127.0.0.1:5173"
            value={url}
            onChange={(e) => setURL(e.target.value)}
          />
          <button
            disabled={busy || !state?.project || !!state?.pending}
            onClick={() => void action("connect", { url })}
          >
            连接
          </button>
        </div>
      </details>
    </section>
  );
}
