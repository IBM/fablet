import * as debugFlag from "./debugflag";

// TODO To use build arguments.
export const DEBUG = debugFlag.DEBUG;

export const getDebugServiceURL = (path) => `http://${document.location.hostname}:8080${path}`;
export const getDebugWebsocketURL = (path) => `ws://${document.location.hostname}:8080${path}`;