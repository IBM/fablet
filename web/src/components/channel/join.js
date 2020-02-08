import React from 'react';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import Typography from '@material-ui/core/Typography';
import { withStyles } from '@material-ui/core/styles';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import Grid from '@material-ui/core/Grid';
import Box from '@material-ui/core/Box';
import Fade from '@material-ui/core/Fade';
import CircularProgress from '@material-ui/core/CircularProgress';
import { getCurrIDConn } from "../../common/utils";
import { log } from "../../common/log";
import i18n from '../../i18n';
import * as CONST from "../../common/constants";

const styles = (theme) => ({
    formField: {
        fontSize: 13,
    },
});

class ChannelJoin extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            ...props,
            // callBack: props.callBack,
            // onClose
            // channelJoinOption
            //     peer

            joinError: "",
            loading: false,
        };

        this.handleCancel = this.handleCancel.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
        this.handleJoined = this.handleJoined.bind(this);
        log.debug("ChannelJoin: constructor");
    }

    componentDidMount() {
    }

    componentDidUpdate() {
    }

    componentWillUnmount() {
    }

    handleCancel() {
        this.setState({ open: false });
        this.state.onClose();
    }

    // TODO change to handleInstalled
    handleJoined() {
        this.setState({ open: false });
        this.state.onClose();
        // Callback from parent, to update the peer overview.
        this.state.callBack();
    }

    handleSubmit(event) {
        event.preventDefault();

        this.setState({
            loading: true,
            joinError: ""
        });

        // To check id/conn validity.
        const currIDConn = getCurrIDConn();
        const formData = new FormData(event.target);
        const reqBody = {
            connection: currIDConn,
            channelID: formData.get("channelID"),
            targets: [this.state.channelJoinOption.peer.name],
            orderer: formData.get("orderer"),
        };

        let myHeaders = new Headers({
            'Accept': 'application/json',
            'Content-Type': 'application/octet-stream'
        });
        let request = new Request(CONST.getServiceURL("/channel/join"), {
            method: 'POST',
            mode: 'cors',
            //credentials: 'include',
            headers: myHeaders,
            body: JSON.stringify(reqBody),
        });

        fetch(request)
            .then(response => response.json())
            .then(result => {
                log.debug("ChannelJoin result", result);

                this.setState({
                    loading: false
                });

                if (result) {
                    if (result.resCode === CONST.RES_OK) {
                        this.handleJoined();
                    }
                    else {
                        this.setState({ joinError: i18n("res_code") + result.resCode + ". " + result.errMsg });
                    }
                }
            })
            .catch(function (err) {
                log.error("Exception in joining:", err);
                this.setState({
                    joinError: String(err),
                    loading: false
                });
            });
    }

    render() {
        if (!this.state.open) {
            return null;
        }

        const classes = this.props.classes;

        return (
            <Dialog
                disableBackdropClick
                disableEscapeKeyDown
                maxWidth="md"
                fullWidth
                open={this.state.open}
                onClose={this.handleCancel}
            >
                <DialogTitle id="dialog_title">{i18n("channel_join")}</DialogTitle>
                <form onSubmit={this.handleSubmit}>
                    <DialogContent>
                        <Grid container spacing={2}>

                            <Grid item xs={4}>
                                <TextField
                                    fullWidth
                                    label={i18n("peer")}
                                    variant="outlined"
                                    required
                                    id="peer"
                                    name="peer"
                                    value={this.state.channelJoinOption.peer.name}
                                    readOnly
                                    InputProps={{
                                        classes: { input: classes.formField }
                                    }}
                                />
                            </Grid>

                            <Grid item xs={4}>
                                <TextField
                                    fullWidth
                                    label={i18n("channel")}
                                    variant="outlined"
                                    required
                                    id="channelID"
                                    name="channelID"
                                    InputProps={{
                                        classes: { input: classes.formField }
                                    }}
                                />
                            </Grid>

                            <Grid item xs={4}>
                                <TextField
                                    fullWidth
                                    label={i18n("orderer")}
                                    variant="outlined"
                                    required
                                    id="orderer"
                                    name="orderer"
                                    InputProps={{
                                        classes: { input: classes.formField }
                                    }}
                                />
                            </Grid>



                        </Grid>
                    </DialogContent>
                    <DialogActions>
                        <Grid container spacing={2}>
                            <Grid item xs={10}>
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
                            <Grid item xs={2}>
                                <Box display="flex" flexDirection="row-reverse">
                                    <Button
                                        type="submit"
                                        autoFocus
                                        variant="contained"
                                        color="primary"
                                        style={{ marginLeft: "auto" }}
                                        disabled={this.state.loading}>
                                        {i18n("join")}
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
                        this.state.joinError ? (
                            <DialogContent>
                                <Grid item xs={12}>
                                <Typography color="error" style={{ marginRight: "auto", fontSize: 11 }}>{this.state.joinError}</Typography>
                            </Grid>
                            </DialogContent>
                        ) : null
                    }

                </form>
            </Dialog>
        );
    }
}

export default withStyles(styles)(ChannelJoin);