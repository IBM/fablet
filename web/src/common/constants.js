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
    // ${document.location.protocol}
    // return `https://${document.location.hostname}:8080${path}`;

    // For prod    
    let port = document.location.port;
    if (port !== "") {
        port = ":" + port;
    }
    return `${document.location.protocol}//${document.location.hostname}${port}${path}`;
}

export const getWebsocketURL = (path) => {
    // For dev service
    // return `wss://${document.location.hostname}:8080${path}`;

    // For prod
    const wsProtocol = document.location.protocol.startsWith("https") ? "wss:" : "ws:";
    let port = document.location.port;
    if (port !== "") {
        port = ":" + port;
    }
    return `${wsProtocol}//${document.location.hostname}${port}${path}`;
}

export const TIMEOUT_NETWORKOVERVIEW_UPDATE = 300000;
export const TIMEOUT_PEERDETAILS_UPDATE = 300000;
export const TIMEOUT_LEDGER_UPDATE = 3000;

export const RES_OK = 200;

export const INITIAL_BLOCKS = 64;