export namespace app {
	
	export class AssetMaterialSet {
	    id: number;
	    name: string;
	    customColor: string;
	
	    static createFrom(source: any = {}) {
	        return new AssetMaterialSet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.customColor = source["customColor"];
	    }
	}
	export class AssetDetails {
	    id: number;
	    filePath: string;
	    fileName: string;
	    fileType: string;
	    thumbnailPath: string;
	    // Go type: time
	    dateAdded: any;
	    // Go type: time
	    lastModified: any;
	    fileSize: number;
	    imageWidth: number;
	    imageHeight: number;
	    fileExtension: string;
	    rating: number;
	    isFavorite: boolean;
	    description: string;
	    isDeleted: boolean;
	    isHidden: boolean;
	    bitDepth: number;
	    fileHash: string;
	    groupId?: string;
	    tags: string[];
	    materialSets: AssetMaterialSet[];
	    dominantColor: string;
	
	    static createFrom(source: any = {}) {
	        return new AssetDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.filePath = source["filePath"];
	        this.fileName = source["fileName"];
	        this.fileType = source["fileType"];
	        this.thumbnailPath = source["thumbnailPath"];
	        this.dateAdded = this.convertValues(source["dateAdded"], null);
	        this.lastModified = this.convertValues(source["lastModified"], null);
	        this.fileSize = source["fileSize"];
	        this.imageWidth = source["imageWidth"];
	        this.imageHeight = source["imageHeight"];
	        this.fileExtension = source["fileExtension"];
	        this.rating = source["rating"];
	        this.isFavorite = source["isFavorite"];
	        this.description = source["description"];
	        this.isDeleted = source["isDeleted"];
	        this.isHidden = source["isHidden"];
	        this.bitDepth = source["bitDepth"];
	        this.fileHash = source["fileHash"];
	        this.groupId = source["groupId"];
	        this.tags = source["tags"];
	        this.materialSets = this.convertValues(source["materialSets"], AssetMaterialSet);
	        this.dominantColor = source["dominantColor"];
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
	
	export class AssetQueryFilters {
	    page: number;
	    pageSize: number;
	    searchQuery: string;
	    tags: string[];
	    matchAllTags: boolean;
	    fileTypes: string[];
	    colors: string[];
	    ratingRange: number[];
	    widthRange: number[];
	    heightRange: number[];
	    fileSizeRange: number[];
	    // Go type: struct { From *string "json:\"from\""; To *string "json:\"to\"" }
	    dateRange: any;
	    hasAlpha?: boolean;
	    onlyFavorites: boolean;
	    onlyUncategorized: boolean;
	    isDeleted: boolean;
	    isHidden: boolean;
	    collectionId?: number;
	    showRepresentativesOnly: boolean;
	    sortOption: string;
	    sortDesc: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AssetQueryFilters(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.searchQuery = source["searchQuery"];
	        this.tags = source["tags"];
	        this.matchAllTags = source["matchAllTags"];
	        this.fileTypes = source["fileTypes"];
	        this.colors = source["colors"];
	        this.ratingRange = source["ratingRange"];
	        this.widthRange = source["widthRange"];
	        this.heightRange = source["heightRange"];
	        this.fileSizeRange = source["fileSizeRange"];
	        this.dateRange = this.convertValues(source["dateRange"], Object);
	        this.hasAlpha = source["hasAlpha"];
	        this.onlyFavorites = source["onlyFavorites"];
	        this.onlyUncategorized = source["onlyUncategorized"];
	        this.isDeleted = source["isDeleted"];
	        this.isHidden = source["isHidden"];
	        this.collectionId = source["collectionId"];
	        this.showRepresentativesOnly = source["showRepresentativesOnly"];
	        this.sortOption = source["sortOption"];
	        this.sortDesc = source["sortDesc"];
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
	export class CreateMaterialSetRequest {
	    name: string;
	    description?: string;
	    coverAssetId?: number;
	    customCoverUrl?: string;
	    customColor?: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateMaterialSetRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.coverAssetId = source["coverAssetId"];
	        this.customCoverUrl = source["customCoverUrl"];
	        this.customColor = source["customColor"];
	    }
	}
	export class LibraryStats {
	    totalAssets: number;
	    totalSize: number;
	    // Go type: time
	    lastScan?: any;
	
	    static createFrom(source: any = {}) {
	        return new LibraryStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalAssets = source["totalAssets"];
	        this.totalSize = source["totalSize"];
	        this.lastScan = this.convertValues(source["lastScan"], null);
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
	export class MaterialSet {
	    id: number;
	    name: string;
	    description?: string;
	    coverAssetId?: number;
	    customCoverUrl?: string;
	    customColor?: string;
	    thumbnailPath: string;
	    // Go type: time
	    dateAdded: any;
	    // Go type: time
	    lastModified: any;
	    totalAssets: number;
	
	    static createFrom(source: any = {}) {
	        return new MaterialSet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.coverAssetId = source["coverAssetId"];
	        this.customCoverUrl = source["customCoverUrl"];
	        this.customColor = source["customColor"];
	        this.thumbnailPath = source["thumbnailPath"];
	        this.dateAdded = this.convertValues(source["dateAdded"], null);
	        this.lastModified = this.convertValues(source["lastModified"], null);
	        this.totalAssets = source["totalAssets"];
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
	export class PagedAssetResult {
	    items: AssetDetails[];
	    totalCount: number;
	    page: number;
	    pageSize: number;
	
	    static createFrom(source: any = {}) {
	        return new PagedAssetResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], AssetDetails);
	        this.totalCount = source["totalCount"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
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
	export class SidebarStats {
	    totalAssets: number;
	    totalUncategorized: number;
	    totalFavorites: number;
	    totalTrash: number;
	    totalHidden: number;
	
	    static createFrom(source: any = {}) {
	        return new SidebarStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalAssets = source["totalAssets"];
	        this.totalUncategorized = source["totalUncategorized"];
	        this.totalFavorites = source["totalFavorites"];
	        this.totalTrash = source["totalTrash"];
	        this.totalHidden = source["totalHidden"];
	    }
	}
	export class Tag {
	    id: number;
	    name: string;
	    assetCount: number;
	
	    static createFrom(source: any = {}) {
	        return new Tag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.assetCount = source["assetCount"];
	    }
	}
	export class UpdateAssetRequest {
	    Description?: string;
	    Rating?: number;
	    IsFavorite?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpdateAssetRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Description = source["Description"];
	        this.Rating = source["Rating"];
	        this.IsFavorite = source["IsFavorite"];
	    }
	}

}

export namespace config {
	
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

}

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
	    groupId: string;
	
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
	        this.groupId = source["groupId"];
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
	export class UpdateAssetFromScanParams {
	    filePath: sql.NullString;
	    scanFolderId: sql.NullInt64;
	    isDeleted: sql.NullBool;
	    fileSize: sql.NullInt64;
	    fileHash: sql.NullString;
	    lastModified: sql.NullTime;
	    lastScanned: sql.NullTime;
	    thumbnailPath: sql.NullString;
	    imageWidth: sql.NullInt64;
	    imageHeight: sql.NullInt64;
	    dominantColor: sql.NullString;
	    bitDepth: sql.NullInt64;
	    hasAlphaChannel: sql.NullBool;
	    id: number;
	
	    static createFrom(source: any = {}) {
	        return new UpdateAssetFromScanParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filePath = this.convertValues(source["filePath"], sql.NullString);
	        this.scanFolderId = this.convertValues(source["scanFolderId"], sql.NullInt64);
	        this.isDeleted = this.convertValues(source["isDeleted"], sql.NullBool);
	        this.fileSize = this.convertValues(source["fileSize"], sql.NullInt64);
	        this.fileHash = this.convertValues(source["fileHash"], sql.NullString);
	        this.lastModified = this.convertValues(source["lastModified"], sql.NullTime);
	        this.lastScanned = this.convertValues(source["lastScanned"], sql.NullTime);
	        this.thumbnailPath = this.convertValues(source["thumbnailPath"], sql.NullString);
	        this.imageWidth = this.convertValues(source["imageWidth"], sql.NullInt64);
	        this.imageHeight = this.convertValues(source["imageHeight"], sql.NullInt64);
	        this.dominantColor = this.convertValues(source["dominantColor"], sql.NullString);
	        this.bitDepth = this.convertValues(source["bitDepth"], sql.NullInt64);
	        this.hasAlphaChannel = this.convertValues(source["hasAlphaChannel"], sql.NullBool);
	        this.id = source["id"];
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

export namespace scanner {
	
	export class ScanResult {
	    Path: string;
	    Err: any;
	    NewAsset?: database.CreateAssetParams;
	    ModifiedAsset?: database.UpdateAssetFromScanParams;
	    ExistingPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Err = source["Err"];
	        this.NewAsset = this.convertValues(source["NewAsset"], database.CreateAssetParams);
	        this.ModifiedAsset = this.convertValues(source["ModifiedAsset"], database.UpdateAssetFromScanParams);
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
	export class ScannerConfigSnapshot {
	    allowedExtensions: string[];
	    maxAllowHashFileSize: number;
	
	    static createFrom(source: any = {}) {
	        return new ScannerConfigSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.allowedExtensions = source["allowedExtensions"];
	        this.maxAllowHashFileSize = source["maxAllowHashFileSize"];
	    }
	}

}

export namespace settings {
	
	export class AppConfigDTO {
	    allowedExtensions: string[];
	    maxAllowHashFileSize: number;
	    debugMode: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.allowedExtensions = source["allowedExtensions"];
	        this.maxAllowHashFileSize = source["maxAllowHashFileSize"];
	        this.debugMode = source["debugMode"];
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

export namespace update {
	
	export class ReleaseNote {
	    tagName: string;
	    body: string;
	
	    static createFrom(source: any = {}) {
	        return new ReleaseNote(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tagName = source["tagName"];
	        this.body = source["body"];
	    }
	}
	export class ReleaseInfo {
	    tagName: string;
	    body: string;
	    downloadUrl: string;
	    isUpdateAvailable: boolean;
	    history: ReleaseNote[];
	
	    static createFrom(source: any = {}) {
	        return new ReleaseInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tagName = source["tagName"];
	        this.body = source["body"];
	        this.downloadUrl = source["downloadUrl"];
	        this.isUpdateAvailable = source["isUpdateAvailable"];
	        this.history = this.convertValues(source["history"], ReleaseNote);
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

