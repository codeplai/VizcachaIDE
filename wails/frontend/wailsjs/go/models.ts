export namespace app {
	
	export class LanguageRegistry {
	
	
	    static createFrom(source: any = {}) {
	        return new LanguageRegistry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

export namespace bridge {
	
	export class FilesService {
	
	
	    static createFrom(source: any = {}) {
	        return new FilesService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class SettingsService {
	
	
	    static createFrom(source: any = {}) {
	        return new SettingsService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

export namespace domain {
	
	export class SourceLocation {
	    file: string;
	    line: number;
	    column: number;
	
	    static createFrom(source: any = {}) {
	        return new SourceLocation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.line = source["line"];
	        this.column = source["column"];
	    }
	}
	export class Breakpoint {
	    location: SourceLocation;
	    condition: string;
	
	    static createFrom(source: any = {}) {
	        return new Breakpoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.location = this.convertValues(source["location"], SourceLocation);
	        this.condition = source["condition"];
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
	export class Capabilities {
	    build: boolean;
	    console: boolean;
	    format: boolean;
	    check: boolean;
	    debugInput: boolean;
	    packageActions: string[];
	    threadsLabel: string;
	
	    static createFrom(source: any = {}) {
	        return new Capabilities(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.build = source["build"];
	        this.console = source["console"];
	        this.format = source["format"];
	        this.check = source["check"];
	        this.debugInput = source["debugInput"];
	        this.packageActions = source["packageActions"];
	        this.threadsLabel = source["threadsLabel"];
	    }
	}
	export class CompletionItem {
	    label: string;
	    kind: string;
	    detail: string;
	    documentation: string;
	    insertText: string;
	
	    static createFrom(source: any = {}) {
	        return new CompletionItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.kind = source["kind"];
	        this.detail = source["detail"];
	        this.documentation = source["documentation"];
	        this.insertText = source["insertText"];
	    }
	}
	export class ConsoleResult {
	    result: string;
	    output: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ConsoleResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.result = source["result"];
	        this.output = source["output"];
	        this.error = source["error"];
	    }
	}
	export class Diagnostic {
	    location?: SourceLocation;
	    severity: string;
	    message: string;
	    rawText: string;
	    source: string;
	    code: string;
	    end?: SourceLocation;
	
	    static createFrom(source: any = {}) {
	        return new Diagnostic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.location = this.convertValues(source["location"], SourceLocation);
	        this.severity = source["severity"];
	        this.message = source["message"];
	        this.rawText = source["rawText"];
	        this.source = source["source"];
	        this.code = source["code"];
	        this.end = this.convertValues(source["end"], SourceLocation);
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
	export class SourceRange {
	    start: SourceLocation;
	    end: SourceLocation;
	
	    static createFrom(source: any = {}) {
	        return new SourceRange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.start = this.convertValues(source["start"], SourceLocation);
	        this.end = this.convertValues(source["end"], SourceLocation);
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
	export class DocumentSymbol {
	    name: string;
	    kind: string;
	    location: SourceLocation;
	    range?: SourceRange;
	    detail: string;
	    children: DocumentSymbol[];
	
	    static createFrom(source: any = {}) {
	        return new DocumentSymbol(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.location = this.convertValues(source["location"], SourceLocation);
	        this.range = this.convertValues(source["range"], SourceRange);
	        this.detail = source["detail"];
	        this.children = this.convertValues(source["children"], DocumentSymbol);
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
	export class ErrorExplanation {
	    explanationId: string;
	    title: string;
	    body: string;
	    fixHint: string;
	    placeholders: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new ErrorExplanation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.explanationId = source["explanationId"];
	        this.title = source["title"];
	        this.body = source["body"];
	        this.fixHint = source["fixHint"];
	        this.placeholders = source["placeholders"];
	    }
	}
	export class ExplainedDiagnostic {
	    diagnostic: Diagnostic;
	    explanation?: ErrorExplanation;
	
	    static createFrom(source: any = {}) {
	        return new ExplainedDiagnostic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.diagnostic = this.convertValues(source["diagnostic"], Diagnostic);
	        this.explanation = this.convertValues(source["explanation"], ErrorExplanation);
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
	export class FileNode {
	    name: string;
	    path: string;
	    isDir: boolean;
	    children: FileNode[];
	
	    static createFrom(source: any = {}) {
	        return new FileNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.isDir = source["isDir"];
	        this.children = this.convertValues(source["children"], FileNode);
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
	export class Variable {
	    name: string;
	    typeName: string;
	    value: string;
	    reference: number;
	    changed: boolean;
	    children: Variable[];
	
	    static createFrom(source: any = {}) {
	        return new Variable(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.typeName = source["typeName"];
	        this.value = source["value"];
	        this.reference = source["reference"];
	        this.changed = source["changed"];
	        this.children = this.convertValues(source["children"], Variable);
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
	export class FrameVariables {
	    arguments: Variable[];
	    locals: Variable[];
	
	    static createFrom(source: any = {}) {
	        return new FrameVariables(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.arguments = this.convertValues(source["arguments"], Variable);
	        this.locals = this.convertValues(source["locals"], Variable);
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
	export class IndentStyle {
	    useTabs: boolean;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new IndentStyle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.useTabs = source["useTabs"];
	        this.size = source["size"];
	    }
	}
	export class InlayHint {
	    line: number;
	    column: number;
	    label: string;
	    kind: string;
	    paddingLeft: boolean;
	    paddingRight: boolean;
	
	    static createFrom(source: any = {}) {
	        return new InlayHint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.line = source["line"];
	        this.column = source["column"];
	        this.label = source["label"];
	        this.kind = source["kind"];
	        this.paddingLeft = source["paddingLeft"];
	        this.paddingRight = source["paddingRight"];
	    }
	}
	export class ToolSpec {
	    id: string;
	    role: string;
	    labelKey: string;
	    missingKey: string;
	    installUrl: string;
	    installCommand: string;
	    providedBy: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.role = source["role"];
	        this.labelKey = source["labelKey"];
	        this.missingKey = source["missingKey"];
	        this.installUrl = source["installUrl"];
	        this.installCommand = source["installCommand"];
	        this.providedBy = source["providedBy"];
	    }
	}
	export class LanguageProfile {
	    id: string;
	    nameKey: string;
	    extensions: string[];
	    indent: IndentStyle;
	    capabilities: Capabilities;
	    tools: ToolSpec[];
	
	    static createFrom(source: any = {}) {
	        return new LanguageProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.nameKey = source["nameKey"];
	        this.extensions = source["extensions"];
	        this.indent = this.convertValues(source["indent"], IndentStyle);
	        this.capabilities = this.convertValues(source["capabilities"], Capabilities);
	        this.tools = this.convertValues(source["tools"], ToolSpec);
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
	export class NewProject {
	    root: string;
	    mainFile: string;
	
	    static createFrom(source: any = {}) {
	        return new NewProject(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.mainFile = source["mainFile"];
	    }
	}
	export class PackageInfo {
	    name: string;
	    version: string;
	    description: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new PackageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.description = source["description"];
	        this.url = source["url"];
	    }
	}
	export class ProjectContext {
	    root: string;
	    kind: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectContext(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.kind = source["kind"];
	        this.name = source["name"];
	    }
	}
	export class Release {
	    version: string;
	    notes: string;
	    notesUrl: string;
	    publishedAt: string;
	    asset: string;
	    assetSize: number;
	
	    static createFrom(source: any = {}) {
	        return new Release(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.notes = source["notes"];
	        this.notesUrl = source["notesUrl"];
	        this.publishedAt = source["publishedAt"];
	        this.asset = source["asset"];
	        this.assetSize = source["assetSize"];
	    }
	}
	export class RunConfiguration {
	    codeLanguage: string;
	    target: string;
	    workingDir: string;
	    mode: string;
	    programArgs: string[];
	    project?: ProjectContext;
	    echo: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RunConfiguration(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.codeLanguage = source["codeLanguage"];
	        this.target = source["target"];
	        this.workingDir = source["workingDir"];
	        this.mode = source["mode"];
	        this.programArgs = source["programArgs"];
	        this.project = this.convertValues(source["project"], ProjectContext);
	        this.echo = source["echo"];
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
	export class Settings {
	    language: string;
	    theme: string;
	    fontSize: number;
	    toolPaths: Record<string, string>;
	    firstRun: boolean;
	    lastFolder: string;
	    defaultCodeLanguage: string;
	    enabledCodeLanguages: string[];
	    formatOnSave: boolean;
	    checkUpdates: boolean;
	    lastUpdateCheck: string;
	    recentFiles: string[];
	    inlayHints: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.theme = source["theme"];
	        this.fontSize = source["fontSize"];
	        this.toolPaths = source["toolPaths"];
	        this.firstRun = source["firstRun"];
	        this.lastFolder = source["lastFolder"];
	        this.defaultCodeLanguage = source["defaultCodeLanguage"];
	        this.enabledCodeLanguages = source["enabledCodeLanguages"];
	        this.formatOnSave = source["formatOnSave"];
	        this.checkUpdates = source["checkUpdates"];
	        this.lastUpdateCheck = source["lastUpdateCheck"];
	        this.recentFiles = source["recentFiles"];
	        this.inlayHints = source["inlayHints"];
	    }
	}
	export class SignatureHelp {
	    label: string;
	    documentation: string;
	    parameters: string[];
	    activeParameter: number;
	
	    static createFrom(source: any = {}) {
	        return new SignatureHelp(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.documentation = source["documentation"];
	        this.parameters = source["parameters"];
	        this.activeParameter = source["activeParameter"];
	    }
	}
	
	
	
	export class ToolStatus {
	    id: string;
	    codeLanguage: string;
	    role: string;
	    version: string;
	    source: string;
	    path: string;
	    advice: string[];
	
	    static createFrom(source: any = {}) {
	        return new ToolStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.codeLanguage = source["codeLanguage"];
	        this.role = source["role"];
	        this.version = source["version"];
	        this.source = source["source"];
	        this.path = source["path"];
	        this.advice = source["advice"];
	    }
	}
	export class UpdateState {
	    status: string;
	    current: string;
	    latest?: Release;
	    downloadedBytes: number;
	    totalBytes: number;
	    installs: boolean;
	    checkedAt: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.current = source["current"];
	        this.latest = this.convertValues(source["latest"], Release);
	        this.downloadedBytes = source["downloadedBytes"];
	        this.totalBytes = source["totalBytes"];
	        this.installs = source["installs"];
	        this.checkedAt = source["checkedAt"];
	        this.error = source["error"];
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

