export const __VER_0_1_0__ = "0.1.0";
export const __CODE_MERCURY__ = "mercury";

export const __VER__ = "ver";
export const __CODE__ = "code";

export const __LOC_EN__ = "en";
export const __LOC_CN__ = "cn";

export const __DATA_VER__ = "data_version";

export const getCurrVer = () => {
    return {
        __VER__: __VER_0_1_0__,
        __CODE__: __CODE_MERCURY__
    }
}

export const getServiceURL = (path) => {
    // For dev service
    return `${document.location.protocol}//${document.location.hostname}:8080${path}`;

    // TODO
    // For prod
    // return path;
}

export const getWebsocketURL = (path) => {
    // For dev service
    return `ws://${document.location.hostname}:8080${path}`;

    // TODO
    // For prod
    // return path;
}

export const TIMEOUT_NETWORKOVERVIEW_UPDATE = 300000;
export const TIMEOUT_PEERDETAILS_UPDATE = 300000;
export const TIMEOUT_LEDGER_UPDATE = 3000;

export const RES_OK = 200;

export const INITIAL_BLOCKS = 64;