import { clientStorage } from "../data/localstore";
import i18n from '../i18n';
import { log } from "./log";
import requestPromise from "request-promise-native";
import * as CONST from "./constants";

const cntTrim = (str, len) => {
    if (!str) {
        return str;
    }
    return str.length > len ? (str.substring(0, len) + "...") : str;
};
// TODO to use common id/conn object.
const getCurrIDConn = () => {
    const currID = clientStorage.getCurrID();
    const currProfile = clientStorage.getCurrProfile()
    const currIDConn = {
        label: currID.label,
        MSPID: currID.MSPID,
        certContent: currID.certContent,
        prvKeyContent: currID.prvKeyContent,
        connProfile: currProfile.content
    };

    return !currIDConn.MSPID || !currIDConn.certContent || !currIDConn.prvKeyContent || !currIDConn.connProfile
        ? {} : currIDConn;
};

const outputServiceErrorResult = (result) => {
    return i18n("res_code") + ": " + result.resCode + ". " + i18n("err_message") + ": " + result.errMsg;
};

const formatTime = (ts) => {
    const t = new Date(ts);
    return t.getFullYear() + "-" + fillOf2((t.getMonth()+1)) + "-" + fillOf2(t.getDate()) + " " 
        + fillOf2(t.getHours()) + ":" + fillOf2(t.getMinutes()) + ":" + fillOf2(t.getSeconds()) + "." + t.getMilliseconds();
};

const fillOf2 = (v) => {
    return v < 10 ? '0' + v : v;
}


const refreshNetwork = async () => {
    const currIDConn = getCurrIDConn();
    const reqBody = {
        connection: currIDConn
    };

    let myHeaders = new Headers({
        'Accept': 'application/json',
        'Content-Type': 'application/json'
    });

    let option = {
        url: CONST.getServiceURL("/network/refresh"),
        method: 'POST',
        headers: myHeaders,
        json: true,
        resolveWithFullResponse: true,
        body: reqBody
    };

    log.debug("Refresh network.");

    let result = await requestPromise(option);

    if (result) {
        result = result.body;
        if (result.resCode === 200) {
            window.location.reload();
        }
        else {
            // this.setState({ executeError: result.resCode + ". " + result.errMsg });
            log.error("Error: ", result.resCode, result.errMsg);
        }
    }
    else {
        // this.setState({ executeError: i18n("no_result_error") });
        log.error("No result returned.")
    }
}


export { cntTrim, getCurrIDConn, outputServiceErrorResult, formatTime, refreshNetwork };