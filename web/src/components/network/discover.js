import React from 'react';

import Grid from '@material-ui/core/Grid';
import PeerOverview from './peer_overview';
import { withStyles } from '@material-ui/core/styles';
import { getCurrIDConn } from "../../common/utils";
import { log } from "../../common/log";
import Typography from '@material-ui/core/Typography';
import * as CONST from "../../common/constants";

const styles = theme => ({
    root: {
        display: 'flex',
    },
    toolbar: {
        paddingRight: 24, // keep right padding when drawer closed
    },
    toolbarIcon: {
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'flex-end',
        padding: '0 8px',
        ...theme.mixins.toolbar,
    },
    appBar: {
        zIndex: theme.zIndex.drawer + 1,
        transition: theme.transitions.create(['width', 'margin'], {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen,
        }),
    },
    appBarSpacer: {
        height: 20
    },
    content: {
        flexGrow: 1,
        height: '80vh',
        overflow: 'auto',
    },
    container: {
        paddingTop: theme.spacing(4),
        paddingBottom: theme.spacing(4),
    },
    paper: {
        padding: theme.spacing(1),
        display: 'flex',
        overflow: 'auto',
        flexDirection: 'column',
    },
    fixedHeight: {
        height: 180,
    },
});


/**
 * 
 * @param {object} props: {peers: [peer, peer, ...]} 
 */
class __DiscoverComp extends React.Component {
    constructor(props) {
        super(props);
        this.state = { ...props };
        // peers, peerStatuses, channelLedgers, channelOrderers, channelChaincodes

        this.state.focusPeer = undefined;

        this.handleFocusPeer = this.handleFocusPeer.bind(this);

        log.debug("__DiscoverComp: constructor");
    }

    componentDidMount() {
    }

    componentDidUpdate() {
    }

    componentWillUnmount() {
    }

    handleFocusPeer(peer) {
        const comp = this;
        return function () {
            log.log("Focus...", peer);
            comp.setState({ focusPeer: peer });
        };
    }

    render() {
        const classes = this.state.classes;
        // const fixedHeightPaper = clsx(classes.paper, classes.fixedHeight);
        return (
            <div className={classes.root}>
                {/* <CssBaseline /> */}

                <div className={classes.appBarSpacer} />
                {/* <Container maxWidth="lg" className={classes.container}> */}
                <Grid container spacing={3}>
                    {
                        this.state.peers ? this.state.peers.map((peer) => {
                            peer.isFocus = this.state.focusPeer && peer.URL === this.state.focusPeer.URL;
                            peer.isShow = !this.state.focusPeer || peer.isFocus
                            return peer;
                        }).filter((peer) => {
                            return peer.isShow;
                        }).map((peer) => {
                            return (
                                <Grid item xs={12} md={peer.isFocus ? 12 : 4} lg={peer.isFocus ? 12 : 3} key={peer.name}>
                                    {/* <Paper className={fixesdHeightPaper}> */}
                                    <PeerOverview key={peer.URL + peer.isFocus}
                                        peer={peer}
                                        peers={this.state.peers}                                        
                                        peerStatuses={this.state.peerStatuses}
                                        channelLedgers={this.state.channelLedgers}
                                        channelOrderers={this.state.channelOrderers}
                                        channelChaincodes={this.state.channelChaincodes}
                                        onFocus={this.handleFocusPeer(peer)}
                                        onLeaveFocus={this.handleFocusPeer(undefined)}
                                        isFocus={peer.isFocus}
                                    />
                                    {/* </Paper> */}
                                </Grid>);
                        }) : null
                    }
                </Grid>
                {/* </Container> */}


            </div>
        );
    }

}

const DiscoverComp = withStyles(styles)(__DiscoverComp);

class Discover extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            ...props,
            resCode: 200,
            peers: [],
            ledgers: {},
            errMsg: "",

        };

        this.isRunning = true;
        this.updateTO = undefined;

        this.discoverUpdate = this.discoverUpdate.bind(this);

        log.debug("Discover: constructor", this.state.isRunning);
    }

    discoverUpdate() {
        // Remove the current timeout to avoid any duplicate invoking.
        if (!this.isRunning) {
            return;
        }

        const currIDConn = getCurrIDConn();

        const reqBody = {
            connection: currIDConn
        };

        let myHeaders = new Headers({
            'Accept': 'application/json',
            'Content-Type': 'application/json'
        });
        let request = new Request(CONST.getServiceURL("/network/discover"), {
            method: 'POST',
            //mode: 'cors',
            //credentials: 'include',
            headers: myHeaders,
            body: JSON.stringify(reqBody),
        });

        // TODO issue: it always running... although it is unmounted.
        // TODO no discover network if the peer is focused with only one
        log.debug("Discover network.");

        fetch(request)
            .then(response => response.json())
            .then(result => {
                log.debug("Discover result", result);
                // result is with peers
                if (this.isRunning) {
                    this.setState(result);
                    // "peers":             networkOverview.Peers,
                    // "channelLedgers":    networkOverview.ChannelLedgers,
                    // "channelChaincodes": networkOverview.ChannelChainCodes,
                    // "channelOrderers":   networkOverview.ChannelOrderers,
                    this.updateTO = setTimeout(this.discoverUpdate, CONST.TIMEOUT_NETWORKOVERVIEW_UPDATE);
                }
            });
    }

    componentDidMount() {
        log.debug("Discover: componentDidMount", this.state.isRunning);
        this.discoverUpdate();
    }

    componentDidUpdate() {
    }

    componentWillUnmount() {
        log.debug("Clear timeout task of the discover update");
        //clearTimeout(this.updateTO);  // To avoid memory leak.
        this.isRunning = false;
        clearTimeout(this.updateTO);
    }

    render() {
        if (this.state.resCode === 200) {
            return (<DiscoverComp key={"peers" + (this.state.peers || []).length}
                peers={this.state.peers}
                peerStatuses={this.state.peerStatuses}
                channelLedgers={this.state.channelLedgers}
                channelChaincodes={this.state.channelChaincodes}
                channelOrderers={this.state.channelOrderers}
            />);
        }
        else {
            return (<Typography>
                Error!
                {this.state.resCode}
                {this.state.errMsg}
            </Typography>);
        }
    }
}

export default Discover;
