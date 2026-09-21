import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { ArrowRight, Check, ChevronRight, CircleDot, Command, ExternalLink, Gauge, GitBranch, Layers3, Loader2, MonitorUp, Orbit, Radio, RefreshCw, Send, Sparkles, Trophy, Zap } from "lucide-react";
import { Evolve, Generate, GetState, OpenStage, SelectConcept } from "../wailsjs/go/main/App";

type Decision = { label: string; value: string; confidence: number; group: string; distribution?: Record<string, number> };
type Scorecard = { originality: number; clarity: number; trust: number; conversion: number; composite: number };
type Blueprint = { version: number; id: string; creativeDirection: string; sections: unknown[] };
type PageSpec = { id: string; name: string; descriptor: string; strategy: string; generation: number; mutation: string; theme: string; hero: string; visual: string; features: string; density: string; navigation: string; motion: string; story: string; cta: string; world: string; brand: string; eyebrow: string; showLogos: boolean; showPricing: boolean; showStats: boolean; title: string; description: string; decisions: Decision[]; scores: Scorecard; blueprint: Blueprint };
type DesignResult = { mode: string; latencyMs: number; prompt: string; generation: number; winner: number; swarmSize: number; mutationLog: string[]; specs: PageSpec[]; generator: string; astSource: string };

const DEFAULT_PROMPT = "为一款实时 AI 数据分析产品做主页。面向开发者，深色、克制、有速度感，重点突出实时分析。";
const QUICK = ["做得像来自未来，但必须可信", "锁住技术感，增加人类情绪", "极端原创，并强化申请内测"];
const PHASES = ["Jev parallel intent", "144 AST candidates", "Constraint tournament", "Critic scoring", "Gen 02 mutation"];

function App() {
  const [prompt, setPrompt] = useState(DEFAULT_PROMPT);
  const [result, setResult] = useState<DesignResult | null>(null);
  const [active, setActive] = useState(0);
  const [generating, setGenerating] = useState(false);
  const [stageURL, setStageURL] = useState("");
  const [hasKey, setHasKey] = useState(false);
  const [autoApply, setAutoApply] = useState(false);
  const [stageOpened, setStageOpened] = useState(false);
  const [phase, setPhase] = useState(0);
  const lastGenerated = useRef("");

  useEffect(() => {
    GetState().then((state) => {
      setResult(state.result as DesignResult);
      setStageURL(state.previewURL);
      setHasKey(state.hasAPIKey);
      return OpenStage();
    }).then(() => setStageOpened(true)).catch(console.error);
  }, []);

  useEffect(() => {
    if (!autoApply || prompt.trim().length < 3 || prompt === lastGenerated.current) return;
    const timer = window.setTimeout(() => void runGenerate(prompt), 700);
    return () => window.clearTimeout(timer);
  }, [prompt, autoApply]);

  async function runGenerate(value: string) {
    const clean = value.trim();
    if (!clean || generating) return;
    lastGenerated.current = clean;
    setGenerating(true);
    setPhase(0);
    const choreography = window.setInterval(() => setPhase((value) => Math.min(value + 1, PHASES.length - 1)), 620);
    try {
      const next = await Generate(clean) as DesignResult;
      setResult(next);
      setActive(next.winner ?? 0);
    } finally { window.clearInterval(choreography); setGenerating(false); setPhase(0); }
  }

  function submit(event: FormEvent) { event.preventDefault(); void runGenerate(prompt); }
  async function choose(index: number) { setActive(index); await SelectConcept(index); }
  async function openStage() { await OpenStage(); setStageOpened(true); }
  async function evolveAgain() {
    if (generating) return;
    setGenerating(true); setPhase(3);
    try { const next = await Evolve() as DesignResult; setResult(next); setActive(next.winner ?? 0); }
    finally { setGenerating(false); setPhase(0); }
  }

  const spec = result?.specs?.[active];
  const groups = useMemo(() => ["Intent", "Experience", "Visual", "Signals"].map((name) => ({ name, items: spec?.decisions.filter((item) => item.group === name) ?? [] })), [spec]);
  const safeURL = stageURL ? stageURL.replace(/\?token=.*/, "") : "Starting local stage…";

  return (
    <main className="director">
      <header className="titlebar">
        <div className="brand"><span><Sparkles size={14} /></span><strong>Forge</strong><small>LIVE UI DIRECTOR</small></div>
        <div className="engine"><i className={result?.mode?.includes("jev") ? "jev" : ""} />{hasKey ? "JEV + PROCEDURAL AST" : "LOCAL PROCEDURAL AST"}</div>
      </header>

      <section className="stage-strip">
        <div className="stage-icon"><MonitorUp size={17} /></div>
        <div><small>LIVE BROWSER STAGE</small><strong>{safeURL}</strong></div>
        <span className={stageOpened ? "connected" : ""}><Radio size={11} /> {stageOpened ? "CONNECTED" : "WAITING"}</span>
        <button onClick={openStage}><ExternalLink size={13} /> Open</button>
      </section>

      <section className="prompt-section">
        <div className="section-heading"><span><b>01</b> Design direction</span><label className="auto"><input type="checkbox" checked={autoApply} onChange={(e) => setAutoApply(e.target.checked)} /><i /> AUTO APPLY</label></div>
        <form onSubmit={submit} className="prompt-card">
          <textarea value={prompt} onChange={(e) => setPrompt(e.target.value)} onKeyDown={(event) => { if (event.key === "Enter" && !event.shiftKey) { event.preventDefault(); void runGenerate(prompt); } }} aria-label="Design prompt" />
          <div className="prompt-actions"><span><Command size={11} /> ENTER COMPILE · SHIFT+ENTER LINE</span><button disabled={generating} aria-label="Compile design">{generating ? <Loader2 className="spin" size={16} /> : <Send size={15} />}</button></div>
        </form>
        <div className="quick-row">{QUICK.map((item) => <button key={item} onClick={() => { setPrompt(item); void runGenerate(item); }}>{item}<ArrowRight size={10} /></button>)}</div>
        <div className={`choreography ${generating ? "active" : ""}`}>{PHASES.map((item, index) => <div key={item} className={index < phase ? "done" : index === phase ? "running" : ""}><span>{index < phase ? <Check size={9} /> : index === phase ? <Loader2 size={9} className="spin" /> : index + 1}</span>{item}</div>)}</div>
      </section>

      <section className="concept-section">
        <div className="section-heading"><span><b>02</b> Universe tournament</span><small>{generating ? "EVOLVING…" : `GEN ${String(result?.generation ?? 1).padStart(2, "0")}`}</small></div>
        <div className="concept-grid">
          {result?.specs.map((item, index) => (
            <button key={item.id} className={`concept ${active === index ? "active" : ""}`} onClick={() => void choose(index)}>
              <span className={`swatch ${item.theme}`}><b>{String.fromCharCode(65 + index)}</b><i /></span>
              <div><strong>{item.name}</strong><small>{item.strategy} · {item.scores.composite || "—"}</small></div>
              {result?.winner === index ? <Trophy size={12} /> : active === index ? <Check size={13} /> : <ChevronRight size={13} />}
            </button>
          ))}
        </div>
        {spec && <div className="scoreboard">{(["originality", "clarity", "trust", "conversion"] as const).map((key) => <div key={key}><span>{key}</span><i><b style={{ width: `${spec.scores[key] || 0}%` }} /></i><strong>{spec.scores[key] || "—"}</strong></div>)}</div>}
      </section>

      <section className="graph-section">
        <div className="section-heading"><span><b>03</b> Decision swarm</span><small>{result?.swarmSize ?? spec?.decisions.length ?? 0} PARALLEL</small></div>
        <div className="source"><Zap size={13} /><span>Natural-language intent</span><i /><em>JEV FAN-OUT</em></div>
        <div className={`graph ${generating ? "resolving" : ""}`}>
          {groups.map((group, groupIndex) => (
            <div className="group" key={group.name}>
              <div className="group-name"><span>0{groupIndex + 1}</span>{group.name}</div>
              <div className="nodes">{group.items.map((node, nodeIndex) => <article className="node" key={`${group.name}-${node.label}-${nodeIndex}`}><div><small>{node.label}</small><strong>{node.value}</strong><i><b style={{ width: `${Math.round(node.confidence * 100)}%` }} /></i></div><span>{Math.round(node.confidence * 100)}</span></article>)}</div>
            </div>
          ))}
        </div>
      </section>

      <section className="evolution-section">
        <div className="section-heading"><span><b>04</b> Recursive evolution</span><small>{result?.mutationLog?.length ?? 0} MUTATIONS</small></div>
        <div className="evolution-card"><div className="generation"><GitBranch size={14} /><span>GEN {String(Math.max(1, (result?.generation ?? 1) - 1)).padStart(2, "0")}</span><i /><Orbit size={14} /><strong>GEN {String(result?.generation ?? 1).padStart(2, "0")}</strong></div><div className="mutations">{result?.mutationLog?.map((item) => <span key={item}>{item}</span>) ?? <span>Waiting for the first mutation cycle</span>}</div><button disabled={generating} onClick={() => void evolveAgain()}><RefreshCw size={11} className={generating ? "spin" : ""} /> EVOLVE AGAIN</button></div>
      </section>

      <footer className="statusbar">
        <span><CircleDot size={10} /> {generating ? PHASES[phase] : "Genome synced"}</span>
        <span><Gauge size={10} /> {result?.latencyMs ?? 0} ms</span>
        <span><Layers3 size={10} /> {result?.specs?.[0]?.blueprint?.sections?.length ?? 0} AST nodes</span>
      </footer>
    </main>
  );
}

export default App;
