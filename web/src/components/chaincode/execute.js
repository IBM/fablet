import React from "react";
import Dialog from "@material-ui/core/Dialog";
import DialogActions from "@material-ui/core/DialogActions";
import DialogContent from "@material-ui/core/DialogContent";
import DialogTitle from "@material-ui/core/DialogTitle";
import Typography from "@material-ui/core/Typography";
import { withStyles } from "@material-ui/core/styles";
import Button from "@material-ui/core/Button";
import FormControl from '@material-ui/core/FormControl';
import FormHelperText from '@material-ui/core/FormHelperText';
import InputLabel from '@material-ui/core/InputLabel';
import Select from '@material-ui/core/Select';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import Checkbox from '@material-ui/core/Checkbox';
import MenuItem from '@material-ui/core/MenuItem';
import TextField from "@material-ui/core/TextField";
import Input from "@material-ui/core/Input";
import Grid from "@material-ui/core/Grid";
import Box from "@material-ui/core/Box";
import Fade from "@material-ui/core/Fade";
import CircularProgress from "@material-ui/core/CircularProgress";
import { getCurrIDConn } from "../../common/utils";
import { log } from "../../common/log";
import i18n from "../../i18n";
import * as CONST from "../../common/constants";

const styles = (theme) => ({
    formField: {
        fontSize: 13,
    },
    formControl: {
        margin: theme.spacing(1),
        minWidth: 120,
    },
    normalField: {
        fontSize: 11,
    },
    normalTitle: {
        fontSize: 11,
        fontWeight: 600,
    },
    fixedHeightBox: {
        maxHeight: window.screen.height * 0.6 - 360,
        overflow: "auto"
    },
});

const MenuProps = {
    PaperProps: {
        style: {
            maxHeight: 300,
            width: 300,
            maxWidth: "50%",
            fontSize: 13,
        },
    },
};

class ChaincodeExecute extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            ...props,
            // callBack: props.callBack,
            // ccExecuteOption
            //    actionType, channelID, name, peer, targets

            loading: false,
            executeError: "",
            executeResult: {},

            targets: props.ccExecuteOption.peer ? [props.ccExecuteOption.peer.URL] : [] // Selected targets
        };

        this.handleCancel = this.handleCancel.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
        this.handleExecuted = this.handleExecuted.bind(this);
        this.handleChangeTargets = this.handleChangeTargets.bind(this);
        log.debug("ChaincodeExecute: constructor");
    }

    componentDidUpdate() {
    }

    componentWillUnmount() {
    }

    handleCancel() {
        this.setState({ open: false });
        this.state.onClose();
    }

    handleExecuted() {
        this.setState({ open: false });
        this.state.onClose();
        // Callback from parent, to update the peer overview.
        this.state.callBack();
    }

    handleSubmit(event) {
        event.preventDefault();

        this.setState({
            loading: true,
            executeError: "",
            executeResult: {}
        });

        // To check id/conn validity.
        const currIDConn = getCurrIDConn();
        const formData = new FormData(event.target);

        // Connection   RequestConnection `json:"connection"`
        // Chaincode    api.Chaincode     `json:"chaincode"`
        // FunctionName string            `json:"functionName"`
        // Arguments    []string          `json:"arguments"`

        const reqBody = {
            connection: currIDConn,
            chaincode: {
                channelID: this.state.ccExecuteOption.channelID,
                name: this.state.ccExecuteOption.name
            },
            actionType: this.state.ccExecuteOption.actionType,
            functionName: formData.get("functionName"),
            // TODO use comma now as temp coding
            arguments: formData.get("arguments").split("\r\n"),
            targets: this.state.targets
        }

        let myHeaders = new Headers({
            "Accept": "application/json",
            "Content-Type": "application/octet-stream"
        });
        let request = new Request(CONST.getServiceURL("/chaincode/execute"), {
            method: "POST",
            mode: "cors",
            //credentials: "include",
            headers: myHeaders,
            body: JSON.stringify(reqBody),
        });

        fetch(request)
            .then(response => response.json())
            .then(result => {
                log.debug("ChaincodeExecute result", result);

                this.setState({
                    loading: false
                });

                if (result) {
                    if (result.resCode === 200) {
                        // if (result.chaincodeStatus === 200) {
                        this.setState({ executeResult: result });
                    }
                    else {
                        this.setState({ executeError: i18n("res_code") + result.resCode + ". " + result.errMsg });
                    }
                }
                else {
                    this.setState({ executeError: i18n("no_result_error") });
                }
            })
            .catch((err) => {
                log.error("Exception in fetching:", err);
                this.setState({
                    executeError: String(err),
                    loading: false
                });
            });
    }

    handleChangeTargets(event) {
        this.setState({ targets: event.target.value });
    };

    showExeResultRow(key, title, content) {
        const classes = this.state.classes;
        return (
            <TableRow key={key}>
                <TableCell align="right"><Typography noWrap className={classes.normalTitle}>{title}</Typography></TableCell>
                <TableCell align="left">
                    {
                        typeof(content) !== "object" ? (
                            <Typography gutterBottom className={classes.normalField}>{content}</Typography>
                        ) : content
                    }
                </TableCell>
            </TableRow>
        );
    }

    showExeResult() {
        const classes = this.state.classes;
        var content = undefined;
        const comp = this;
        if (this.state.executeError) {
            content = (
                <Grid item xs={12}>
                    <Typography color="error" className={classes.normalField}>{this.state.executeError}</Typography>
                </Grid>
                )
        }
        else if (this.state.executeResult.transactionID) {
            content = (
                <Grid item xs={12}>
                    <Table stickyHeader className={classes.table} size="small" aria-label="exe_result">
                        <TableBody key={"tbody_cc_exe_result"}>
                            {this.showExeResultRow("tx_id", i18n("transaction_id"), this.state.executeResult.transactionID)}
                            {this.showExeResultRow("tx_code", i18n("transaction_validation_code"), this.state.executeResult.txValidationCode)}
                            {this.showExeResultRow("statsu", i18n("chaicode_status"), this.state.executeResult.chaincodeStatus)}
                            {this.showExeResultRow("payload", i18n("chaincode_payload"), this.state.executeResult.payload)}
                            {this.showExeResultRow("peer_response", i18n("chaicode_peer_response"), (
                                <Box className={classes.fixedHeightBox}>
                                    <Table stickyHeader size="small" aria-label="exe_result_peer_response">
                                        <TableBody key={"tbody_cc_exe_result_peer"}>
                                            {
                                                (this.state.executeResult.peerResponses || []).map(resp => {
                                                    return (
                                                        <React.Fragment key={"resp_" + resp.endorser}>
                                                            {comp.showExeResultRow("resp_endorser", i18n("endorser"), resp.endorser)}
                                                            {comp.showExeResultRow("resp_version", i18n("version"), resp.version)}
                                                            {comp.showExeResultRow("resp_status", i18n("status"), resp.status)}
                                                            {comp.showExeResultRow("resp_payload", i18n("payload"), resp.payload)}
                                                        </React.Fragment>
                                                    );
                                                })
                                            }

                                        </TableBody>
                                    </Table>
                                </Box>
                            ))}
                        </TableBody>
                    </Table>
                </Grid>
            );
        }

        return (content ? <DialogContent>{content}</DialogContent> : null);
    }

    render() {
        if (!this.state.open) {
            return null;
        }

        const classes = this.props.classes;
        const isExecute = this.state.ccExecuteOption.actionType === "execute";

        return (
            <Dialog
                disableBackdropClick
                disableEscapeKeyDown
                maxWidth="md"
                fullWidth
                open={this.state.open}
                onClose={this.handleCancel}
            >
                <DialogTitle id="dialog_title">{i18n(isExecute ? "chaincode_execute" : "chaincode_query")}</DialogTitle>
                <form onSubmit={this.handleSubmit}>
                    <DialogContent>
                        <Typography className={classes.formField}>
                            {i18n("chaincode_execute_with",
                                i18n(this.state.ccExecuteOption.actionType),
                                this.state.ccExecuteOption.name,
                                this.state.ccExecuteOption.channelID)}
                        </Typography>
                    </DialogContent>

                    <DialogContent>
                        <Grid container spacing={2}>
                            <Grid item xs={4}>
                                <TextField
                                    fullWidth
                                    label={i18n("chaincode_function_name")}
                                    variant="outlined"
                                    required
                                    id="functionName"
                                    name="functionName"
                                    InputProps={{
                                        classes: { input: classes.formField }
                                    }}
                                />
                            </Grid>

                            <Grid item xs={8}>
                                <FormControl className={classes.formControl} fullWidth variant="outlined">
                                    <InputLabel id="target_label">{i18n("targets")} ({i18n("targets_auto")})</InputLabel>
                                    <Select
                                        //labelId="target_label"
                                        id="target"
                                        multiple
                                        value={this.state.targets}
                                        onChange={this.handleChangeTargets}
                                        input={<Input />}
                                        renderValue={selected => selected.join(', ')}
                                        MenuProps={MenuProps}
                                        style={{
                                            fontSize: 13,
                                        }}
                                        fullWidth
                                    >
                                        {
                                            (this.state.ccExecuteOption.targets || []).map(target => (
                                                <MenuItem value={target.URL} className={classes.formField} key={target.URL}>
                                                    <Checkbox size="small" checked={this.state.targets.indexOf(target.URL) > -1} />
                                                    <Typography className={classes.formField}>{target.URL}</Typography>
                                                </MenuItem>
                                            )
                                            )
                                        }
                                    </Select>
                                </FormControl>

                            </Grid>


                            <Grid item xs={12}>
                                <TextField
                                    fullWidth
                                    label={i18n("chaincode_arguments")}
                                    variant="outlined"
                                    id="arguments"
                                    name="arguments"
                                    multiline
                                    rows="4"
                                    InputProps={{
                                        classes: { input: classes.formField }
                                    }}
                                />
                                <FormHelperText id="component-helper-text">{i18n("chaincode_constructor_remark")}</FormHelperText>
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
                            <Grid item xs={8}>
                                <Fade
                                    in={this.state.loading}
                                    style={{
                                        transitionDelay: this.state.loading ? '100ms' : '0ms',
                                        width: 20,
                                        height: 20,
                                    }}
                                    unmountOnExit
                                >
                                    <CircularProgress style={{ marginLeft: 60 }} />
                                </Fade>
                            </Grid>

                            <Grid item xs={4}>
                                <Box display="flex" flexDirection="row-reverse">
                                    <Button
                                        type="submit"
                                        autoFocus
                                        variant="contained"
                                        color="primary"
                                        style={{ marginLeft: "auto" }}
                                        disabled={this.state.loading}>
                                        {i18n(isExecute ? "chaincode_execute" : "chaincode_query")}
                                    </Button>
                                    &nbsp;
                                    <Button
                                        onClick={this.handleCancel}
                                        variant="contained"
                                        color="primary"
                                        disabled={this.state.loading}>
                                        {i18n("cancel")}
                                    </Button>
                                </Box>
                            </Grid>
                        </Grid>

                    </DialogActions>

                    {this.showExeResult()}

                </form>

            </Dialog>
        );
    }
}

export default withStyles(styles)(ChaincodeExecute);