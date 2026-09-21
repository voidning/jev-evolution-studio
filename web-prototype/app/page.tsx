"use client";

import {
  ArrowRight, Check, ChevronDown, CircleDot, Code2, Eye, Gauge,
  Layers3, Loader2, Monitor, Moon, Play, Plus, RefreshCw, Send,
  Smartphone, Sparkles, WandSparkles, Zap,
} from "lucide-react";
import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";

type ThemeName = "violet" | "ember" | "mint";
type HeroName = "split" | "centered" | "terminal";
type FeatureName = "bento" | "columns" | "timeline";
type DeviceName = "desktop" | "mobile";

export type Decision = {
  label: string;
  value: string;
  confidence: number;
  group: "Intent" | "Structure" | "Visual";
};

export type PageSpec = {
  id: string;
  name: string;
  descriptor: string;
  theme: ThemeName;
  hero: HeroName;
  features: FeatureName;
  density: "airy" | "balanced" | "compact";
  showLogos: boolean;
  showPricing: boolean;
  showStats: boolean;
  title: string;
  description: string;
  decisions: Decision[];
};

type DesignResponse = {
  mode: "jev" | "local";
  latencyMs: number;
  prompt: string;
  specs: PageSpec[];
};

const INITIAL_PROMPT = "为一款实时 AI 数据分析产品做主页。面向开发者，深色、克制、有速度感，重点突出实时分析。";

const INITIAL_SPECS: PageSpec[] = [
  {
    id: "signal", name: "Signal", descriptor: "Technical minimal", theme: "violet",
    hero: "split", features: "bento", density: "balanced", showLogos: true,
    showPricing: false, showStats: true, title: "See the signal.\nBefore everyone else.",
    description: "Turn live product data into decisions your team can act on — without waiting for another dashboard refresh.",
    decisions: [
      { label: "Archetype", value: "Developer product", confidence: 0.94, group: "Intent" },
      { label: "Primary goal", value: "Explore product", confidence: 0.88, group: "Intent" },
      { label: "Hero", value: "Split dashboard", confidence: 0.91, group: "Structure" },
      { label: "Features", value: "Asymmetric bento", confidence: 0.84, group: "Structure" },
      { label: "Visual tone", value: "Technical minimal", confidence: 0.93, group: "Visual" },
      { label: "Density", value: "Balanced", confidence: 0.78, group: "Visual" },
    ],
  },
  {
    id: "pulse", name: "Pulse", descriptor: "Bold data-first", theme: "mint",
    hero: "terminal", features: "timeline", density: "compact", showLogos: false,
    showPricing: false, showStats: true, title: "Your product,\nat live speed.",
    description: "One continuous stream from event to insight. Query, detect, and ship the next decision in milliseconds.",
    decisions: [
      { label: "Archetype", value: "Developer product", confidence: 0.89, group: "Intent" },
      { label: "Primary goal", value: "Understand speed", confidence: 0.86, group: "Intent" },
      { label: "Hero", value: "Terminal first", confidence: 0.82, group: "Structure" },
      { label: "Features", value: "Live timeline", confidence: 0.79, group: "Structure" },
      { label: "Visual tone", value: "Bold technical", confidence: 0.85, group: "Visual" },
      { label: "Density", value: "Compact", confidence: 0.73, group: "Visual" },
    ],
  },
  {
    id: "orbit", name: "Orbit", descriptor: "Editorial systems", theme: "ember",
    hero: "centered", features: "columns", density: "airy", showLogos: true,
    showPricing: true, showStats: false, title: "Intelligence that\nkeeps up.",
    description: "A calmer, clearer way to understand every moving part of your product — as it happens.",
    decisions: [
      { label: "Archetype", value: "Data product", confidence: 0.83, group: "Intent" },
      { label: "Primary goal", value: "Build trust", confidence: 0.81, group: "Intent" },
      { label: "Hero", value: "Centered product", confidence: 0.77, group: "Structure" },
      { label: "Features", value: "Editorial columns", confidence: 0.76, group: "Structure" },
      { label: "Visual tone", value: "Editorial systems", confidence: 0.8, group: "Visual" },
      { label: "Density", value: "Airy", confidence: 0.75, group: "Visual" },
    ],
  },
];

const QUICK_PROMPTS = ["做得更大胆，但保持专业", "换成暖色，减少企业感", "保留结构，增加定价"];

const THEME = {
  violet: { accent: "#9674ff", soft: "rgba(150,116,255,.16)", glow: "rgba(117,72,255,.34)", panel: "#111118" },
  mint: { accent: "#75f2c2", soft: "rgba(117,242,194,.13)", glow: "rgba(62,218,164,.28)", panel: "#0d1513" },
  ember: { accent: "#ff8e64", soft: "rgba(255,142,100,.14)", glow: "rgba(255,105,67,.27)", panel: "#17110f" },
};

function MiniChart({ accent }: { accent: string }) {
  const id = `fill-${accent.replace("#", "")}`;
  return (
    <svg viewBox="0 0 420 160" className="mini-chart" role="img" aria-label="Rising live activity chart">
      <defs><linearGradient id={id} x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor={accent} stopOpacity=".34" /><stop offset="1" stopColor={accent} stopOpacity="0" /></linearGradient></defs>
      {[28, 66, 104, 142].map((y) => <line key={y} x1="0" x2="420" y1={y} y2={y} stroke="rgba(255,255,255,.07)" />)}
      <path d="M0 128 C38 132,42 104,82 111 S139 84,174 91 S220 49,257 62 S303 36,336 45 S374 21,420 27 L420 160 L0 160 Z" fill={`url(#${id})`} />
      <path d="M0 128 C38 132,42 104,82 111 S139 84,174 91 S220 49,257 62 S303 36,336 45 S374 21,420 27" fill="none" stroke={accent} strokeWidth="2.5" strokeLinecap="round" />
      <circle cx="420" cy="27" r="5" fill={accent} />
    </svg>
  );
}

function ProductVisual({ spec }: { spec: PageSpec }) {
  const theme = THEME[spec.theme];
  if (spec.hero === "terminal") {
    return (
      <div className="preview-terminal">
        <div className="preview-windowbar"><span /><span /><span /><small>live-query.ts</small></div>
        <div className="terminal-lines">
          <p><i>01</i><b>stream</b>(<em>&quot;product.events&quot;</em>)</p>
          <p><i>02</i> .detect(<em>&quot;activation_spike&quot;</em>)</p>
          <p><i>03</i> .window(<em>&quot;200ms&quot;</em>)</p>
          <p><i>04</i> .onSignal(<b>shipDecision</b>)</p>
          <p className="terminal-result"><i>›</i> signal detected <span>+24.8%</span></p>
        </div>
      </div>
    );
  }
  return (
    <div className="product-card" style={{ background: theme.panel }}>
      <div className="product-sidebar"><div className="product-mark" style={{ background: theme.accent }} />{[1, 2, 3, 4].map((item) => <span key={item} />)}</div>
      <div className="product-main">
        <div className="product-head"><div><small>Live workspace</small><strong>Activation</strong></div><span className="live-pill"><i style={{ background: theme.accent }} /> LIVE</span></div>
        <div className="metric-row"><div><small>Active now</small><strong>8,428</strong><em>+18.4%</em></div><div><small>Signals</small><strong>312</strong><em>+6.2%</em></div><div><small>Latency</small><strong>84ms</strong><em>−12ms</em></div></div>
        <div className="chart-wrap"><MiniChart accent={theme.accent} /></div>
      </div>
    </div>
  );
}

function PreviewPage({ spec }: { spec: PageSpec }) {
  const theme = THEME[spec.theme];
  const centered = spec.hero === "centered";
  return (
    <div className={`generated-page density-${spec.density}`} style={{ "--accent": theme.accent, "--accent-soft": theme.soft, "--glow": theme.glow } as React.CSSProperties}>
      <nav className="generated-nav"><div className="generated-logo"><span><Zap size={12} fill="currentColor" /></span>Velocity</div><div className="generated-links"><a>Product</a><a>Developers</a><a>Company</a></div><button>Start building <ArrowRight size={12} /></button></nav>
      <section className={`generated-hero ${centered ? "hero-centered" : "hero-split"}`}>
        <div className="hero-copy"><div className="eyebrow"><CircleDot size={12} /> Built for live products</div><h1>{spec.title.split("\n").map((line) => <span key={line}>{line}</span>)}</h1><p>{spec.description}</p><div className="hero-actions"><button>Start building <ArrowRight size={13} /></button><a><Play size={12} fill="currentColor" /> Watch 90 sec</a></div>{!centered && <div className="hero-proof"><span>TRUSTED BY TEAMS AT</span><b>northstar</b><b>linear</b><b>orbit</b></div>}</div>
        <div className="hero-visual"><ProductVisual spec={spec} /></div>
      </section>
      {spec.showLogos && centered && <div className="logo-strip"><span>TRUSTED BY PRODUCT TEAMS AT</span><b>northstar</b><b>Linear</b><b>ORBIT</b><b>cascade</b></div>}
      {spec.showStats && <section className="stats-strip"><div><strong>84<span>ms</span></strong><small>median signal latency</small></div><div><strong>12.4<span>B</span></strong><small>events processed</small></div><div><strong>99.99<span>%</span></strong><small>pipeline uptime</small></div></section>}
      <section className={`feature-section features-${spec.features}`}><header><small>ONE CONTINUOUS SYSTEM</small><h2>From raw event to<br />clear decision.</h2></header><div className="feature-grid"><article className="feature-primary"><span><Gauge size={15} /></span><small>REAL-TIME ENGINE</small><h3>Every signal,<br />while it matters.</h3><p>Stream product events into one live model of what your users are doing right now.</p><MiniChart accent={theme.accent} /></article><article><span><Layers3 size={15} /></span><small>SEMANTIC LAYER</small><h3>Ask in plain language.</h3><p>No brittle queries. Describe the change you need to understand.</p></article><article><span><Code2 size={15} /></span><small>DEVELOPER FIRST</small><h3>Ship the decision.</h3><p>Move from detected pattern to product action in a few typed lines.</p></article></div></section>
      {spec.showPricing && <section className="pricing-section"><div><small>SIMPLE PRICING</small><h2>Start live.<br />Scale when it works.</h2></div><article><span>PRO</span><strong>$49<small>/month</small></strong><p>Unlimited live queries<br />30-day event history<br />Priority pipelines</p><button>Start free <ArrowRight size={13} /></button></article></section>}
    </div>
  );
}

function DecisionRail({ decisions, generating }: { decisions: Decision[]; generating: boolean }) {
  const groups = ["Intent", "Structure", "Visual"] as const;
  return (
    <aside className="decision-rail">
      <div className="rail-head"><div><span className={generating ? "thinking" : ""}><Sparkles size={14} /></span><div><strong>Decision graph</strong><small>{generating ? "Evaluating in parallel…" : `${decisions.length} nodes resolved`}</small></div></div><button aria-label="Decision graph options"><ChevronDown size={14} /></button></div>
      <div className="decision-flow"><div className="source-node"><WandSparkles size={13} /><span>Natural language brief</span><Check size={12} /></div>{groups.map((group, groupIndex) => <div className="decision-group" key={group}><div className="group-label"><span>{String(groupIndex + 1).padStart(2, "0")}</span>{group}</div>{decisions.filter((decision) => decision.group === group).map((decision, index) => <div className={`decision-node ${generating ? "resolving" : ""}`} style={{ animationDelay: `${(groupIndex * 2 + index) * 90}ms` }} key={decision.label}><div><small>{decision.label}</small><strong>{decision.value}</strong></div><span>{Math.round(decision.confidence * 100)}</span></div>)}</div>)}<div className="compiled-node"><Check size={12} /><span>PageSpec compiled</span><small>0 errors</small></div></div>
      <div className="rail-foot"><span><i /> Typed decisions</span><span><i /> Confidence</span></div>
    </aside>
  );
}

export default function Home() {
  const [prompt, setPrompt] = useState(INITIAL_PROMPT);
  const [specs, setSpecs] = useState(INITIAL_SPECS);
  const [activeIndex, setActiveIndex] = useState(0);
  const [device, setDevice] = useState<DeviceName>("desktop");
  const [generating, setGenerating] = useState(false);
  const [mode, setMode] = useState<"jev" | "local">("local");
  const [latency, setLatency] = useState(384);
  const [showDecisions, setShowDecisions] = useState(true);
  const [history, setHistory] = useState<string[]>([INITIAL_PROMPT]);
  const activeSpec = specs[activeIndex] ?? specs[0];

  const generate = useCallback(async (nextPrompt: string) => {
    const cleanPrompt = nextPrompt.trim();
    if (!cleanPrompt || generating) return;
    setGenerating(true);
    setHistory((current) => [...current.slice(-3), cleanPrompt]);
    const started = performance.now();
    try {
      const response = await fetch("/api/design", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ prompt: cleanPrompt, currentSpecs: specs }) });
      if (!response.ok) throw new Error("Design request failed");
      const result = (await response.json()) as DesignResponse;
      setSpecs(result.specs); setMode(result.mode); setLatency(Math.max(result.latencyMs, Math.round(performance.now() - started))); setActiveIndex(0);
    } catch { setMode("local"); setLatency(Math.round(performance.now() - started)); }
    finally { setGenerating(false); }
  }, [generating, specs]);

  function submit(event: FormEvent) { event.preventDefault(); void generate(prompt); }

  useEffect(() => {
    const modelContext = (document as Document & { modelContext?: { registerTool?: (tool: unknown, options?: { signal?: AbortSignal }) => void | Promise<void> } }).modelContext;
    if (!modelContext?.registerTool) return;
    const lifecycle = new AbortController();
    try {
      void Promise.resolve(modelContext.registerTool({
        name: "generate_ui_concepts", title: "Generate UI concepts", description: "Generate three live website concepts from a natural-language design brief.",
        inputSchema: { type: "object", properties: { prompt: { type: "string", minLength: 3 } }, required: ["prompt"], additionalProperties: false },
        annotations: { readOnlyHint: false, untrustedContentHint: false },
        execute: async (input: unknown) => { const value = input as { prompt?: unknown }; if (typeof value.prompt !== "string" || value.prompt.trim().length < 3) throw new Error("A design brief is required"); setPrompt(value.prompt); await generate(value.prompt); return { status: "generated", concepts: 3 }; },
      }, { signal: lifecycle.signal })).catch(() => undefined);
    } catch { /* Optional in browsers without WebMCP. */ }
    return () => lifecycle.abort();
  }, [generate]);

  const statusText = useMemo(() => mode === "jev" ? "Jev 1.13" : "Local decision engine", [mode]);

  return (
    <main className="studio-shell">
      <header className="studio-header"><div className="brand"><span><Sparkles size={15} /></span><strong>Forge</strong><small>UI compiler</small></div><div className="project-title"><span>Velocity / Homepage</span><ChevronDown size={13} /></div><div className="header-actions"><span className="engine-status"><i className={mode === "jev" ? "live" : ""} />{statusText}</span><button className="icon-button" aria-label="Toggle theme"><Moon size={15} /></button><button className="export-button"><Code2 size={13} /> Export</button></div></header>
      <div className={`studio-grid ${showDecisions ? "" : "rail-hidden"}`}>
        <aside className="prompt-panel"><div className="panel-heading"><div><small>01</small><span>Describe</span></div><button aria-label="New design"><Plus size={15} /></button></div><div className="conversation"><div className="system-note"><Sparkles size={13} /><p><strong>Ready to design.</strong><br />Describe the page, audience, and feeling you want.</p></div>{history.map((item, index) => <div className="user-message" key={`${item}-${index}`}>{item}</div>)}{generating && <div className="thinking-message"><Loader2 size={13} /> Evaluating design space…</div>}</div><div className="quick-prompts">{QUICK_PROMPTS.map((item) => <button key={item} onClick={() => { setPrompt(item); void generate(item); }}>{item}<ArrowRight size={11} /></button>)}</div><form className="prompt-box" onSubmit={submit}><textarea value={prompt} onChange={(event) => setPrompt(event.target.value)} aria-label="Design prompt" /><div><span><kbd>⌘</kbd><kbd>↵</kbd> to generate</span><button type="submit" disabled={generating} aria-label="Generate concepts">{generating ? <Loader2 className="spin" size={15} /> : <Send size={14} />}</button></div></form></aside>
        <section className="canvas-panel"><div className="canvas-toolbar"><div className="concept-tabs">{specs.map((spec, index) => <button className={index === activeIndex ? "active" : ""} key={spec.id} onClick={() => setActiveIndex(index)}><span>{String.fromCharCode(65 + index)}</span><div><strong>{spec.name}</strong><small>{spec.descriptor}</small></div></button>)}</div><div className="view-tools"><div><button className={device === "desktop" ? "active" : ""} onClick={() => setDevice("desktop")} aria-label="Desktop preview"><Monitor size={14} /></button><button className={device === "mobile" ? "active" : ""} onClick={() => setDevice("mobile")} aria-label="Mobile preview"><Smartphone size={14} /></button></div><button onClick={() => setShowDecisions((value) => !value)} className={showDecisions ? "active" : ""}><Eye size={14} /><span>Decisions</span></button></div></div><div className="canvas-stage"><div className={`browser-frame ${device}`}><div className="browser-bar"><div><span /><span /><span /></div><p><i><Gauge size={10} /></i>velocity.dev</p><RefreshCw size={11} /></div><div className="site-scroll"><PreviewPage spec={activeSpec} /></div></div><div className="build-status"><span><i className={generating ? "processing" : ""} />{generating ? "Compiling decisions" : "Compiled"}</span><b>{latency} ms</b><span>3 concepts</span><span>0 errors</span></div></div></section>
        {showDecisions && <DecisionRail decisions={activeSpec.decisions} generating={generating} />}
      </div>
    </main>
  );
}
