export namespace main {
	
	export class BlueprintItem {
	    label: string;
	    title: string;
	    body: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new BlueprintItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.title = source["title"];
	        this.body = source["body"];
	        this.value = source["value"];
	    }
	}
	export class NodeContent {
	    eyebrow?: string;
	    headline?: string;
	    body?: string;
	    value?: string;
	    items?: BlueprintItem[];
	
	    static createFrom(source: any = {}) {
	        return new NodeContent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.eyebrow = source["eyebrow"];
	        this.headline = source["headline"];
	        this.body = source["body"];
	        this.value = source["value"];
	        this.items = this.convertValues(source["items"], BlueprintItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class InteractionSpec {
	    trigger?: string;
	    motion?: string;
	    strength?: number;
	
	    static createFrom(source: any = {}) {
	        return new InteractionSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.trigger = source["trigger"];
	        this.motion = source["motion"];
	        this.strength = source["strength"];
	    }
	}
	export class VisualSpec {
	    kind?: string;
	    position?: string;
	    intensity?: number;
	
	    static createFrom(source: any = {}) {
	        return new VisualSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.position = source["position"];
	        this.intensity = source["intensity"];
	    }
	}
	export class StyleSpec {
	    surface?: string;
	    scale?: string;
	    emphasis?: string;
	    shape?: string;
	
	    static createFrom(source: any = {}) {
	        return new StyleSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.surface = source["surface"];
	        this.scale = source["scale"];
	        this.emphasis = source["emphasis"];
	        this.shape = source["shape"];
	    }
	}
	export class LayoutSpec {
	    axis?: string;
	    columns?: number;
	    gap?: string;
	    span?: number;
	    align?: string;
	    minHeight?: string;
	    inset?: string;
	    reverse?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LayoutSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.axis = source["axis"];
	        this.columns = source["columns"];
	        this.gap = source["gap"];
	        this.span = source["span"];
	        this.align = source["align"];
	        this.minHeight = source["minHeight"];
	        this.inset = source["inset"];
	        this.reverse = source["reverse"];
	    }
	}
	export class DesignNode {
	    id: string;
	    primitive: string;
	    role?: string;
	    layout?: LayoutSpec;
	    style?: StyleSpec;
	    visual?: VisualSpec;
	    interaction?: InteractionSpec;
	    content?: NodeContent;
	    children?: DesignNode[];
	
	    static createFrom(source: any = {}) {
	        return new DesignNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.primitive = source["primitive"];
	        this.role = source["role"];
	        this.layout = this.convertValues(source["layout"], LayoutSpec);
	        this.style = this.convertValues(source["style"], StyleSpec);
	        this.visual = this.convertValues(source["visual"], VisualSpec);
	        this.interaction = this.convertValues(source["interaction"], InteractionSpec);
	        this.content = this.convertValues(source["content"], NodeContent);
	        this.children = this.convertValues(source["children"], DesignNode);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PageBlueprint {
	    version: number;
	    id: string;
	    creativeDirection: string;
	    root: DesignNode;
	
	    static createFrom(source: any = {}) {
	        return new PageBlueprint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.id = source["id"];
	        this.creativeDirection = source["creativeDirection"];
	        this.root = this.convertValues(source["root"], DesignNode);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DesignControls {
	    spatialTension: number;
	    visualAbstraction: number;
	    informationDensity: number;
	    narrativeDepth: number;
	    trustPriority: number;
	    motionEnergy: number;
	    symmetry: number;
	
	    static createFrom(source: any = {}) {
	        return new DesignControls(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.spatialTension = source["spatialTension"];
	        this.visualAbstraction = source["visualAbstraction"];
	        this.informationDensity = source["informationDensity"];
	        this.narrativeDepth = source["narrativeDepth"];
	        this.trustPriority = source["trustPriority"];
	        this.motionEnergy = source["motionEnergy"];
	        this.symmetry = source["symmetry"];
	    }
	}
	export class Scorecard {
	    originality: number;
	    clarity: number;
	    trust: number;
	    conversion: number;
	    composite: number;
	
	    static createFrom(source: any = {}) {
	        return new Scorecard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.originality = source["originality"];
	        this.clarity = source["clarity"];
	        this.trust = source["trust"];
	        this.conversion = source["conversion"];
	        this.composite = source["composite"];
	    }
	}
	export class Decision {
	    label: string;
	    value: string;
	    confidence: number;
	    group: string;
	    distribution?: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new Decision(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.value = source["value"];
	        this.confidence = source["confidence"];
	        this.group = source["group"];
	        this.distribution = source["distribution"];
	    }
	}
	export class MetricContent {
	    value: string;
	    unit: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new MetricContent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.unit = source["unit"];
	        this.label = source["label"];
	    }
	}
	export class FeatureContent {
	    kicker: string;
	    title: string;
	    body: string;
	
	    static createFrom(source: any = {}) {
	        return new FeatureContent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kicker = source["kicker"];
	        this.title = source["title"];
	        this.body = source["body"];
	    }
	}
	export class PageSpec {
	    id: string;
	    name: string;
	    descriptor: string;
	    strategy: string;
	    generation: number;
	    mutation: string;
	    theme: string;
	    hero: string;
	    visual: string;
	    features: string;
	    density: string;
	    navigation: string;
	    motion: string;
	    story: string;
	    cta: string;
	    world: string;
	    brand: string;
	    eyebrow: string;
	    sectionLabel: string;
	    sectionTitle: string;
	    featuresContent: FeatureContent[];
	    metrics: MetricContent[];
	    showLogos: boolean;
	    showPricing: boolean;
	    showStats: boolean;
	    title: string;
	    description: string;
	    decisions: Decision[];
	    scores: Scorecard;
	    controls: DesignControls;
	    blueprint: PageBlueprint;
	
	    static createFrom(source: any = {}) {
	        return new PageSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.descriptor = source["descriptor"];
	        this.strategy = source["strategy"];
	        this.generation = source["generation"];
	        this.mutation = source["mutation"];
	        this.theme = source["theme"];
	        this.hero = source["hero"];
	        this.visual = source["visual"];
	        this.features = source["features"];
	        this.density = source["density"];
	        this.navigation = source["navigation"];
	        this.motion = source["motion"];
	        this.story = source["story"];
	        this.cta = source["cta"];
	        this.world = source["world"];
	        this.brand = source["brand"];
	        this.eyebrow = source["eyebrow"];
	        this.sectionLabel = source["sectionLabel"];
	        this.sectionTitle = source["sectionTitle"];
	        this.featuresContent = this.convertValues(source["featuresContent"], FeatureContent);
	        this.metrics = this.convertValues(source["metrics"], MetricContent);
	        this.showLogos = source["showLogos"];
	        this.showPricing = source["showPricing"];
	        this.showStats = source["showStats"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.decisions = this.convertValues(source["decisions"], Decision);
	        this.scores = this.convertValues(source["scores"], Scorecard);
	        this.controls = this.convertValues(source["controls"], DesignControls);
	        this.blueprint = this.convertValues(source["blueprint"], PageBlueprint);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DesignResult {
	    mode: string;
	    latencyMs: number;
	    prompt: string;
	    generation: number;
	    winner: number;
	    swarmSize: number;
	    mutationLog: string[];
	    specs: PageSpec[];
	    generator: string;
	    astSource: string;
	    candidateCount: number;
	    designSpace: string;
	
	    static createFrom(source: any = {}) {
	        return new DesignResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.latencyMs = source["latencyMs"];
	        this.prompt = source["prompt"];
	        this.generation = source["generation"];
	        this.winner = source["winner"];
	        this.swarmSize = source["swarmSize"];
	        this.mutationLog = source["mutationLog"];
	        this.specs = this.convertValues(source["specs"], PageSpec);
	        this.generator = source["generator"];
	        this.astSource = source["astSource"];
	        this.candidateCount = source["candidateCount"];
	        this.designSpace = source["designSpace"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AppState {
	    previewURL: string;
	    result: DesignResult;
	    hasAPIKey: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.previewURL = source["previewURL"];
	        this.result = this.convertValues(source["result"], DesignResult);
	        this.hasAPIKey = source["hasAPIKey"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	
	
	
	
	
	
	
	
	
	

}

