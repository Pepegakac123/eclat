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

