export namespace database {
	
	export class ScanFolder {
	    id: number;
	    path: string;
	    isActive: boolean;
	    lastScanned: sql.NullTime;
	    // Go type: time
	    dateAdded: any;
	    isDeleted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ScanFolder(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.isActive = source["isActive"];
	        this.lastScanned = this.convertValues(source["lastScanned"], sql.NullTime);
	        this.dateAdded = this.convertValues(source["dateAdded"], null);
	        this.isDeleted = source["isDeleted"];
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

export namespace services {
	
	export class PaletteColor {
	    name: string;
	    hex: string;
	
	    static createFrom(source: any = {}) {
	        return new PaletteColor(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.hex = source["hex"];
	    }
	}
	export class ScannerConfig {
	    allowedExtensions: string[];
	    maxAllowHashFileSize: number;
	
	    static createFrom(source: any = {}) {
	        return new ScannerConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.allowedExtensions = source["allowedExtensions"];
	        this.maxAllowHashFileSize = source["maxAllowHashFileSize"];
	    }
	}

}

export namespace sql {
	
	export class NullTime {
	    // Go type: time
	    Time: any;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NullTime(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = this.convertValues(source["Time"], null);
	        this.Valid = source["Valid"];
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

