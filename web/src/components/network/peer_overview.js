import React from 'react';
import Typography from '@material-ui/core/Typography';
import { withStyles } from '@material-ui/core/styles';
import Card from '@material-ui/core/Card';
import Tooltip from '@material-ui/core/Tooltip';
import Button from '@material-ui/core/Button';
import IconButton from '@material-ui/core/IconButton';
import Grid from '@material-ui/core/Grid';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import CardActions from '@material-ui/core/CardActions';
import CardContent from '@material-ui/core/CardContent';
import TocIcon from '@material-ui/icons/Toc';
import CloseIcon from '@material-ui/icons/Close';
import ViewHeadlineIcon from '@material-ui/icons/ViewHeadline';
import Chip from '@material-ui/core/Chip';
import i18n from '../../i18n';
import ChaincodeInstall from "../chaincode/install";
import ChaincodeInstantiate from "../chaincode/instantiate";
import ChaincodeExecute from "../chaincode/execute";
import ChannelCreate from "../channel/create";
import ChannelJoin from "../channel/join";
import LedgerDetail from "../ledger/detail";
import { cntTrim, getCurrIDConn } from "../../common/utils";
import { log } from "../../common/log";
import * as CONST from "../../common/constants";

const styles = (theme) => ({
    card: {
        width: "100%",
    },
    highLightCard: {
        width: "100%",
        backgroundColor: "#cfe8fc"
    },
    cardContent: {
        //display: 'flex',
        //flexWrap: 'wrap',
        paddingLeft: 12,
        paddingTop: 6,
        paddingBottom: 6,
        paddingRight: 6,
    },
    bullet: {
        display: 'inline-block',
        margin: '0 2px',
        transform: 'scale(0.8)',
    },
    title: {
        fontSize: 13,
        fontWeight: 600
    },
    normal: {
        fontSize: 13,
    },
    running: {
        background: "#12C412",
    },
    timeFlag: {
        fontSize: 11,
        marginLeft: "auto",
    },
    fieldDetail: {
        fontSize: 11,
    },
    tooltipTableWrapper: {
        maxHeight: 450,
        overflow: 'auto',
    },
    detailTableWrapper: {
        maxHeight: 450,
        overflow: 'auto',
    },
});

const HtmlTooltip = withStyles(theme => ({
    tooltip: {
        backgroundColor: "#f5f5f9",
        color: "rgba(0, 0, 0, 0.87)",
        maxWidth: 650,
        width: 650,
        fontSize: theme.typography.pxToRem(12),
        border: "1px solid #dadde9"
    },
}))(Tooltip);

class PeerOverview extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            ...props,
            // peer, peers, channelLedgers, channelOrderers, channelChaincodes
            // onFocus: props.onFocus
            // onLeaveFocus: props.onLeaveFocus
            // isFocus: props.isFocus

            openInstallChaincode: false,
            openInstantiateChaincode: false,
            openExecuteChaincode: false,
            openLedgerDetail: false,
            openCreateChannel: false,
            openJoinChannel: false,
            ccInstantiateOption: {},
            ccInstallOption: {},
            ccExecuteOption: {},
            ledgerQueryOption: {},
            channelCreateOption: {},
            channelJoinOption: {},

            // Initial chaincodes, all from channel chaincodes.
            mergedCCs: this.mergeCCs([], props.channelChaincodes),
            // Means installed chaincodes query error, or, it is the initial state, not sure of installed chaincodes.
            installedCCQueryError: true
        };

        this.isRunning = true;

        if (!this.state.peer) {
            this.state.peer = {};
        }

        this.handleChaincodeInstall = this.handleChaincodeInstall.bind(this);
        this.handleChaincodeInstantiate = this.handleChaincodeInstantiate.bind(this);
        this.handleChaincodeExecute = this.handleChaincodeExecute.bind(this);
        this.handleLedgerQuery = this.handleLedgerQuery.bind(this);
        this.handleChannelCreate = this.handleChannelCreate.bind(this);
        this.handleChannelJoin = this.handleChannelJoin.bind(this);

        this.handleCloseChaincodeInstall = this.handleCloseChaincodeInstall.bind(this);
        this.handleCloseChaincodeInstantiate = this.handleCloseChaincodeInstantiate.bind(this);
        this.handleCloseChaincodeExecute = this.handleCloseChaincodeExecute.bind(this);
        this.handleCloseLedgerQuery = this.handleCloseLedgerQuery.bind(this);
        this.handleCloseChannelCreate = this.handleCloseChannelCreate.bind(this);
        this.handleCloseChannelJoin = this.handleCloseChannelJoin.bind(this);


        this.discoverUpdate = this.discoverUpdate.bind(this);

        log.debug("PeerOverview: constructor");
    }

    discoverUpdate() {
        const comp = this;

        // Remove the current timeout to avoid any duplicate invoking.
        // clearTimeout(this.updateTO);
        if (!comp.isRunning) {
            return;
        }

        const currIDConn = getCurrIDConn();

        const reqBody = {
            connection: currIDConn,
            target: comp.state.peer.URL,
            channels: comp.state.peer.channels || []
        };

        let myHeaders = new Headers({
            'Accept': 'application/json',
            'Content-Type': 'application/json'
        });
        let request = new Request(CONST.getServiceURL("/peer/details"), {
            method: 'POST',
            //mode: 'cors',
            //credentials: 'include',
            headers: myHeaders,
            body: JSON.stringify(reqBody),
        });

        log.debug("Discover peer:", comp.state.peer.URL);

        fetch(request)
            .then(response => response.json())
            .then(result => {
                log.debug("Peer details result", result);
                if (comp.isRunning) {
                    if (result) {
                        log.debug("Discover peer:", comp.state.peer.URL, ", found installed ", (result.installedChaincodes || []).length, "chaincodes.");
                        const mergedCCs = comp.mergeCCs(result.installedChaincodes, result.channelChaincodes);
                        comp.setState({
                            mergedCCs: mergedCCs,
                            installedCCQueryError: result.installedCCQueryError,
                            channelLedgers: result.channelLedgers,
                            peer: { ...comp.state.peer, channels: result.channels }
                        });
                    }
                    // TODO timeout
                    comp.updateTO = setTimeout(comp.discoverUpdate, 30000);
                }
            });
    }

    componentDidMount() {
        log.debug("PeerOverview: componentDidMount");
        if (this.state.peer) {
            // TODO stop
            this.discoverUpdate();
        }
    }

    componentDidUpdate() {
    }

    componentWillUnmount() {
        log.debug("Clear timeout task of the peer overview update");
        //clearTimeout(this.updateTO);  // To avoid memory leak.
        this.isRunning = false;
        clearTimeout(this.updateTO);
    }

    handleChannelJoin(peer) {
        const comp = this;
        return function () {
            comp.setState({ openJoinChannel: true });
            comp.setState({
                channelJoinOption: {
                    peer: peer
                }
            });
        };
    }

    handleChannelCreate() {
        const comp = this;
        return function () {
            comp.setState({ openCreateChannel: true });
            comp.setState({
                channelCreateOption: {}
            });
        };
    }


    handleChaincodeInstall(peer, name, version, type, path) {
        const comp = this;
        return function () {
            comp.setState({ openInstallChaincode: true });

            comp.setState({
                ccInstallOption: {
                    name: name,
                    version: version,
                    type: type,
                    path: path,
                    target: peer
                }
            });
        };
    }

    handleChaincodeInstantiate(existInstantiatedCC, name, version, path, policy, constructor, channelID, peer, orderer) {
        const comp = this;
        return function () {
            comp.setState({ openInstantiateChaincode: true });

            // TODO HACK MOCKING
            // comp.state.channelOrderers["mychannel"][0].name="orderer2.example.com:8050";
            // comp.state.channelOrderers["mychannel"][0].URL="orderer2.example.com:8050";

            comp.setState({
                ccInstantiateOption: {
                    // Predefined
                    existInstantiatedCC,
                    name: name,
                    version: version,
                    path: path,
                    policy: policy,
                    constructor: constructor,
                    channelID: channelID,
                    peer: peer,
                    orderer: orderer,
                    // Options
                    channelOrderers: comp.state.channelOrderers,
                }
            });
        }
    }

    getChannelPeers(channelID) {
        const peers = [];
        (this.state.peers || []).forEach(peer => {
            if ((peer.channels || []).includes(channelID)) {
                peers.push(peer);
            }
        });
        return peers;
    }

    handleChaincodeExecute(actionType, channelID, name, peer) {
        const comp = this;
        return function () {
            comp.setState({ openExecuteChaincode: true });

            comp.setState({
                ccExecuteOption: {
                    actionType: actionType,
                    channelID: channelID,
                    name: name,
                    peer: peer,
                    targets: comp.getChannelPeers(channelID)
                }
            });
        };
    }

    handleLedgerQuery(channelID, target) {
        const comp = this;
        return function () {
            comp.setState({
                openLedgerDetail: true,
                ledgerQueryOption: {
                    channelID: channelID,
                    target: target
                }
            });
        };
    }

    handleCloseChaincodeInstall() {
        this.setState({ openInstallChaincode: false });
    }

    handleCloseChaincodeInstantiate() {
        this.setState({ openInstantiateChaincode: false });
    }

    handleCloseChaincodeExecute() {
        this.setState({ openExecuteChaincode: false });
    }

    handleCloseLedgerQuery() {
        this.setState({ openLedgerDetail: false });
    }

    handleCloseChannelJoin() {
        this.setState({ openJoinChannel: false });
    }

    handleCloseChannelCreate() {
        this.setState({ openCreateChannel: false });
    }

    timeLast(t) {
        // const last = ((new Date()).getTime() - t) / 1000; // Second
        // if (last < 0) {
        //     return i18n("time_last_update_invalid");
        // }
        // const hour = last / 3600;
        // const minute = (last % 3600) / 60;
        // const second = last % 60;
        const last = new Date(t);
        const hour = last.getHours();
        const minute = last.getMinutes();
        const second = last.getSeconds();
        let timeRes = `${hour}:${minute}:${second}`;
        return timeRes;
    }

    mergeCCs(installedCCs, channelChaincodes) {
        const instCCs = installedCCs || [];
        const chanCCs = channelChaincodes || [];
        var ccs = [];

        for (var channelID in chanCCs) {
            for (var i in (chanCCs[channelID] || [])) {
                const instantiatedCC = chanCCs[channelID][i];
                const idx = this.findCCByName(instCCs, instantiatedCC);
                if (idx < 0) {
                    instantiatedCC.installed = false;
                    ccs.push(instantiatedCC);   // To install, 
                }
                else {
                    instantiatedCC.installed = true;
                    ccs.push(instantiatedCC);   // To query, execute

                    const installedCC = instCCs[idx];
                    instCCs[idx] = undefined;   // Remove what found
                    if (instantiatedCC.version !== installedCC.version) {
                        installedCC.instantiatedCC = instantiatedCC;
                        ccs.push(installedCC);  // To upgrade
                    }
                }
            }

        }
        // Add all installed remained.
        instCCs.forEach(cc => {
            if (cc) {
                const idx = this.findCCByName(ccs, cc);
                if (idx >= 0 && ccs[idx].channelID) {
                    cc.instantiatedCC = ccs[idx];
                }
                ccs.push(cc);
            }
        });

        ccs = ccs.sort(function (a, b) {
            if (a.channelID !== b.channelID) {
                if (!a.channelID) {
                    return 1;
                }
                if (!b.channelID) {
                    return -1;
                }
                return a.channelID.localeCompare(b.channelID);
            }
            if (a.name !== b.name) {
                return a.name.localeCompare(b.name);
            }
            return a.version.localeCompare(b.version);
        });

        return ccs;
    }

    // flatChannelChaincodes(ccs) {
    //     const flatCCs = [];
    //     for (var channelID in ccs) {
    //         for (var idx in ccs[channelID]) {
    //             flatCCs.push(ccs[channelID][idx]);
    //         }
    //     }
    //     return flatCCs;
    // }

    findCCByName(ccList, cc) {
        if (!ccList || !cc) {
            return -1;
        }
        for (var idx in ccList) {
            if (ccList[idx] && ccList[idx].name === cc.name) {
                return idx;
            }
        }
        return -1;
    }

    getLedgerAction(channelID, target) {
        return (
            <Button size="small" color="primary" style={{ marginLeft: "auto", fontSize: 11 }}
                onClick={this.handleLedgerQuery(channelID, target)}
            >
                {i18n("ledger_query")}
            </Button>);
    }

    getChaincodeAction(cc, peer) {
        var executeAction = null;
        var installAction = null;
        var instantiateAction = null;

        if (cc.channelID) {
            executeAction = (
                <React.Fragment>
                    <Button size="small" color="primary" style={{ marginLeft: "auto", fontSize: 11 }}
                        onClick={this.handleChaincodeExecute("query", cc.channelID, cc.name, peer)}>
                        {i18n("query")}
                    </Button>
                    <Button size="small" color="primary" style={{ marginLeft: "auto", fontSize: 11 }}
                        onClick={this.handleChaincodeExecute("execute", cc.channelID, cc.name, peer)}>
                        {i18n("execute")}
                    </Button>
                </React.Fragment>);
        }

        // !installedCCQueryError means surely !installed
        if (!this.state.installedCCQueryError && !cc.installed) {
            installAction = (
                //name, version, path, policy, constructor, channelID, targetEndpoint, ordererEndpoint,
                //channelIDList, ordererEndpointList
                <Button size="small" color="primary" style={{ marginLeft: "auto", fontSize: 11 }}
                    onClick={this.handleChaincodeInstall(this.state.peer, cc.name, cc.version, cc.type, cc.path)}
                    value={cc.name + ":" + cc.version}>{i18n("install")}
                </Button>);
        }

        if (cc.installed && !cc.channelID) {
            instantiateAction = (
                <Button size="small" color="primary" style={{ marginLeft: "auto", fontSize: 11 }}
                    onClick={this.handleChaincodeInstantiate(cc.instantiatedCC, cc.name, cc.version, cc.path, "", "", "",
                        this.state.peer, "")}
                    value={cc.name + ":" + cc.version}>{i18n(cc.instantiatedCC ? "upgrade" : "instantiate")}
                </Button>
            );

        }

        return (
            <React.Fragment>
                {executeAction}
                {installAction}
                {instantiateAction}
            </React.Fragment>
        );
    }

    getChaincodeTable(mergedCCs, peer) {
        const classes = this.props.classes;
        return (<Table stickyHeader className={classes.table} size="small" aria-label="chaincodes">
            <TableHead>
                <TableRow>
                    <TableCell>{i18n("chaincode")}</TableCell>
                    <TableCell align="right">{i18n("version")}</TableCell>
                    <TableCell align="right">{i18n("channel")}</TableCell>
                    <TableCell align="right">{i18n("chaincode_operation")}</TableCell>
                </TableRow>
            </TableHead>

            {/* TODO there is a key duplication issue when mouse hover. */}
            <TableBody key={"tbody_cc_" + peer.name}>
                {mergedCCs.map((cc) => {
                    return (
                        <TableRow key={"instantiate_" + cc.name + cc.version + cc.channelID + peer.name}>
                            <TableCell component="th" scope="row"><Typography noWrap className={classes.fieldDetail}>{cc.name}</Typography></TableCell>
                            <TableCell align="right"><Typography noWrap className={classes.fieldDetail}>{cc.version}</Typography></TableCell>
                            <TableCell align="right"><Typography noWrap className={classes.fieldDetail}>{cc.channelID}</Typography></TableCell>
                            <TableCell align="right">{this.getChaincodeAction(cc, peer)}</TableCell>
                        </TableRow>
                    );
                })}
            </TableBody>
        </Table>);
    }

    getChannelTable(peer) {
        const classes = this.props.classes;
        const channelLedgers = this.state.channelLedgers || {};
        const peerChannels = peer.channels.sort(function (a, b) {
            return a.localeCompare(b);
        });

        return (<Table stickyHeader className={classes.table} size="small" aria-label="channels">
            <TableHead>
                <TableRow>
                    <TableCell>{i18n("channel")}</TableCell>
                    <TableCell>{i18n("ledger_height")}</TableCell>
                    <TableCell>{i18n("ledger_currentblockhash")}</TableCell>
                    <TableCell align="right">{i18n("channel_operation")}</TableCell>
                </TableRow>
            </TableHead>

            {/* TODO there is a key duplication issue when mouse hover. */}
            <TableBody key={"tbody_ch_" + peer.name}>
                {peerChannels.map((channel) => {
                    const ldg = channelLedgers[channel] || {};
                    return (
                        <TableRow key={"channel_" + channel + peer.name}>
                            <TableCell component="th" scope="row"><Typography noWrap className={classes.fieldDetail}>{channel}</Typography></TableCell>
                            <TableCell component="th" scope="row"><Typography noWrap className={classes.fieldDetail}>{ldg.height}</Typography></TableCell>
                            <TableCell component="th" scope="row">
                                <Tooltip interactive title={ldg.currentBlockHash || ""}>
                                    <Typography noWrap className={classes.fieldDetail}>
                                        {cntTrim(ldg.currentBlockHash, 20)}
                                    </Typography>
                                </Tooltip>
                            </TableCell>
                            <TableCell align="right">
                                {ldg.currentBlockHash ? this.getLedgerAction(channel, peer.URL) : null}
                            </TableCell>
                        </TableRow>
                    );
                })}
            </TableBody>
        </Table>);
    }

    render() {
        const classes = this.props.classes;

        if (!this.state.peer) {
            return ("Error! No peer result.");
        }
        const peer = this.state.peer;

        var ccs = {};
        if (this.state.mergedCCs) {
            this.state.mergedCCs.forEach(cc => {
                ccs[cc.name] = "";
            });
        }
        var ccsStr = Object.keys(ccs).join(", ") || i18n("chaincode_noany");

        var chs = {};
        if (peer.channels) {
            peer.channels.forEach(ch => {
                chs[ch] = "";
            });
        }

        var channels = Object.keys(chs);

        channels = channels.sort(function (a, b) {
            return a.localeCompare(b);
        });

        var chsStr = channels.join(", ") || i18n("channel_noany");

        return (
            <React.Fragment>
                <Card> {/* className={peer.isConnected ? classes.highLightCard : classes.card} */}
                    <CardActions>
                        <Chip variant="outlined" color="primary" label="Peer" size="small" />
                        <Chip variant="outlined" label={cntTrim(peer.MSPID, 20)} size="small" />

                        {!this.state.isFocus ?
                            (<Tooltip title={i18n("show_details")}>
                                <IconButton aria-label={i18n("show_details")} size="small" style={{ marginLeft: "auto" }} onClick={this.state.onFocus}>
                                    <TocIcon fontSize="inherit" color="primary" />
                                </IconButton>
                            </Tooltip>)
                            :
                            (<Tooltip title={i18n("close_details")}>
                                <IconButton aria-label={i18n("close_details")} size="small" style={{ marginLeft: "auto" }} onClick={this.state.onLeaveFocus}>
                                    <CloseIcon fontSize="inherit" color="primary" />
                                </IconButton>
                            </Tooltip>)}
                    </CardActions>

                    <CardContent className={classes.cardContent}>
                        <Typography className={classes.title} noWrap>{peer.name}</Typography>
                        <Typography className={classes.normal} noWrap>{peer.URL}</Typography>

                        <Typography display="inline" className={classes.title}>{i18n("channels")}</Typography>

                        <Button size="small" color="primary" style={{ marginLeft: "auto" }} onClick={this.handleChannelCreate()}>+ {i18n("create")}</Button>
                        <Button size="small" color="primary" style={{ marginLeft: "auto" }} onClick={this.handleChannelJoin(peer)}>+ {i18n("join")}</Button>

                        {this.state.isFocus ?
                            (
                                <div className={classes.detailTableWrapper}>
                                    {this.getChannelTable(peer)}
                                </div>
                            )
                            :
                            (
                                <Grid container >
                                    <Grid item xs={11}>
                                        <Typography noWrap className={classes.normal}>{chsStr}</Typography>
                                    </Grid>
                                    <Grid item xs={1}>
                                        <HtmlTooltip interactive title={
                                            (peer.channels && peer.channels.length > 0 ?
                                                (
                                                    <div className={classes.tooltipTableWrapper}>
                                                        {this.getChannelTable(peer)}
                                                    </div>
                                                ) :
                                                (<Typography className={classes.normal} key={"no_channel"}>
                                                    {chsStr}
                                                </Typography>))
                                        }>
                                            <ViewHeadlineIcon style={{ marginLeft: "auto" }} fontSize="inherit" color="primary" />
                                        </HtmlTooltip>
                                    </Grid>
                                </Grid>)}


                        <Typography display="inline" className={classes.title}>{i18n("chaincodes")}</Typography>
                        <Button size="small" color="primary" style={{ marginLeft: "auto" }} onClick={this.handleChaincodeInstall(this.state.peer)}>+ {i18n("install")}</Button>

                        {this.state.isFocus ?
                            (
                                <div className={classes.detailTableWrapper}>
                                    {this.getChaincodeTable(this.state.mergedCCs, peer)}
                                </div>
                            )
                            :
                            (
                                <Grid container >
                                    <Grid item xs={11}>
                                        <Typography noWrap className={classes.normal}>{ccsStr}</Typography>
                                    </Grid>
                                    <Grid item xs={1}>
                                        <HtmlTooltip interactive title={
                                            (this.state.mergedCCs && this.state.mergedCCs.length > 0 ?
                                                (
                                                    <div className={classes.tooltipTableWrapper}>
                                                        {this.getChaincodeTable(this.state.mergedCCs, peer)}
                                                    </div>
                                                ) :
                                                (<Typography className={classes.normal} key={"no_channel"}>
                                                    {ccsStr}
                                                </Typography>))
                                        }>
                                            <ViewHeadlineIcon style={{ marginLeft: "auto" }} fontSize="inherit" color="primary" />
                                        </HtmlTooltip>
                                    </Grid>
                                </Grid>)}

                    </CardContent>
                    <CardActions>
                        <Chip color="primary" size="small" label="ping" className={classes.running} />
                        <Chip color="primary" size="small" label="grpc" className={classes.running} />
                        <Typography className={classes.timeFlag} style={{ marginLeft: "auto" }}>{this.timeLast(peer.updateTime)}</Typography>
                    </CardActions>
                </Card>

                {
                    (this.state.openInstallChaincode) ? (
                        <ChaincodeInstall key={"ccinstall_" + this.state.openInstallChaincode}
                            open={this.state.openInstallChaincode}
                            onClose={this.handleCloseChaincodeInstall}
                            callBack={this.discoverUpdate}
                            ccInstallOption={this.state.ccInstallOption}
                        />
                    ) : null
                }

                {
                    (this.state.openInstantiateChaincode) ? (
                        <ChaincodeInstantiate key={"ccinstantiate_" + this.state.openInstantiateChaincode}
                            open={this.state.openInstantiateChaincode}
                            onClose={this.handleCloseChaincodeInstantiate}
                            callBack={this.discoverUpdate}
                            ccInstantiateOption={this.state.ccInstantiateOption}
                        />
                    ) : null
                }

                {
                    (this.state.openExecuteChaincode) ? (
                        <ChaincodeExecute key={"ccexecute_" + this.state.openExecuteChaincode}
                            open={this.state.openExecuteChaincode}
                            onClose={this.handleCloseChaincodeExecute}
                            callBack={this.discoverUpdate}
                            ccExecuteOption={this.state.ccExecuteOption}
                        />
                    ) : null
                }

                {
                    (this.state.openLedgerDetail) ? (
                        <LedgerDetail key={"ledger" + this.state.openLedgerDetail}
                            open={this.state.openLedgerDetail}
                            onClose={this.handleCloseLedgerQuery}
                            ledgerQueryOption={this.state.ledgerQueryOption}
                        />
                    ) : null
                }

                {
                    (this.state.openCreateChannel) ? (
                        <ChannelCreate key={"channelcreate_" + this.state.openCreateChannel}
                            open={this.state.openCreateChannel}
                            onClose={this.handleCloseChannelCreate}
                            callBack={this.discoverUpdate}
                            channelCreateOption={this.state.channelCreateOption}
                        />
                    ) : null
                }

                {
                    (this.state.openJoinChannel) ? (
                        <ChannelJoin key={"channeljoin_" + this.state.openJoinChannel}
                            open={this.state.openJoinChannel}
                            onClose={this.handleCloseChannelJoin}
                            callBack={this.discoverUpdate}
                            channelJoinOption={this.state.channelJoinOption}
                        />
                    ) : null
                }

            </React.Fragment>
        );
    }
}

export default withStyles(styles)(PeerOverview);