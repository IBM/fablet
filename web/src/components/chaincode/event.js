import React from 'react';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import Typography from '@material-ui/core/Typography';
import { withStyles } from '@material-ui/core/styles';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Box from '@material-ui/core/Box';
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import Grid from '@material-ui/core/Grid';
import Fade from '@material-ui/core/Fade';
import CircularProgress from '@material-ui/core/CircularProgress';
import { getCurrIDConn } from "../../common/utils";
import { log } from "../../common/log";
import i18n from '../../i18n';
import * as CONST from "../../common/constants";
import { Tooltip } from '@material-ui/core';

const styles = (theme) => ({
    normalField: {
        fontSize: 13,
    },
    normalFieldFontWidth: {
        fontSize: 11,
        fontFamily: "Consolas, Monaco, monospace",
    },
    normalTitle: {
        fontSize: 13,
        fontWeight: 600,
    },
    ledgerSummary: {
        minWidth: 120,
        display: "inline-block",
        margin: 5,
    },
    ledgerSummaryField: {
        fontSize: 18,
    },
    ledgerDetail: {
        width: 240,
        height: 90,
        display: "inline-block",
        margin: 5,
        cursor: "pointer"
    },
    fixedHeightBox: {
        maxHeight: window.screen.height * 0.6 - 160,
        overflow: "auto"
    },
    fixedHeightBox2: {
        maxHeight: window.screen.height * 0.6 - 100,
        overflow: "auto"
    },
    fieldDetail: {
        fontSize: 11,
    },
});

class ChaincodeEvent extends React.Component {
    constructor(props) {
        super(props);        
        this.state = {
            ...props,
            // onClose: props.onClose
            // open: props.open
            // ccEventOption
            //    channelID, chaincodeID, eventFilter
            ccEvents: [],
            errMsg: ""
        };

        this.isRunning = true;
        this.websocket = null;

        this.handleChangeEventFilter = this.handleChangeEventFilter.bind(this);
        this.handleFinished = this.handleFinished.bind(this);

        log.debug("ChaincodeEvent: constructor");
    }

    componentDidMount() {
        // Use block event to monitor
        this.monitorChaincodeEvent(this.state.ccEventOption.channelID, this.state.ccEventOption.chaincodeID);
    }

    componentDidUpdate() {
    }

    componentWillUnmount() {
        this.setState({ open: false });
        this.isRunning = false;
        this.closeChaincodeEvent();
    }

    handleFinished() {
        this.setState({ open: false, isRunning: false });
        this.state.onClose();
    }

    handleChangeEventFilter(event) {
        event.preventDefault();
        if (this.websocket) {
            const formData = new FormData(event.target);
            const reqBody = this.createReqBody(formData.get("eventFilter"));            
            this.websocket.send(JSON.stringify(reqBody));
        }
    }

    closeChaincodeEvent() {
        // TODO the actual closing will happen after several seconds. Please see pong period example at gorilla/websocket.
        // It's better to use active pong from client to server, but make it easy now.
        if (this.websocket) {
            log.debug("Close websocket when quit.");
            this.websocket.close();
        }
    }

    createReqBody(eventFilter) {
        if (!eventFilter) {
            eventFilter= ".*";
        }
        const currIDConn = getCurrIDConn();
        const reqBody = {
            connection: currIDConn,
            channelID: this.state.ccEventOption.channelID,
            chaincodeID: this.state.ccEventOption.chaincodeID, 
            eventFilter: eventFilter
        };
        return reqBody;
    }

    monitorChaincodeEvent(channelID, chaincodeID) {
        if (!this.isRunning || !this.state.open) {
            return;
        }

        const detailPage = this;
        log.debug(`Begin chaincode event websocket for channel ${channelID}, chaincode ${chaincodeID}`);
        detailPage.websocket = new WebSocket(CONST.getWebsocketURL("/event/chaincodeevent"));

        // Default event filter ".*".
        const reqBody = this.createReqBody();

        detailPage.websocket.onopen = function (e) {
            log.debug("Websocket connection established and then sending reqBody to server.");
            detailPage.websocket.send(JSON.stringify(reqBody));
        };

        detailPage.websocket.onmessage = function (event) {
            log.debug(`Data received from server`);
            if (event && event.data) {
                console.info("Get chaincode event: ", event.data);
                const res = JSON.parse(event.data);
                if (res.error) {
                    detailPage.setState({errMsg: res.error});
                }
                else {
                    const existCCEvents = [res].concat(detailPage.state.ccEvents);
                    detailPage.setState({
                        ccEvents: existCCEvents,
                        errMsg: ""
                    });
                }
            }
        };

        // Normally won't be closed.
        detailPage.websocket.onclose = function (event) {
            log.debug(`Websocket connection closed, wasClean: ${event.wasClean}, code: ${event.code}, reason: ${event.reason}`);
            setTimeout(function () {
                log.debug("Reconnect the websocket.");
                detailPage.monitorChaincodeEvent(channelID, chaincodeID);
            }, 1000);
        };

        detailPage.websocket.onerror = function (error) {
            log.debug(`Websocket connection error: ${error.message}. And then restart it.`);
            //setTimeout
        };
    }

    render() {
        const classes = this.props.classes;

        return this.state.ccEventOption ? (
            <Dialog
                disableBackdropClick
                disableEscapeKeyDown
                maxWidth={false}
                style={{ maxHeight: window.screen.height * 0.6 }}
                open={this.state.open}
            >
                <DialogTitle id="dialog_title">{i18n("chaincode_event")}</DialogTitle>
                <DialogContent style={{ height: window.screen.height * 0.6 }}>
                    <Grid container spacing={2}>
                        
                                <Grid item xs={5}>
                                    <Card className={classes.ledgerSummary}>
                                        <CardContent>
                                            <Typography color="textSecondary" gutterBottom>
                                                {i18n("channel")}
                                            </Typography>
                                            <Tooltip title={this.state.ccEventOption.channelID}>
                                                <Typography nowrap="true" className={classes.ledgerSummaryField}>
                                                    {this.state.ccEventOption.channelID}
                                                </Typography>
                                            </Tooltip>
                                        </CardContent>
                                    </Card>
                                    <Card className={classes.ledgerSummary}>
                                        <CardContent>
                                            <Typography color="textSecondary" gutterBottom>
                                                {i18n("chaincode")}
                                            </Typography>
                                            <Tooltip title={this.state.ccEventOption.channelID}>
                                                <Typography nowrap="true" className={classes.ledgerSummaryField}>
                                                {this.state.ccEventOption.chaincodeID}
                                                </Typography>
                                            </Tooltip>
                                        </CardContent>
                                    </Card>
                                </Grid>

                                <Grid item xs={7}>
                                    <form onSubmit={this.handleChangeEventFilter}>
                                        <Grid container spacing={3}>
                                            <Grid item xs={10}>
                                                <TextField
                                                    fullWidth
                                                    label={i18n("chaincode_event_filter")}
                                                    variant="outlined"
                                                    required
                                                    id="eventFilter"
                                                    name="eventFilter"
                                                    defaultValue={this.state.ccEventOption.eventFilter || ".*"}
                                                    // TODO readonly
                                                    InputProps={{ classes: { input: classes.normalField }, readOnly: true }}
                                                />
                                            </Grid>
                                            <Grid item xs={2}>
                                                <Button
                                                    type="submit"
                                                    autoFocus
                                                    variant="contained"
                                                    color="primary"
                                                    style={{ marginLeft: "auto", height: 52 }}>
                                                    {i18n("query")}
                                                </Button>
                                            </Grid>

                                            <Grid item xs={12}>
                                                {
                                                    this.state.errMsg ? (
                                                        <Typography color="error" style={{ marginRight: "auto", fontSize: 11 }}>{this.state.errMsg}</Typography>
                                                    ) : null
                                                }
                                            </Grid>
                                        </Grid>
                                    </form>
                                </Grid>

                                <Grid item xs={12}>
                                    <Box>
                                        <Table stickyHeader className={classes.table} size="small" aria-label="chaincodeeventss">
                                            <TableHead>
                                                <TableRow>
                                                    <TableCell>{i18n("chaincode")}</TableCell>
                                                    <TableCell align="right">{i18n("channel")}</TableCell>
                                                    <TableCell align="right">{i18n("event_name")}</TableCell>
                                                    <TableCell align="right">{i18n("payload")}</TableCell>
                                                    <TableCell align="right">{i18n("transaction_id")}</TableCell>
                                                    <TableCell align="right">{i18n("block_number")}</TableCell>
                                                    <TableCell align="right">{i18n("source_url")}</TableCell>
                                                </TableRow>
                                            </TableHead>

                                            <TableBody key={"tbody_ccevent_" + this.state.ccEventOption.chaincodeID}>
                                                {this.state.ccEvents.map((ccEvent) => {
                                                    return (
                                                        <TableRow key={"ccevent_" + ccEvent.TXID}>
                                                            <TableCell component="th" scope="row"><Typography noWrap className={classes.fieldDetail}>{ccEvent.chaincodeID}</Typography></TableCell>
                                                            <TableCell align="right"><Typography noWrap className={classes.fieldDetail}>{this.state.ccEventOption.channelID}</Typography></TableCell>
                                                            <TableCell align="right"><Typography noWrap className={classes.fieldDetail}>{ccEvent.eventName}</Typography></TableCell>
                                                            <TableCell align="right"><Typography noWrap className={classes.fieldDetail}>{ccEvent.payload}</Typography></TableCell>
                                                            <TableCell align="right"><Typography noWrap className={classes.fieldDetail}>{ccEvent.TXID}</Typography></TableCell>
                                                            <TableCell align="right"><Typography noWrap className={classes.fieldDetail}>{ccEvent.blockNumber}</Typography></TableCell>
                                                            <TableCell align="right"><Typography noWrap className={classes.fieldDetail}>{ccEvent.sourceURL}</Typography></TableCell>
                                                        </TableRow>
                                                    );
                                                })}
                                            </TableBody>

                                        </Table>

                                    </Box>
                                </Grid>
                    
                    </Grid>
                </DialogContent>

                <DialogActions>
                    <Grid
                        container
                        direction="row"
                        justify="flex-end"
                        alignItems="center"
                    >
                        <Grid item xs={11}>
                            <Fade
                                in={this.state.loading}
                                style={{
                                    transitionDelay: this.state.loading ? '100ms' : '0ms',
                                    width: 20,
                                    height: 20,
                                }}
                                // TODO if it can be used for other components?
                                unmountOnExit
                            >
                                <CircularProgress style={{ marginLeft: 60 }} />
                            </Fade>
                            <Typography color="error" style={{ marginRight: "auto", fontSize: 11 }}>{this.state.executeError}</Typography>
                            <Typography color="primary" style={{ marginRight: "auto", fontSize: 11 }}>{this.state.executeResult}</Typography>
                        </Grid>

                        <Grid item xs={1}>
                            <Box display="flex" flexDirection="row-reverse">
                                <Button
                                    onClick={this.handleFinished}
                                    autoFocus
                                    variant="contained"
                                    color="primary"
                                    style={{ marginLeft: "auto" }}
                                // disabled={this.state.loading}
                                >
                                    {i18n("ok")}
                                </Button>
                            </Box>
                        </Grid>
                    </Grid>

                </DialogActions>
            </Dialog >
        )
            : null;
    }
}

export default withStyles(styles)(ChaincodeEvent);