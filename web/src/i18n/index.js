import EN_TEXT from "./en";
import CN_TEXT from "./cn";
import * as CONST from "../common/constants";

function getTextProvider(loc) {
    switch (loc) {
        case CONST.__LOC_EN__: return EN_TEXT;
        case CONST.__LOC_CN__: return CN_TEXT;
        default: return EN_TEXT;
    }
}

function textWithLoc(loc, key, ...params) {
    let txt = getTextProvider(loc)[key] || key;
    if (params && params.length>0) {
        for (var idx in params) {
            txt = txt.replace("%v", params[idx]);
        }
    }
    return txt;
}

const i18n = function(key, ...params) {
    // TODO Local to be persisted
    return textWithLoc(CONST.__LOC_EN__, key, ...params);
}

export default i18n;