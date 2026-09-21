import { NextRequest, NextResponse } from "next/server";

type ThemeName = "violet" | "ember" | "mint";
type HeroName = "split" | "centered" | "terminal";
type FeatureName = "bento" | "columns" | "timeline";
type Decision = { label: string; value: string; confidence: number; group: "Intent" | "Structure" | "Visual" };
type PageSpec = {
  id: string; name: string; descriptor: string; theme: ThemeName; hero: HeroName;
  features: FeatureName; density: "airy" | "balanced" | "compact"; showLogos: boolean;
  showPricing: boolean; showStats: boolean; title: string; description: string; decisions: Decision[];
};

type Answers = Record<string, { choice?: string; confidence?: number; noul?: number; score?: number; probabilities?: Record<string, number> }>;

const COPY = {
  technical: { title: "See the signal.\nBefore everyone else.", description: "Turn live product data into decisions your team can act on — without waiting for another dashboard refresh." },
  bold: { title: "Your product,\nat live speed.", description: "One continuous stream from event to insight. Query, detect, and ship the next decision in milliseconds." },
  editorial: { title: "Intelligence that\nkeeps up.", description: "A calmer, clearer way to understand every moving part of your product — as it happens." },
};

function includesAny(prompt: string, terms: string[]) { return terms.some((term) => prompt.toLowerCase().includes(term)); }

function localAnswers(prompt: string): Answers {
  const warm = includesAny(prompt, ["暖", "warm", "橙", "人文"]);
  const bold = includesAny(prompt, ["大胆", "bold", "冲击", "未来"]);
  const compact = includesAny(prompt, ["紧凑", "dense", "信息密度", "数据密集"]);
  const airy = includesAny(prompt, ["留白", "克制", "极简", "airy", "calm"]);
  return {
    archetype: { choice: includesAny(prompt, ["企业", "b2b", "机构"]) ? "enterprise" : "developer", confidence: .9 },
    goal: { choice: includesAny(prompt, ["购买", "转化", "定价", "pricing"]) ? "convert" : "explore", confidence: .86 },
    tone: { choice: warm ? "editorial" : bold ? "bold" : "technical", confidence: .89 },
    theme: { choice: warm ? "ember" : bold ? "mint" : "violet", confidence: .92 },
    hero: { choice: bold ? "terminal" : airy ? "centered" : "split", confidence: .87 },
    features: { choice: compact ? "timeline" : airy ? "columns" : "bento", confidence: .82 },
    density: { choice: compact ? "compact" : airy ? "airy" : "balanced", confidence: .79 },
    pricing: { noul: includesAny(prompt, ["定价", "pricing", "价格", "套餐"]) ? .94 : includesAny(prompt, ["删除定价", "不要定价"]) ? .05 : .28 },
    logos: { noul: includesAny(prompt, ["logo", "客户", "信任", "企业"]) ? .88 : .68 },
    stats: { noul: includesAny(prompt, ["数据", "实时", "速度", "分析", "指标"]) ? .93 : .61 },
  };
}

function pick<T extends string>(answer: Answers[string] | undefined, fallback: T, allowed: readonly T[]): T {
  return answer?.choice && allowed.includes(answer.choice as T) ? answer.choice as T : fallback;
}

function confidence(answer: Answers[string] | undefined, fallback: number) { return Math.max(.5, Math.min(.99, answer?.confidence ?? fallback)); }

function makeSpec(id: string, name: string, descriptor: string, answers: Answers, overrides: Partial<PageSpec> = {}): PageSpec {
  const theme = pick(answers.theme, "violet", ["violet", "ember", "mint"] as const);
  const hero = pick(answers.hero, "split", ["split", "centered", "terminal"] as const);
  const features = pick(answers.features, "bento", ["bento", "columns", "timeline"] as const);
  const density = pick(answers.density, "balanced", ["airy", "balanced", "compact"] as const);
  const tone = pick(answers.tone, "technical", ["technical", "bold", "editorial"] as const);
  const copy = COPY[tone];
  const archetype = pick(answers.archetype, "developer", ["developer", "enterprise", "data_product"] as const);
  const goal = pick(answers.goal, "explore", ["explore", "convert", "trust"] as const);
  const finalTheme = overrides.theme ?? theme;
  const finalHero = overrides.hero ?? hero;
  const finalFeatures = overrides.features ?? features;
  const finalDensity = overrides.density ?? density;
  const decisions: Decision[] = [
    { label: "Archetype", value: archetype === "enterprise" ? "Enterprise SaaS" : archetype === "data_product" ? "Data product" : "Developer product", confidence: confidence(answers.archetype, .9), group: "Intent" },
    { label: "Primary goal", value: goal === "convert" ? "Drive conversion" : goal === "trust" ? "Build trust" : "Explore product", confidence: confidence(answers.goal, .84), group: "Intent" },
    { label: "Hero", value: finalHero === "terminal" ? "Terminal first" : finalHero === "centered" ? "Centered product" : "Split dashboard", confidence: confidence(answers.hero, .86), group: "Structure" },
    { label: "Features", value: finalFeatures === "timeline" ? "Live timeline" : finalFeatures === "columns" ? "Editorial columns" : "Asymmetric bento", confidence: confidence(answers.features, .82), group: "Structure" },
    { label: "Visual tone", value: tone === "bold" ? "Bold technical" : tone === "editorial" ? "Editorial systems" : "Technical minimal", confidence: confidence(answers.tone, .9), group: "Visual" },
    { label: "Density", value: finalDensity[0].toUpperCase() + finalDensity.slice(1), confidence: confidence(answers.density, .78), group: "Visual" },
  ];
  return {
    id, name, descriptor, theme: finalTheme, hero: finalHero, features: finalFeatures, density: finalDensity,
    showLogos: overrides.showLogos ?? (answers.logos?.noul ?? .6) > .5,
    showPricing: overrides.showPricing ?? (answers.pricing?.noul ?? .2) > .55,
    showStats: overrides.showStats ?? (answers.stats?.noul ?? .7) > .5,
    title: overrides.title ?? copy.title, description: overrides.description ?? copy.description, decisions,
  };
}

function compileConcepts(answers: Answers): PageSpec[] {
  const baseTheme = pick(answers.theme, "violet", ["violet", "ember", "mint"] as const);
  const alternatives: Record<ThemeName, [ThemeName, ThemeName]> = { violet: ["mint", "ember"], mint: ["violet", "ember"], ember: ["violet", "mint"] };
  const [secondTheme, thirdTheme] = alternatives[baseTheme];
  const base = makeSpec("signal", "Signal", "Best semantic match", answers);
  const second = makeSpec("pulse", "Pulse", "Higher energy", answers, { theme: secondTheme, hero: base.hero === "terminal" ? "split" : "terminal", features: "timeline", density: "compact", showPricing: false });
  const third = makeSpec("orbit", "Orbit", "Calmer editorial", answers, { theme: thirdTheme, hero: "centered", features: "columns", density: "airy", showLogos: true });
  return [base, second, third];
}

async function askJev(prompt: string, currentSpecs: unknown): Promise<Answers | null> {
  const apiKey = process.env.TYPESAFE_API_KEY;
  if (!apiKey) return null;
  const response = await fetch("https://api.typesafe.ai/v1/systemone", {
    method: "POST",
    headers: { Authorization: `Bearer ${apiKey}`, "Content-Type": "application/json" },
    body: JSON.stringify({
      model: "jev-1.13.0",
      state: { request: prompt, current_page: currentSpecs, instruction: "Interpret this as a UI design or edit request. Preserve current choices unless the request asks to change them." },
      questions: {
        archetype: { type: "choice", instructions: "Choose the best page archetype for this request", criteria: { developer: "A developer tool or technical product", enterprise: "A trust-oriented B2B or enterprise product", data_product: "A data, analytics, or intelligence product" } },
        goal: { type: "choice", instructions: "Choose the primary job of the page", criteria: { explore: "Help visitors understand and explore the product", convert: "Drive signup or purchase conversion", trust: "Build credibility and reduce perceived risk" } },
        tone: { type: "choice", instructions: "Choose the best visual tone", criteria: { technical: "Precise, restrained, technical minimalism", bold: "Energetic, futuristic, high-impact", editorial: "Calm, warm, spacious, editorial" } },
        theme: { type: "choice", instructions: "Choose the accent color family", criteria: { violet: "Electric violet for intelligent technical products", mint: "Signal green for speed, live systems, and energy", ember: "Warm orange for a more human editorial feel" } },
        hero: { type: "choice", instructions: "Choose the best hero layout", criteria: { split: "Copy beside a live product dashboard", centered: "Centered message above the product", terminal: "Code terminal is the primary product proof" } },
        features: { type: "choice", instructions: "Choose the best feature presentation", criteria: { bento: "Asymmetric product capability cards", columns: "Calm editorial columns", timeline: "Compact continuous event flow" } },
        density: { type: "choice", instructions: "Choose the appropriate information density", criteria: { airy: "Spacious and focused", balanced: "Balanced detail and breathing room", compact: "Dense, fast, information rich" } },
        pricing: { type: "noul", instructions: "The requested page should contain a pricing section" },
        logos: { type: "noul", instructions: "The requested page should show a customer logo trust strip" },
        stats: { type: "noul", instructions: "The requested page should show quantitative product metrics" },
      },
    }),
    cache: "no-store",
  });
  if (!response.ok) return null;
  const data = await response.json() as { answers?: Answers };
  return data.answers ?? null;
}

export async function POST(request: NextRequest) {
  const started = Date.now();
  const body = await request.json().catch(() => ({})) as { prompt?: unknown; currentSpecs?: unknown };
  if (typeof body.prompt !== "string" || body.prompt.trim().length < 3) return NextResponse.json({ error: "A design prompt is required" }, { status: 400 });
  const prompt = body.prompt.trim().slice(0, 2000);
  let answers: Answers | null = null;
  try { answers = await askJev(prompt, body.currentSpecs); } catch { answers = null; }
  const mode = answers ? "jev" : "local";
  answers ??= localAnswers(prompt);
  return NextResponse.json({ mode, latencyMs: Date.now() - started, prompt, specs: compileConcepts(answers) });
}
