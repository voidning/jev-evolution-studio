export namespace main {
	
	export class ApplyRequest {
	    prompt: string;
	    session: string;
	    targetId: string;
	
	    static createFrom(source: any = {}) {
	        return new ApplyRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.prompt = source["prompt"];
	        this.session = source["session"];
	        this.targetId = source["targetId"];
	    }
	}
	export class ColorValue {
	    kind: string;
	    value: string;
	    shade?: string;
	
	    static createFrom(source: any = {}) {
	        return new ColorValue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.value = source["value"];
	        this.shade = source["shade"];
	    }
	}
	export class IntentArguments {
	    channel?: string;
	    color?: ColorValue;
	    utility?: string;
	    element?: string;
	    placement?: string;
	    label?: string;
	
	    static createFrom(source: any = {}) {
	        return new IntentArguments(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.channel = source["channel"];
	        this.color = this.convertValues(source["color"], ColorValue);
	        this.utility = source["utility"];
	        this.element = source["element"];
	        this.placement = source["placement"];
	        this.label = source["label"];
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
	export class EditIntent {
	    arguments: IntentArguments;
	    state?: string;
	    scope: string;
	    family: string;
	    operation: string;
	    direction: string;
	    magnitude: string;
	    breakpoint: string;
	    value?: string;
	    preserve: string[];
	    confidence: number;
	
	    static createFrom(source: any = {}) {
	        return new EditIntent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.arguments = this.convertValues(source["arguments"], IntentArguments);
	        this.state = source["state"];
	        this.scope = source["scope"];
	        this.family = source["family"];
	        this.operation = source["operation"];
	        this.direction = source["direction"];
	        this.magnitude = source["magnitude"];
	        this.breakpoint = source["breakpoint"];
	        this.value = source["value"];
	        this.preserve = source["preserve"];
	        this.confidence = source["confidence"];
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
	export class Stage {
	    name: string;
	    ms: number;
	
	    static createFrom(source: any = {}) {
	        return new Stage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.ms = source["ms"];
	    }
	}
	export class Target {
	    revision: string;
	    fingerprint: string;
	    start: number;
	    end: number;
	    className: string;
	    classKind: string;
	    repeated: boolean;
	    empty: boolean;
	    id: string;
	    tag: string;
	    source: string;
	    line: number;
	    parent: string;
	    styles: Record<string, string>;
	    important: string[];
	    count: number;
	    width: number;
	
	    static createFrom(source: any = {}) {
	        return new Target(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.revision = source["revision"];
	        this.fingerprint = source["fingerprint"];
	        this.start = source["start"];
	        this.end = source["end"];
	        this.className = source["className"];
	        this.classKind = source["classKind"];
	        this.repeated = source["repeated"];
	        this.empty = source["empty"];
	        this.id = source["id"];
	        this.tag = source["tag"];
	        this.source = source["source"];
	        this.line = source["line"];
	        this.parent = source["parent"];
	        this.styles = source["styles"];
	        this.important = source["important"];
	        this.count = source["count"];
	        this.width = source["width"];
	    }
	}
	export class EditorState {
	    files: string[];
	    executor: string;
	    project: string;
	    session: string;
	    previewURL: string;
	    connected: boolean;
	    browserConnected: boolean;
	    selected?: Target;
	    hasAPIKey: boolean;
	    mode: string;
	    busy: boolean;
	    pending: boolean;
	    canUndo: boolean;
	    selecting: boolean;
	    intents: EditIntent[];
	    diff: string;
	    error: string;
	    stages: Stage[];
	    changeId: string;
	
	    static createFrom(source: any = {}) {
	        return new EditorState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.files = source["files"];
	        this.executor = source["executor"];
	        this.project = source["project"];
	        this.session = source["session"];
	        this.previewURL = source["previewURL"];
	        this.connected = source["connected"];
	        this.browserConnected = source["browserConnected"];
	        this.selected = this.convertValues(source["selected"], Target);
	        this.hasAPIKey = source["hasAPIKey"];
	        this.mode = source["mode"];
	        this.busy = source["busy"];
	        this.pending = source["pending"];
	        this.canUndo = source["canUndo"];
	        this.selecting = source["selecting"];
	        this.intents = this.convertValues(source["intents"], EditIntent);
	        this.diff = source["diff"];
	        this.error = source["error"];
	        this.stages = this.convertValues(source["stages"], Stage);
	        this.changeId = source["changeId"];
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

