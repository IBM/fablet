import React from 'react';
import requestPromise from "request-promise-native";
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import Typography from '@material-ui/core/Typography';
import { withStyles } from '@material-ui/core/styles';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import Box from '@material-ui/core/Box';
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import Grid from '@material-ui/core/Grid';
import Fade from '@material-ui/core/Fade';
import CircularProgress from '@material-ui/core/CircularProgress';
import { getCurrIDConn, outputServiceErrorResult, formatTime, cntTrim } from "../../common/utils";
import { log } from "../../common/log";
import i18n from '../../i18n';
import * as CONST from "../../common/constants";
import { Tooltip } from '@material-ui/core';

import BlockEventChart from "./blockchart";

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
    }
});

class LedgerDetail extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            ...props,
            // onClose: props.onClose
            // open: props.open
            // ledgerQueryOption
            //    channelID, target

            loading: false,
            ledger: {},
            blocks: [],
            currBlock: undefined,

            queryError: "",
            isQuerying: false,

            fullChart: false
        };

        this.isRunning = true;
        this.websocket = null;

        this.chart = null;
        this.eventDataList = [];
        this.updateBlockEventChart = null;
        // 3 minutes
        this.chartDuration = 60 * 1000 * 3;

        this.handleFinished = this.handleFinished.bind(this);
        this.ledgerUpdate = this.ledgerUpdate.bind(this)
        this.blocksUpdate = this.blocksUpdate.bind(this)
        this.handleShowBlockDetail = this.handleShowBlockDetail.bind(this);
        this.handleQueryBlock = this.handleQueryBlock.bind(this);
        this.handleUpdatePreviousBlocks = this.handleUpdatePreviousBlocks.bind(this);
        this.handleFullChart = this.handleFullChart.bind(this);
        log.debug("LedgerDetail: constructor");
    }

    componentDidMount() {
        // Update at first
        this.ledgerUpdate();

        // Use block event to monitor
        this.monitorBlockEvent(this.state.ledgerQueryOption.channelID);

        // Show chart
        this.initChart();
        // Continue showing chart
        this.showBlockEventChart();
    }

    handleFullChart() {
        this.setState({ fullChart: !this.state.fullChart });
        setTimeout(() => this.chart.resize(this.state.fullChart), 300);
    }

    getFullChartCardStyle() {
        return this.state.fullChart ? {
            width: window.screen.width * 0.6,
            height: window.screen.height * 0.6 - 200,
            position: "absolute",
            zIndex: 999,
            backgroundColor: "#ffffff",
            border: "3px #eee solid",
            cursor: "pointer"
        } : {
                cursor: "pointer"
            };
    }

    getFullChartDivStyle() {
        return this.state.fullChart ? {
            width: window.screen.width * 0.6 - 100,
            height: window.screen.height * 0.6 - 250,
            cursor: "pointer"
        } : {
                width: 320,
                height: 56,
                cursor: "pointer"
            };
    }

    initChart() {
        log.debug("detail ............ initChart");
        if (!document.getElementById("chartDiv")) {
            // Use '()=>' function to have 'this' works.
            setTimeout(() => this.initChart(), 1000);
            return;
        }
        let chart = new BlockEventChart("chartDiv");
        chart.initChart();
        this.chart = chart;
    }

    showBlockEventChart() {
        log.debug("showBlockEventChart...");

        const detailPage = this;

        if (!detailPage.chart) {
            setTimeout(() => detailPage.showBlockEventChart(), 1000);
            return;
        }

        const data = {};
        const now = this.floorMilliSecond((new Date()).getTime());
        // At most 60 seconds.
        const LEN = this.chartDuration;
        // Add 1 second
        const BEGIN = now - LEN + 1000;
        detailPage.eventDataList
            .map(d => [d.updateTime, d.TXNumber])
            .filter(d => d[0] >= BEGIN)
            .forEach(d => {
                if (data[d[0]]) {
                    data[d[0]] += d[1];
                }
                else {
                    data[d[0]] = d[1];
                }
            });;

        const filledData = [];
        for (let i = BEGIN; i <= now; i += 1000) {
            filledData.push([i, data[i] ? data[i] : 0]);
        }

        detailPage.chart.showChart(filledData);
        this.updateBlockEventChart = setTimeout(() => detailPage.showBlockEventChart(), 2000);
    }

    componentDidUpdate() {
    }

    componentWillUnmount() {
        this.setState({ open: false });
        this.isRunning = false;
        //clearTimeout(this.updateTO);
        //clearTimeout(this.updateBlocksTO);
        clearTimeout(this.updateBlockEventChart);
        this.closeBlockEvent();
    }

    handleFinished() {
        this.setState({ open: false, isRunning: false });
        this.state.onClose();
    }

    handleShowBlockDetail(block) {
        const comp = this;
        return function () {
            comp.setState({ queryError: "", currBlock: block });
        };
    }

    handleUpdatePreviousBlocks() {
        this.blocksUpdatePrevious();
    }

    closeBlockEvent() {
        // TODO the actual closing will happen after 60 seconds. Please see pong period example at gorilla/websocket.
        // It's better to use active pong from client to server, but make it easy now.
        if (this.websocket) {
            log.debug("Close websocket when quit.");
            this.websocket.close();
        }
    }

    addBlockEventData(eventData) {
        log.debug("Add event data: ", eventData.updateTime, eventData.number);
        if (this.state.ledger.height && (this.state.ledger.height - 1) >= eventData.number) {
            return;
        }
        eventData.updateTime = this.floorMilliSecond(eventData.updateTime);
        this.eventDataList.push(eventData);
        const now = this.floorMilliSecond((new Date()).getTime());
        // TODO concurrent issue.
        // Only less than 60 second records remain
        this.eventDataList = this.eventDataList.filter(d => (now - d.updateTime) < this.chartDuration);
    }

    floorMilliSecond(milliSec) {
        return Math.floor(milliSec / 1000) * 1000;
    }

    monitorBlockEvent(channelID) {
        if (!this.isRunning || !this.state.open) {
            return;
        }

        const detailPage = this;
        log.debug(`Begin block event websocket for channel ${channelID}`);
        detailPage.websocket = new WebSocket(CONST.getWebsocketURL("/event/blockevent"));

        const currIDConn = getCurrIDConn();
        const reqBody = {
            connection: currIDConn,
            channelID: channelID
        };

        detailPage.websocket.onopen = function (e) {
            log.debug("Websocket connection established and then sending reqBody to server.");
            detailPage.websocket.send(JSON.stringify(reqBody));
        };

        detailPage.websocket.onmessage = function (event) {
            log.debug(`Data received from server`);
            if (event && event.data) {
                // TODO
                // There might be problem if many updates happen concurrently.
                detailPage.addBlockEventData(JSON.parse(event.data));
                detailPage.ledgerUpdate();
            }
        };

        // Normally won't be closed.
        detailPage.websocket.onclose = function (event) {
            log.debug(`Websocket connection closed, wasClean: ${event.wasClean}, code: ${event.code}, reason: ${event.reason}`);
            setTimeout(function () {
                log.debug("Reconnect the websocket.");
                detailPage.monitorBlockEvent(channelID);
            }, 1000);
        };

        detailPage.websocket.onerror = function (error) {
            log.debug(`Websocket connection error: ${error.message}. And then restart it.`);
            //setTimeout
        };
    }

    async ledgerUpdate() {
        log.log("Begin to update the ledger details.", this.state.ledgerQueryOption.channelID, this.state.ledgerQueryOption.target);

        //clearTimeout(this.updateTO);
        if (!this.isRunning || !this.state.open) {
            return;
        }

        // To check id/conn validity.
        const currIDConn = getCurrIDConn();

        const reqBody = {
            connection: currIDConn,
            channelID: this.state.ledgerQueryOption.channelID,
            targets: [this.state.ledgerQueryOption.target]
        };

        let myHeaders = new Headers({
            'Accept': 'application/json',
            'Content-Type': 'application/octet-stream'
        });

        let option = {
            url: CONST.getServiceURL("/ledger/query"),
            method: 'POST',
            headers: myHeaders,
            json: true,
            resolveWithFullResponse: true,
            body: reqBody
        };

        try {
            log.debug("LedgerQuery result");
            this.setState({ loading: true });
            let result = await requestPromise(option);
            this.setState({ loading: false });
            if (result) {
                result = result.body;
                if (result.resCode === 200) {
                    this.setState({ ledger: result.ledger });
                    this.blocksUpdate();
                }
                else {
                    // this.setState({ executeError: result.resCode + ". " + result.errMsg });
                }
            }
            else {
                // this.setState({ executeError: i18n("no_result_error") });
            }
            // TODO timeout
            // this.updateTO = setTimeout(this.ledgerUpdate, CONST.TIMEOUT_LEDGER_UPDATE);
        }
        catch (e) {
            log.error("Exception in fetching:", e);
            this.setState({
                // executeError: String(err),
                loading: false
            });
        }
    }

    // TODO the timeout function might be removed and won't back.
    async blocksUpdate() {
        if (!this.isRunning || !this.state.open) {
            return;
        }
        log.log("Begin to update the ledger blocks.", this.state.ledgerQueryOption.channelID, this.state.ledgerQueryOption.target);

        let latestBlockNum = -1;
        if (this.state.blocks && this.state.blocks.length > 0) {
            latestBlockNum = this.state.blocks[0].number;
        }
        let ledgerHeight = 0;
        if (this.state.ledger) {
            ledgerHeight = this.state.ledger.height;
        }
        if (ledgerHeight <= 0 || latestBlockNum >= ledgerHeight - 1) {
            log.log("The current blocks list is up to date.");
            return;
        }
        if (latestBlockNum === -1) {
            // Initial update, to avoid too many blocks.
            latestBlockNum = ledgerHeight - CONST.INITIAL_BLOCKS - 1;
            if (latestBlockNum < -1) {
                latestBlockNum = -1;
            }
        }

        const begin = latestBlockNum + 1;
        const len = (ledgerHeight - 1) - (latestBlockNum + 1) + 1;

        try {
            log.debug("LedgerUpdate result");
            this.setState({ loading: true });
            let result = await this.blocksUpdateWithRange(begin, len);
            this.setState({ loading: false });

            if (result) {
                result = result.body;
                if (result.resCode === 200) {
                    this.setState({ blocks: this.insertBlocks(result.blocks) });
                }
                else {
                    // this.setState({ executeError: result.resCode + ". " + result.errMsg });
                }
            }
            else {
                // this.setState({ executeError: i18n("no_result_error") });
            }
            // TODO timeout
            // this.updateBlocksTO = setTimeout(this.blocksUpdate, CONST.TIMEOUT_LEDGER_UPDATE);
        }
        catch (e) {
            log.error("Exception in fetching:", e);
            this.setState({
                // executeError: String(err),
                loading: false
            });
        }
    }

    async blocksUpdatePrevious() {
        this.setState({ loading: true });
        log.log("Begin to update the previous ledger blocks.", this.state.ledgerQueryOption.channelID, this.state.ledgerQueryOption.target);

        if (!this.state.blocks || this.state.blocks.length <= 0) {
            return;
        }
        const earliest = this.state.blocks[this.state.blocks.length - 1].number;
        if (earliest <= 0) {
            return;
        }
        let begin = (earliest - 1) - CONST.INITIAL_BLOCKS + 1;
        if (begin < 0) {
            begin = 0;
        }
        const len = (earliest - 1) - begin + 1;

        try {
            log.debug("blocksUpdatePrevious result");

            let result = await this.blocksUpdateWithRange(begin, len);
            this.setState({ loading: false });

            if (result) {
                result = result.body;
                if (result.resCode === 200) {
                    this.setState({ blocks: this.appendBlocks(result.blocks) });
                }
                else {
                    // this.setState({ executeError: result.resCode + ". " + result.errMsg });
                }
            }
            else {
                // this.setState({ executeError: i18n("no_result_error") });
            }
            // TODO timeout
            // this.updateBlocksTO = setTimeout(this.blocksUpdate, CONST.TIMEOUT_LEDGER_UPDATE);
        }
        catch (e) {
            log.error("Exception in fetching:", e);
            this.setState({
                // executeError: String(err),
                loading: false
            });
        }
    }

    // Only append previous, no insert now.
    appendBlocks(blocks) {
        if (!blocks || blocks.length <= 0) {
            return;
        }
        return (this.state.blocks || []).concat(...(blocks.reverse()));
    }

    insertBlocks(blocks) {
        if (!blocks || blocks.length <= 0) {
            return;
        }
        return blocks.reverse().concat(...(this.state.blocks));
    }

    async blocksUpdateWithRange(begin, len) {
        // To check id/conn validity.
        const currIDConn = getCurrIDConn();

        const reqBody = {
            connection: currIDConn,
            channelID: this.state.ledgerQueryOption.channelID,
            targets: [this.state.ledgerQueryOption.target],
            begin: begin,
            len: len
        }

        let myHeaders = new Headers({
            'Accept': 'application/json',
            'Content-Type': 'application/octet-stream'
        });
        let option = {
            url: CONST.getServiceURL("/ledger/block"),
            method: 'POST',
            //mode: 'cors',
            //credentials: 'include',
            headers: myHeaders,
            body: reqBody,
            json: true,
            resolveWithFullResponse: true,
        };

        return await requestPromise(option);
    }

    handleQueryBlock(event) {
        event.preventDefault();

        this.setState({
            queryError: "",
            isQuerying: true
        });

        // To check id/conn validity.
        const currIDConn = getCurrIDConn();
        const formData = new FormData(event.target);

        const reqBody = {
            connection: currIDConn,
            channelID: this.state.ledgerQueryOption.channelID,
            targets: [this.state.ledgerQueryOption.target],
            queryKey: formData.get("queryKey")
        };

        log.log("Instantiate chaincode: ", reqBody);

        let myHeaders = new Headers({
            'Accept': 'application/json',
            'Content-Type': 'application/octet-stream'
        });
        let request = new Request(CONST.getServiceURL("/ledger/blockany"), {
            method: 'POST',
            mode: 'cors',
            //credentials: 'include',
            headers: myHeaders,
            body: JSON.stringify(reqBody),
        });

        fetch(request)
            .then(response => response.json())
            .then(result => {
                log.debug("Block query result", result);

                this.setState({
                    isQuerying: false
                });

                if (result) {
                    if (result.resCode === 200) {
                        this.setState({ currBlock: result.block })
                    }
                    else {
                        this.setState({ queryError: outputServiceErrorResult(result) });
                    }
                }
                else {
                    this.setState({ queryError: i18n("no_result_error") });
                }
            })
            .catch(function (err) {
                this.setState({
                    queryError: String(err),
                    isQuerying: false
                });
            });
    }


    showTitleValue(title, value, key) {
        const classes = this.props.classes;
        return (
            <React.Fragment key={key}>
                <Grid item xs={2} className={classes.normalTitle}>{title}</Grid>
                <Grid item xs={10} className={classes.normalFieldFontWidth}>{value}</Grid>
            </React.Fragment>
        );
    }

    showActionEndorsers(endorsers) {
        const classes = this.props.classes;
        let idx = 0;
        return endorsers.map(endorser => {
            return (
                <Grid container key={"endorser_" + (idx++)}>
                    <Grid item xs={12} className={classes.normalTitle}>{i18n("block_endorser_of_all_endorsers", idx, endorsers.length)}</Grid>
                    <Grid container style={{ marginLeft: 20 }}>
                        {this.showTitleValue(i18n("msp_id"), endorser.MSPID)}
                        {/* {this.showTitleValue(i18n("id_cn"), endorser.commonName)} */}
                        {this.showTitleValue(i18n("id_subject"), endorser.subject)}
                        {this.showTitleValue(i18n("id_issuer"), endorser.issuer)}
                    </Grid>
                </Grid>
            );
        });
    }

    showTransactionActions(actions) {
        const classes = this.props.classes;
        let idx = 0;
        return actions.map(action => {
            return (
                <Grid container style={{ marginLeft: 20 }} key={"action_" + (idx++)}>
                    <Grid item xs={12} className={classes.normalTitle}>{i18n("block_action_of_all_actions", idx, actions.length)}</Grid>
                    <Grid container style={{ marginLeft: 20 }}>
                        {this.showTitleValue(i18n("chaincode"), action.chaincodeName)}
                        {/* TODO lscc arguments with byte code not fine, maybe others have same issue. */}
                        {this.showTitleValue(i18n("chaincode_arguments"), action.arguments.join(";  "))}
                        {this.showActionEndorsers(action.endorsers)}
                        {this.showProposalResponse(action.proposalResponse)}
                    </Grid>
                </Grid>
            );
        });
    }

    showReadSet(readSet) {
        const classes = this.props.classes;
        const readsetList = readSet || [];
        let idx = 0;
        return (
            <Grid container style={{ marginLeft: 20 }}>
                {
                    readsetList.map(read => {
                        return (
                            <React.Fragment key={read.key + read.verBlockNum + read.verTxNum}>
                                <Grid item xs={12} className={classes.normalTitle}>{i18n("readset_of_all", ++idx, readsetList.length)}</Grid>
                                <Grid container style={{ marginLeft: 20 }}>
                                    {this.showTitleValue(i18n("key"), read.key)}
                                    {this.showTitleValue(i18n("block"), read.verBlockNum)}
                                    {this.showTitleValue(i18n("transaction"), read.verTxNum)}
                                </Grid>
                            </React.Fragment>
                        );
                    })
                }
            </Grid>);
    }

    showWriteSet(writeSet, isLscc) {
        const classes = this.props.classes;
        const writesetList = writeSet || [];
        let idx = 0;
        return (
            <Grid container style={{ marginLeft: 20 }}>
                {
                    writesetList.map(write => {
                        return (
                            <React.Fragment key={write.key}>
                                <Grid item xs={12} className={classes.normalTitle}>{i18n("writeset_of_all", ++idx, writesetList.length)}</Grid>
                                <Grid container style={{ marginLeft: 20 }}>
                                    {this.showTitleValue(i18n("key"), write.key)}
                                    {this.showTitleValue(i18n("isDelete"), String(write.isDelete))}
                                    {isLscc ? this.showChaincodeData(write.value) : this.showTitleValue(i18n("value"), write.value)}
                                </Grid>
                            </React.Fragment>
                        );
                    })
                }
            </Grid>);
    }

    showChaincodeData(ccd) {
        const classes = this.props.classes;
        let idx = 0;
        return (
            <React.Fragment>
                <Grid item xs={12} className={classes.normalTitle}>{i18n("chaincode_deployed")}</Grid>
                <Grid container style={{ marginLeft: 20 }}>
                    {this.showTitleValue(i18n("chaincode"), ccd.name + "@" + ccd.version)}
                    {
                        (ccd.principals || []).map(princ => {
                            return (
                                this.showTitleValue(i18n("principal"), princ, "princ" + (idx++))
                            );
                        })
                    }
                    {this.showTitleValue(i18n("rule"), ccd.rule)}
                </Grid>
            </React.Fragment>);
    }

    showProposalResponse(pr) {
        const classes = this.props.classes;
        let idx = 0;
        let isLscc = (pr.chaincode || {}).name === "lscc";
        const rwContent = ((pr.txReadWriteSet || {}).nsReadWriteSets || []).map(nsrw => {
            return (
                <Grid container key={"txrw" + (idx++)}>
                    {this.showTitleValue(i18n("namespace"), nsrw.nameSpace)}
                    {this.showReadSet(nsrw.kvReadSet)}
                    {this.showWriteSet(nsrw.kvWriteSet, isLscc)}
                </Grid>
            );
        });

        return (
            <Grid container>
                <Grid item xs={12} className={classes.normalTitle}>{i18n("proposal_response")}</Grid>
                <Grid container style={{ marginLeft: 20 }}>
                    {this.showTitleValue(i18n("chaincode"), (pr.chaincode || {}).name + "@" + (pr.chaincode || {}).version)}
                    {isLscc ? this.showChaincodeData(pr.response) : this.showTitleValue(i18n("response"), pr.response)}
                    {rwContent}
                </Grid>
            </Grid>
        );
    }

    showBlockDetail(block) {
        if (!this.state.open) {
            return null;
        }

        const classes = this.props.classes;
        let idx = 0;

        return (
            <Grid container>
                {this.showTitleValue(i18n("block_number"), block.number)}
                {this.showTitleValue(i18n("block_hash"), block.blockHash)}
                {this.showTitleValue(i18n("block_data_hash"), block.dataHash)}
                {this.showTitleValue(i18n("block_previous_hash"), block.previousHash)}
                {this.showTitleValue(i18n("block_time"), formatTime(block.time))}

                {
                    block.transactions.map(transaction => {
                        return (
                            <React.Fragment key={"transaction_" + (idx++)}>
                                <Grid item xs={12} className={classes.normalTitle} style={{ marginTop: 10 }}>{i18n("block_tx_of_all_tx", idx, block.transactions.length)}</Grid>
                                {this.showTransactionActions(transaction.actions)}
                            </React.Fragment>
                        )
                    })
                }

            </Grid>);
    }

    showAllBlocks() {
        const classes = this.props.classes;
        const allBlocks = (this.state.blocks || []).map(block => {
            return (
                <Card className={classes.ledgerDetail} key={"block_" + block.number} onClick={this.handleShowBlockDetail(block)}>
                    <CardContent>
                        <Typography display="inline" color="textSecondary">
                            {i18n("block")}: {block.number}
                        </Typography>
                        <Tooltip interactive title={i18n("block_hash") + ": " + block.blockHash}>
                            <Typography display="inline" className={classes.normalField} style={{ marginLeft: 20 }}>
                                {cntTrim(block.blockHash, 8)}
                            </Typography>
                        </Tooltip>

                        <Typography className={classes.normalField}>
                            {i18n("transaction")}: {block.transactions.length}
                        </Typography>

                        <Typography className={classes.normalField}>
                            {formatTime(block.time)}
                        </Typography>
                    </CardContent>
                </Card>);
        });

        return (
            <React.Fragment>
                {allBlocks}
                {
                    this.state.blocks && this.state.blocks.length > 0 && this.state.blocks[this.state.blocks.length - 1].number > 0 ?
                        this.showPreviousBlocks() : null
                }
            </React.Fragment>
        );
    }

    showPreviousBlocks() {
        const classes = this.props.classes;
        return (
            <Card className={classes.ledgerDetail} key={"block_prev_"} onClick={this.handleUpdatePreviousBlocks}>
                <CardContent>
                    <Typography display="inline" color="textSecondary">
                        {i18n("previous_blocks")}...
                    </Typography>
                </CardContent>
            </Card>);
    }

    render() {
        const classes = this.props.classes;

        return this.state.ledger ? (
            <Dialog
                disableBackdropClick
                disableEscapeKeyDown
                maxWidth={false}
                fullWidth
                style={{ maxHeight: window.screen.height * 0.9 }}
                open={this.state.open}
                onClose={this.handleFinished}
            >
                <DialogTitle id="dialog_title">{i18n("ledger_query")}</DialogTitle>
                <DialogContent>
                    <Grid container style={{ height: window.screen.height * 0.6 }} spacing={2}>
                        <Grid item xs={8}>
                            <Grid container spacing={3}>
                                <Grid item xs={12}>
                                    <Card className={classes.ledgerSummary}>
                                        <CardContent>
                                            <Typography color="textSecondary" gutterBottom>
                                                {i18n("ledger_height")}
                                            </Typography>
                                            <Typography className={classes.ledgerSummaryField}>
                                                {this.state.ledger.height}
                                            </Typography>
                                        </CardContent>
                                    </Card>
                                    <Card className={classes.ledgerSummary}>
                                        <CardContent>
                                            <Typography color="textSecondary" gutterBottom>
                                                {i18n("channel")}
                                            </Typography>
                                            <Tooltip title={this.state.ledgerQueryOption.channelID}>
                                                <Typography nowrap="true" className={classes.ledgerSummaryField}>
                                                    {this.state.ledgerQueryOption.channelID}
                                                </Typography>
                                            </Tooltip>
                                        </CardContent>
                                    </Card>

                                    {/* <Card className={classes.ledgerSummary}>
                                        <CardContent>
                                            <Typography color="textSecondary" gutterBottom>
                                                {i18n("endorser")}
                                            </Typography>
                                            <Tooltip title={this.state.ledger.endorser || ""}>
                                                <Typography className={classes.ledgerSummaryField}>
                                                    {this.state.ledger.endorser}
                                                </Typography>
                                            </Tooltip>
                                        </CardContent>
                                    </Card> */}

                                    <Card className={classes.ledgerSummary}>
                                        <CardContent style={this.getFullChartCardStyle()} onClick={this.handleFullChart}>
                                            <div id="chartDiv" style={this.getFullChartDivStyle()}></div>
                                        </CardContent>
                                    </Card>
                                </Grid>

                                <Grid item xs={12}>
                                    <Box className={classes.fixedHeightBox}>
                                        {
                                            this.state.blocks ? this.showAllBlocks() : null
                                        }
                                    </Box>
                                </Grid>


                            </Grid>


                        </Grid>
                        <Grid item xs={4}>
                            <Grid container spacing={3}>
                                <Grid item xs={12}>
                                    <form onSubmit={this.handleQueryBlock}>
                                        <Grid container spacing={3}>
                                            <Grid item xs={10}>
                                                <TextField
                                                    fullWidth
                                                    label={i18n("block_query_any")}
                                                    variant="outlined"
                                                    required
                                                    id="queryKey"
                                                    name="queryKey"
                                                    InputProps={{ classes: { input: classes.normalField } }}
                                                />
                                            </Grid>
                                            <Grid item xs={2}>
                                                <Button
                                                    type="submit"
                                                    autoFocus
                                                    variant="contained"
                                                    color="primary"
                                                    style={{ marginLeft: "auto", height: 52 }}
                                                    disabled={this.state.isQuerying}>
                                                    {i18n("query")}
                                                </Button>
                                            </Grid>
                                        </Grid>
                                    </form>
                                </Grid>

                                <Grid item xs={12}>
                                    <Box className={classes.fixedHeightBox2}>
                                        {
                                            this.state.queryError ?
                                                (<Typography className={classes.normalField}>{this.state.queryError}</Typography>)
                                                :
                                                this.state.currBlock ?
                                                    this.showBlockDetail(this.state.currBlock)
                                                    :
                                                    null
                                        }
                                    </Box>
                                </Grid>

                            </Grid>
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

export default withStyles(styles)(LedgerDetail);