import React from 'react';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import { withStyles } from '@material-ui/core/styles';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import Typography from '@material-ui/core/Typography';
import FormControl from '@material-ui/core/FormControl';
import InputLabel from '@material-ui/core/InputLabel';
import Grid from '@material-ui/core/Grid';
import Box from "@material-ui/core/Box";
import Select from '@material-ui/core/Select';
import MenuItem from '@material-ui/core/MenuItem';
import FormHelperText from '@material-ui/core/FormHelperText';
import Fade from '@material-ui/core/Fade';
import CircularProgress from '@material-ui/core/CircularProgress';
import { getCurrIDConn, outputServiceErrorResult } from "../../common/utils";
import { log } from "../../common/log";
import i18n from '../../i18n';
import * as CONST from "../../common/constants";

const styles = (theme) => ({
    formField: {
        fontSize: 13,
        marginLeft: "auto",

    }
});

class ChaincodeInstantiate extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            ...props,
            // open: props.open,
            // onClose: props.onClose,
            // callBack: props.callBack,
            // ccInstantiateOption: props.ccInstantiateOption,
            /////////////////////////////////////////////////
            // existInstantiatedCC,
            // name: name,
            // version: version,
            // path: path,
            // policy: policy,
            // constructor: constructor,
            // channelID: channelID,
            // peer: peer,
            // orderer: orderer,
            // channels
            // channelOrderers

            orderersOfChannel: [],
            instantiateError: "",
            loading: false,
        };

        if (!this.state.ccInstantiateOption) {
            this.state.ccInstantiateOption = {};
        }

        const ccOption = this.state.ccInstantiateOption;

        if (!ccOption.channelID && ccOption.peer && ccOption.peer.channels
            && ccOption.peer.channels.length > 0) {
            ccOption.channelID = ccOption.peer.channels[0];
        }

        if (!ccOption.orderer && ccOption.channelOrderers && ccOption.channelOrderers[ccOption.channelID]
            && ccOption.channelOrderers[ccOption.channelID].length > 0) {
            ccOption.orderer = ccOption.channelOrderers[ccOption.channelID][0];
        }

        this.state.orderersOfChannel = (ccOption.channelOrderers || {})[ccOption.channelID];

        this.handleCancel = this.handleCancel.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
        this.handleInstantiate = this.handleInstantiate.bind(this);
        this.handleChangeChannelID = this.handleChangeChannelID.bind(this);

        log.debug("ChaincodeInstantiate: constructor");
    }

    componentDidMount() {
    }

    componentDidUpdate() {
    }

    componentWillUnmount() {
    }

    handleCancel() {
        this.setState({ open: false });
        // It is required to update open status in parent peer_overview component.
        this.state.onClose();
    }

    // TODO to change to handleInstantiated
    handleInstantiate() {
        this.setState({ open: false });
        this.state.onClose();
        // Callback from parent, to update the peer overview.
        this.state.callBack();
    }

    handleChangeChannelID(event) {
        this.setState({
            ccInstantiateOption: { ...this.state.ccInstantiateOption, channelID: event.target.value },
            orderersOfChannel: (this.state.ccInstantiateOption.channelOrderers || {})[event.target.value]
        });
    }

    handleSubmit(event) {
        event.preventDefault();

        this.setState({
            instantiateError: "",
            loading: true
        });

        // To check id/conn validity.
        const currIDConn = getCurrIDConn();
        const formData = new FormData(event.target);

        const reqBody = {
            connection: currIDConn,
            chaincode: {
                name: formData.get("name"),
                version: formData.get("version"),
                path: formData.get("path"),
                policy: formData.get("policy"),
                // TODO a strong arguments splitter
                constructor: formData.get("constructor").split(","),
                channelID: formData.get("channelID"),
            },
            target: formData.get("target"),
            orderer: formData.get("orderer"),
            isUpgrade: !!this.state.ccInstantiateOption.existInstantiatedCC
        };

        log.log("Instantiate chaincode: ", reqBody);

        let myHeaders = new Headers({
            'Accept': 'application/json',
            'Content-Type': 'application/octet-stream'
        });
        let request = new Request(CONST.getServiceURL(this.state.ccInstantiateOption.existInstantiatedCC ? "/chaincode/upgrade" : "/chaincode/instantiate"), {
            method: 'POST',
            mode: 'cors',
            //credentials: 'include',
            headers: myHeaders,
            body: JSON.stringify(reqBody),
        });

        fetch(request)
            .then(response => response.json())
            .then(result => {
                log.debug("Chaincode instantiation result", result);

                this.setState({
                    loading: false
                });

                if (result) {
                    if (result.resCode === 200) {
                        this.handleInstantiate();
                    }
                    else {
                        this.setState({ instantiateError: outputServiceErrorResult(result) });
                    }
                }
                else {
                    this.setState({ instantiateError: i18n("no_result_error") });
                }
            })
            .catch(function (err) {
                this.setState({
                    instantiateError: String(err),
                    loading: false
                });
            });
    }

    render() {
        if (!this.state.open) {
            return null;
        }

        const classes = this.props.classes;
        const ccOption = this.state.ccInstantiateOption;

        return (
            <Dialog
                disableBackdropClick
                disableEscapeKeyDown
                maxWidth="md"
                open={this.state.open}
                onClose={this.handleCancel}
            >
                <DialogTitle id="dialog_title">{i18n(ccOption.existInstantiatedCC ? "chaincode_upgrade" : "chaincode_instantiate")}</DialogTitle>
                <form onSubmit={this.handleSubmit}>
                    <DialogContent>
                        <Grid container spacing={2}>
                            <Grid item xs={4}>
                                <TextField
                                    fullWidth
                                    label={i18n("chaincode_name")}
                                    variant="outlined"
                                    required
                                    id="name"
                                    name="name"
                                    defaultValue={ccOption.name}
                                    InputProps={{
                                        readOnly: true,
                                        classes: { input: classes.formField }
                                    }}
                                />
                            </Grid>

                            <Grid item xs={4}>
                                <TextField
                                    fullWidth
                                    label={i18n("chaincode_version")}
                                    variant="outlined"
                                    required
                                    id="version"
                                    name="version"
                                    defaultValue={ccOption.version}
                                    InputProps={{
                                        readOnly: true,
                                        classes: { input: classes.formField }
                                    }}
                                />
                            </Grid>

                            <Grid item xs={4}>
                                <TextField
                                    fullWidth
                                    label={i18n("chaincode_path")}
                                    variant="outlined"
                                    required
                                    id="path"
                                    name="path"
                                    defaultValue={ccOption.path}
                                    InputProps={{
                                        readOnly: true,
                                        classes: { input: classes.formField }
                                    }}
                                />
                            </Grid>

                            <Grid item xs={4}>
                                <TextField
                                    fullWidth
                                    label={i18n("target")}
                                    variant="outlined"
                                    required
                                    id="target"
                                    name="target"
                                    // TODO donot allow target selection
                                    defaultValue={(ccOption.peer || {}).URL}
                                    InputProps={{
                                        readOnly: true,
                                        classes: { input: classes.formField }
                                    }}
                                />
                            </Grid>

                            <Grid item xs={4}>
                                <TextField
                                    fullWidth
                                    label={i18n("chaincode_policy")}
                                    variant="outlined"
                                    required
                                    id="policy"
                                    name="policy"
                                    defaultValue={ccOption.policy}
                                    InputProps={{
                                        classes: { input: classes.formField }
                                    }}
                                />
                                <FormHelperText id="component-helper-text">{i18n("chaincode_policy_remark")}</FormHelperText>
                            </Grid>

                            <Grid item xs={4}>
                            </Grid>

                            <Grid item xs={4}>
                                <FormControl className={classes.formControl}>
                                    <InputLabel id="channel_label">{i18n("channel")}</InputLabel>
                                    <Select
                                        fullWidth
                                        required
                                        id="channelID"
                                        name="channelID"
                                        value={ccOption.channelID}
                                        onChange={this.handleChangeChannelID}
                                        style={{ fontSize: 13, width: 280 }}
                                        readOnly={!!ccOption.existInstantiatedCC}
                                    >
                                        {
                                            (ccOption.peer.channels || []).map(chID => {
                                                return (<MenuItem value={chID} key={chID} className={classes.formField}>{chID}</MenuItem>);
                                            })
                                        }
                                    </Select>
                                </FormControl>
                            </Grid>

                            <Grid item xs={4}>
                                <FormControl className={classes.formControl}>
                                    <InputLabel id="orderer_label">{i18n("orderer")}</InputLabel>
                                    <Select
                                        fullWidth
                                        required
                                        id="orderer"
                                        name="orderer"
                                        value={(ccOption.orderer || {}).URL}
                                        style={{ fontSize: 13, width: 280 }}
                                    >
                                        {
                                            (this.state.orderersOfChannel || []).map(orderer => {
                                                return (<MenuItem value={orderer.URL} key={orderer.URL} className={classes.formField}>{orderer.URL}</MenuItem>);
                                            })
                                        }
                                    </Select>
                                </FormControl>
                            </Grid>

                            <Grid item xs={4}>
                            </Grid>

                            <Grid item xs={12}>
                                <TextField
                                    fullWidth
                                    label={i18n("chaincode_constructor")}
                                    variant="outlined"
                                    id="constructor"
                                    name="constructor"
                                    multiline
                                    rows="4"
                                    defaultValue={ccOption.constructor}
                                    InputProps={{
                                        classes: { input: classes.formField }
                                    }}
                                />
                                <FormHelperText id="component-helper-text">{i18n("chaincode_constructor_remark")}</FormHelperText>
                            </Grid>

                        </Grid>
                    </DialogContent>
                    <DialogActions>
                        <Grid container spacing={2} alignItems="flex-end">
                            <Grid item xs={9}>
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

                            <Grid item xs={3}>
                                <Box display="flex" flexDirection="row-reverse">
                                    <Button
                                        type="submit"
                                        autoFocus
                                        variant="contained"
                                        color="primary"
                                        style={{ marginLeft: "auto" }}
                                        disabled={this.state.loading}>
                                        {i18n(this.state.ccInstantiateOption.existInstantiatedCC ? "upgrade" : "instantiate")}
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

                    {
                        this.state.instantiateError ? (
                            <Grid item xs={12}>
                                <Typography color="error" style={{ marginRight: "auto", fontSize: 11 }}>{this.state.instantiateError}</Typography>
                            </Grid>
                        ) : null
                    }

                </form>
            </Dialog>
        );
    }
}

export default withStyles(styles)(ChaincodeInstantiate);