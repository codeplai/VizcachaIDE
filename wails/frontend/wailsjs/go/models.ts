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
	export class GoModule {
	    root: string;
	    modulePath: string;
	
	    static createFrom(source: any = {}) {
	        return new GoModule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.modulePath = source["modulePath"];
	    }
	}
	export class RunConfiguration {
	    target: string;
	    workingDir: string;
	    mode: string;
	    programArgs: string[];
	    module?: GoModule;
	
	    static createFrom(source: any = {}) {
	        return new RunConfiguration(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.target = source["target"];
	        this.workingDir = source["workingDir"];
	        this.mode = source["mode"];
	        this.programArgs = source["programArgs"];
	        this.module = this.convertValues(source["module"], GoModule);
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
	    goPath: string;
	    delvePath: string;
	    goplsPath: string;
	    firstRun: boolean;
	    lastFolder: string;
	    formatOnSave: boolean;
	    recentFiles: string[];
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.theme = source["theme"];
	        this.fontSize = source["fontSize"];
	        this.goPath = source["goPath"];
	        this.delvePath = source["delvePath"];
	        this.goplsPath = source["goplsPath"];
	        this.firstRun = source["firstRun"];
	        this.lastFolder = source["lastFolder"];
	        this.formatOnSave = source["formatOnSave"];
	        this.recentFiles = source["recentFiles"];
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
	
	
	export class ToolchainInfo {
	    goVersion: string;
	    delveVersion: string;
	    goplsVersion: string;
	    goSource: string;
	    delveSource: string;
	    goplsSource: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolchainInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.goVersion = source["goVersion"];
	        this.delveVersion = source["delveVersion"];
	        this.goplsVersion = source["goplsVersion"];
	        this.goSource = source["goSource"];
	        this.delveSource = source["delveSource"];
	        this.goplsSource = source["goplsSource"];
	    }
	}

}

