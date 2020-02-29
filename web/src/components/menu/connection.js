import React from "react";
import i18n from "../../i18n";
import InputLabel from "@material-ui/core/InputLabel";
import MenuItem from "@material-ui/core/MenuItem";
import Grid from "@material-ui/core/Grid";
import Select from "@material-ui/core/Select";
import Tooltip from "@material-ui/core/Tooltip";
import IconButton from "@material-ui/core/IconButton";
import Typography from "@material-ui/core/Typography";
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import DialogContentText from '@material-ui/core/DialogContentText';
import Button from "@material-ui/core/Button";
import AddIcon from "@material-ui/icons/Add";
import ClearIcon from "@material-ui/icons/Clear";
import { log } from "../../common/log";
import { withStyles } from "@material-ui/core/styles";
import { clientStorage } from "../../data/localstore";
import { refreshNetwork } from "../../common/utils";

const styles = theme => ({
    normalText: {
        fontSize: 13
    },
    selectEmpty: {
        marginTop: theme.spacing(2),
    },
});

/**
 * 
 */
class ConnectionMenu extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            ...props,
            //onOpenProfile: props.onOpenProfile

            connProfiles: [],
            currConnProfileName: "",
            identities: [],
            currIDLabel: "",
            connected: false,

            openConfirm: false,
            confirmedAction: null

        };

        this.handleChangeConnProfile = this.handleChangeConnProfile.bind(this);
        this.handleChangeID = this.handleChangeID.bind(this);
        this.handleConnectTo = this.handleConnectTo.bind(this);
        this.handleClearProfile = this.handleClearProfile.bind(this);
        this.handleClearIdentity = this.handleClearIdentity.bind(this);
        this.handleCancelConfirm = this.handleCancelConfirm.bind(this);
        this.handleDeleteConfirm = this.handleDeleteConfirm.bind(this);
        
        this.__ConnProfileName = "";
        this.__IDLabel = "";

        log.debug("ConnectionMenu: constructor");
    }

    updateItems() {
        const connProfiles = clientStorage.loadConnectionProfiles(); //loadConnectionProfiles();
        let currConnProfileName = clientStorage.getCurrConnProfileName();
        if (!currConnProfileName && connProfiles && connProfiles.length > 0) {
            currConnProfileName = connProfiles[0].name;
            clientStorage.setCurrConnProfileName(currConnProfileName);
        }

        const identities = clientStorage.loadIdentities();
        let currIDLabel = clientStorage.getCurrIDLabel();
        if (!currIDLabel && identities && identities.length > 0) {
            currIDLabel = identities[0].label;
            clientStorage.setCurrIDLabel(currIDLabel);
        }

        this.setState({ currConnProfileName, connProfiles, currIDLabel, identities, connected: true });

        log.debug("ConnectionMenu: updateItems");
    }

    componentDidMount() {
        this.updateItems();
    }

    componentDidUpdate() {
    }

    componentWillUnmount() {
    }

    // TODO has issue yet, when change another select box.
    handleChangeConnProfile(event) {
        this.setState({ currConnProfileName: event.target.value });
        this.setState({
            connected: event.target.value === clientStorage.getCurrConnProfileName()
                && this.state.currIDLabel === clientStorage.getCurrIDLabel()
        });
        log.debug("ConnectionMenu: handleChangeConnProfile ", event.target.value);
    };

    handleChangeID(event) {
        this.setState({ currIDLabel: event.target.value });
        this.setState({
            connected: event.target.value === clientStorage.getCurrIDLabel()
                && this.state.currConnProfileName === clientStorage.getCurrConnProfileName()
        });
        log.debug("ConnectionMenu: handleChangeID ", event.target.value);
    };

    handleConnectTo() {
        clientStorage.setCurrConnProfileName(this.state.currConnProfileName);
        clientStorage.setCurrIDLabel(this.state.currIDLabel);
        this.setState({ connected: true });

        log.debug("ConnectionMenu: connect to", this.state.currConnProfileName, this.state.currIDLabel);

        window.location.reload();
    }

    handleClearProfile(name) {
        const comp = this;
        return function () {
            comp.setState({
                openConfirm: true,
                confirmedAction: function () {
                    clientStorage.removeConnectionProfile(name);
                    window.location.reload();
                }
            });
        }
    }

    handleClearIdentity(label) {
        const comp = this;
        return function () {
            comp.setState({
                openConfirm: true,
                confirmedAction: function () {
                    clientStorage.removeIdentity(label);
                    window.location.reload();
                }
            });
        }
    }

    handleCancelConfirm() {
        this.setState({openConfirm: false});
    }

    handleDeleteConfirm() {
        if (this.state.confirmedAction) {
            this.state.confirmedAction();
        }
    }

    render() {
        const classes = this.props.classes;

        return (
            <React.Fragment>
                      
                <Grid container>
                    <Grid item xs={12} style={{ paddingLeft: 25, paddingBottom: 20 }}>
                        <InputLabel id="connProfileValue" className={classes.normalText}>{i18n("connection_profile")}</InputLabel>
                        <Select
                            id="connProfileSelect"
                            value={this.state.currConnProfileName || ""}
                            onChange={this.handleChangeConnProfile}
                            autoWidth
                            style={{ width: 180, fontSize: 13, }}
                            renderValue={selected => selected}
                        >
                            {
                                (this.state.connProfiles || []).map(p => {
                                    return (
                                        <MenuItem value={p.name} key={p.name}>
                                            <ClearIcon fontSize="inherit" color="primary" style={{marginRight: 10}} onClick={this.handleClearProfile(p.name)} />
                                            <Typography className={classes.normalText}> {p.name} </Typography>
                                        </MenuItem>
                                    );
                                })
                            }
                        </Select>
                        <Tooltip title={i18n("connection_profile_add")}>
                            <IconButton aria-label={i18n("connection_profile_add")} className={classes.margin} size="small" onClick={this.state.onOpenProfile(true, false)}>
                                <AddIcon fontSize="inherit" color="primary" />
                            </IconButton>
                        </Tooltip>
                    </Grid>


                    <Grid item xs={12} style={{ paddingLeft: 25 }}>
                        <InputLabel id="identityValue" className={classes.normalText}>{i18n("Identity")}</InputLabel>
                        <Select
                            id="idSelect"
                            value={this.state.currIDLabel || ""}
                            onChange={this.handleChangeID}
                            autoWidth
                            style={{ width: 180, fontSize: 13 }}
                            renderValue={selected => selected}
                        >
                            {
                                (this.state.identities || []).map(id => {
                                    return (
                                        <MenuItem value={id.label} key={id.label}>
                                            <ClearIcon fontSize="inherit" color="primary" style={{marginRight: 10}} onClick={this.handleClearIdentity(id.label)} />
                                            <Typography className={classes.normalText}> {id.label} </Typography>
                                        </MenuItem>
                                    );
                                })
                            }
                        </Select>
                        <Tooltip title={i18n("identity_add")}>
                            <IconButton aria-label={i18n("identity_add")} className={classes.margin} size="small" onClick={this.state.onOpenProfile(false, true)}>
                                <AddIcon fontSize="inherit" color="primary" />
                            </IconButton>
                        </Tooltip>
                    </Grid>

                    <Grid item xs={12} style={{ paddingLeft: 25, paddingTop: 10 }}>
                        <Button
                            onClick={this.state.connected ? refreshNetwork : this.handleConnectTo}
                            variant="contained"
                            color="primary"
                            size="small"
                        >
                            {i18n(this.state.connected ? "refresh_connection" : "connect_to")}
                        </Button>
                    </Grid>
                </Grid>

                {this.state.openConfirm ? (
                    <Dialog
                        open={this.state.openConfirm}
                        disableBackdropClick
                        disableEscapeKeyDown
                        onClose={this.handleCancelConfirm}
                        aria-labelledby="alert-dialog-title"
                        aria-describedby="alert-dialog-description"
                    >
                        <DialogTitle id="alert-dialog-title">{i18n("delete_confirmation")}</DialogTitle>
                        <DialogContent>
                            <DialogContentText id="alert-dialog-description">
                                {i18n("delete_confirmation_remark")}
                            </DialogContentText>
                        </DialogContent>
                        <DialogActions>
                            <Button
                                onClick={this.handleCancelConfirm}
                                variant="contained"
                                color="primary">
                                {i18n("cancel")}
                            </Button>
                            <Button
                                onClick={this.handleDeleteConfirm}
                                autoFocus
                                variant="contained"
                                color="primary"
                                style={{ marginLeft: "auto" }}>
                                {i18n("delete")}
                            </Button>
                        </DialogActions>
                    </Dialog>

                ) : null}                

            </React.Fragment>
            
        );
    }

}

export default withStyles(styles)(ConnectionMenu);