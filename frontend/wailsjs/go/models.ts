export namespace database {
	
	export class CreateAssetParams {
	    scanFolderId: sql.NullInt64;
	    fileName: string;
	    filePath: string;
	    fileType: string;
	    fileSize: number;
	    thumbnailPath: string;
	    fileHash: sql.NullString;
	    imageWidth: sql.NullInt64;
	    imageHeight: sql.NullInt64;
	    dominantColor: sql.NullString;
	    bitDepth: sql.NullInt64;
	    hasAlphaChannel: sql.NullBool;
	    // Go type: time
	    lastModified: any;
	    // Go type: time
	    lastScanned: any;
	
	    static createFrom(source: any = {}) {
	        return new CreateAssetParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.scanFolderId = this.convertValues(source["scanFolderId"], sql.NullInt64);
	        this.fileName = source["fileName"];
	        this.filePath = source["filePath"];
	        this.fileType = source["fileType"];
	        this.fileSize = source["fileSize"];
	        this.thumbnailPath = source["thumbnailPath"];
	        this.fileHash = this.convertValues(source["fileHash"], sql.NullString);
	        this.imageWidth = this.convertValues(source["imageWidth"], sql.NullInt64);
	        this.imageHeight = this.convertValues(source["imageHeight"], sql.NullInt64);
	        this.dominantColor = this.convertValues(source["dominantColor"], sql.NullString);
	        this.bitDepth = this.convertValues(source["bitDepth"], sql.NullInt64);
	        this.hasAlphaChannel = this.convertValues(source["hasAlphaChannel"], sql.NullBool);
	        this.lastModified = this.convertValues(source["lastModified"], null);
	        this.lastScanned = this.convertValues(source["lastScanned"], null);
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
	export class ScanFolderDTO {
	    id: number;
	    path: string;
	    isActive: boolean;
	    lastScanned?: string;
	    dateAdded: string;
	    isDeleted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ScanFolderDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.isActive = source["isActive"];
	        this.lastScanned = source["lastScanned"];
	        this.dateAdded = source["dateAdded"];
	        this.isDeleted = source["isDeleted"];
	    }
	}
	export class ScanResult {
	    Path: string;
	    Err: any;
	    NewAsset?: database.CreateAssetParams;
	    ExistingPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Err = source["Err"];
	        this.NewAsset = this.convertValues(source["NewAsset"], database.CreateAssetParams);
	        this.ExistingPath = source["ExistingPath"];
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
	
	export class NullBool {
	    Bool: boolean;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NullBool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Bool = source["Bool"];
	        this.Valid = source["Valid"];
	    }
	}
	export class NullInt64 {
	    Int64: number;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NullInt64(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Int64 = source["Int64"];
	        this.Valid = source["Valid"];
	    }
	}
	export class NullString {
	    String: string;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NullString(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.String = source["String"];
	        this.Valid = source["Valid"];
	    }
	}

}

export namespace sync {
	
	export class WaitGroup {
	
	
	    static createFrom(source: any = {}) {
	        return new WaitGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

