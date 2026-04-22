export namespace ai {
	
	export class ExplanationResponse {
	    originalText: string;
	    explanation: string;
	    tone: string;
	    examples: string[];
	
	    static createFrom(source: any = {}) {
	        return new ExplanationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.originalText = source["originalText"];
	        this.explanation = source["explanation"];
	        this.tone = source["tone"];
	        this.examples = source["examples"];
	    }
	}

}

