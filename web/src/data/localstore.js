import { log } from "../common/log";

const __KEY_CONN_PROFS__ = "conn_profs";
const __KEY_CURR_CONN_PROF__ = "curr_conn_prof";
const __KEY_IDENTITIES__ = "identities";
const __KEY_CURR_ID__ = "curr_id";

// TODO TO check if localstorage is supported
// TODO to cache the connection profile and identity.
class ConnectionProfile {
    constructor(options){
        log.log("Construct ConnectionProfile");
        this.name = options.name;
        this.path = options.path;
        this.content = options.content;
    }
    isValid() {
        return this.name && this.path && this.content;
    }
}

const parseJSON = (str) => {
    try {
        return JSON.parse(str);
    }
    catch (e) {
        return [];
    }
}

class Identity {
    constructor(options){
        console.log("Construct ConnectionProfile");
        this.label = options.label;
        this.MSPID = options.MSPID;
        this.certPath = options.certPath;
        this.certContent = options.certContent;
        this.prvKeyPath = options.prvKeyPath;
        this.prvKeyContent = options.prvKeyContent;
    }
    isValid() {
        return this.label && this.MSPID && this.certContent && this.prvKeyContent;
    }
}

class ClientStorage {
    constructor() {
        log.debug("Construct ClientStorage...............................................");
        this.profiles = [];
        this.identities = [];
        this.currConnProfileName = "";
        this.currIDLabel = "";
        
        this.init();
    }
    init() {
        log.log("Init connection profiles.");
        this.profiles = function() {
            const profiles = parseJSON(localStorage.getItem(__KEY_CONN_PROFS__));
            if (profiles && profiles instanceof Array) {
                return profiles
                .map(p => new ConnectionProfile(p))
                .filter(p => p.isValid());
            }
            return [];
        }();
        this.identities = function(){
            const ids = parseJSON(localStorage.getItem(__KEY_IDENTITIES__));
            if (ids && ids instanceof Array) {
                return ids
                .map(i => new Identity(i))
                .filter(i => i.isValid());
            }
            return [];
        }();
        this.currIDLabel = localStorage.getItem(__KEY_CURR_ID__) || "";
        this.currConnProfileName = localStorage.getItem(__KEY_CURR_CONN_PROF__) || "";

        let needRefresh = false;
        if (!this.getIDByName(this.identities, this.currIDLabel)){
            this.setCurrIDLabel(this.identities.length > 0 ? this.identities[0].label: "")
            needRefresh = this.identities.length > 0;
        }
        if (!this.getProfileByName(this.profiles, this.currConnProfileName)) {
            this.setCurrConnProfileName(this.profiles.length > 0 ? this.profiles[0].name : "");
            needRefresh = needRefresh || this.profiles.length > 0;
        }
        if (needRefresh) {
            window.location.reload();
        }
    }

    loadConnectionProfiles() {
        return this.profiles;
    }

    saveConnectionProfile(options) {
        const temp = new ConnectionProfile(options);
        if (!temp.isValid()) {
            return;
        }

        const origName = temp.name;
        const updatedConnProfs = this.loadConnectionProfiles();

        // At most 1024 duplicates.
        for (let i=1; i<1025; i++) {
            if (this.getProfileByName(updatedConnProfs, temp.name)) {
                temp.name = origName + `(${i})`;
            }
            else {
                break;
            }
        }

        updatedConnProfs.push(temp);
        localStorage.setItem(__KEY_CONN_PROFS__, JSON.stringify(updatedConnProfs));

        log.debug("saveConnectionProfile");
    }

    setCurrConnProfileName (currName) {
        localStorage.setItem(__KEY_CURR_CONN_PROF__, currName);
        this.currConnProfileName = currName;
    }

    getCurrConnProfileName() {
        return this.currConnProfileName;
    }

    getCurrProfile() {
        return this.getProfileByName(this.loadConnectionProfiles(), this.getCurrConnProfileName());
    }

    getProfileByName(profiles, name) {
        for (let i=0; profiles && i<profiles.length; i++) {
            if (profiles[i].name === name) {
                return profiles[i];
            }
        }
        return null;
    }

    removeConnectionProfile(name) {
        const updatedConnProfs = this.loadConnectionProfiles().filter(p => p.name !== name);
        localStorage.setItem(__KEY_CONN_PROFS__, JSON.stringify(updatedConnProfs));
        log.debug("removeConnectionProfile");
    }

    loadIdentities() {
        return this.identities;
    }
    
    saveIdentity(options) {
        const temp = new Identity(options);
        if (!temp.isValid()) {
            return;
        }    
        const updatedIdentities = this.loadIdentities();
        const origLabel = temp.label;

        // At most 1024 duplicates.
        for (let i=1; i<1025; i++) {
            if (this.getIDByName(updatedIdentities, temp.label)) {
                temp.label = origLabel + `(${i})`;
            }
            else {
                break;
            }
        }

        updatedIdentities.push(temp);
        localStorage.setItem(__KEY_IDENTITIES__, JSON.stringify(updatedIdentities));
    
        log.debug("saveIdentity");
    }
    
    getCurrIDLabel() {
        return this.currIDLabel;
    }
    
    setCurrIDLabel(currIDLabel) {
        localStorage.setItem(__KEY_CURR_ID__, currIDLabel);
        this.currIDLabel = currIDLabel;
    }
    
    getCurrID() {
        return this.getIDByName(this.loadIdentities(), this.getCurrIDLabel());
    };

    getIDByName(ids, label) {
        for (let i=0; ids && i<ids.length; i++) {
            if (ids[i].label === label) {
                return ids[i];
            }
        }
        return null;
    }
    
    removeIdentity(label) {
        const updatedIdentities = this.loadIdentities().filter(id => id.label !== label);
        localStorage.setItem(__KEY_IDENTITIES__, JSON.stringify(updatedIdentities));
        log.debug("removeIdentity");
    }

}

const clientStorage = new ClientStorage();

export { ConnectionProfile, clientStorage};