export const DEBUG = false;

export const getDebugServiceURL = (path) => `http://${document.location.hostname}:8080${path}`;
export const getDebugWebsocketURL = (path) => `ws://${document.location.hostname}:8080${path}`;