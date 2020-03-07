import * as debugFlag from "./debugflag";

export const DEBUG = debugFlag.DEBUG;

export const getDebugServiceURL = (path) => `http://${document.location.hostname}:8080${path}`;
export const getDebugWebsocketURL = (path) => `ws://${document.location.hostname}:8080${path}`;