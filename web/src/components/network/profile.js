import React from 'react';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import { withStyles } from '@material-ui/core/styles';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import Typography from '@material-ui/core/Typography';
import Grid from '@material-ui/core/Grid';
import { log } from "../../common/log";
import i18n from '../../i18n';
import { clientStorage } from "../../data/localstore";

const styles = (theme) => ({
    normalText: {
        fontSize: 13
    },
    formField: {
        fontSize: 13,
        marginLeft: "auto",
    },
    hiddenInput: {
        display: 'none',
    },
    selectedFile: {
        fontSize: 13,
        fontWeight: 600,
        marginLeft: 10,
        marginTop: 10
    }
});

class Profile extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            ...props,
            // open: props.open,
            // needConnProf: props.needConnProf,
            // needIdentity: props.needIdentity,
            // onClose: props.onClose,  // to notice parent to close the profile
            // callBack: props.callBack,    // to notice parent after configured
            selectedFiles: {},
            tempForm: {},
        };

        this.handleCancel = this.handleCancel.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
        this.handleConfigured = this.handleConfigured.bind(this);
        this.onSelectFile = this.onSelectFile.bind(this);

        log.debug("Profile: constructor", this.state.open);
    }

    componentDidMount() {
    }

    componentDidUpdate() {
    }

    componentWillUnmount() {
    }

    onSelectFile(event) {
        // TODO To support Safari
        const fileReader = new FileReader();
        const tempForm = this.state.tempForm;
        const fieldName = event.target.name;
        const f = event.target.files[0];
        if (!f) {
            return;
        }
        fileReader.readAsText(f);
        fileReader.onload = (e) => {
            tempForm[fieldName] = {
                path: f.name,
                content: e.target.result
            };

            const selectedFiles = this.state.selectedFiles;
            selectedFiles[fieldName] = f.name;
            this.setState({ selectedFiles: selectedFiles });

            log.info(this.state.selectedFiles);
            log.info(tempForm);
        };
    };

    handleCancel() {
        this.setState({ open: false });
        // It is required to update open status in parent peer_overview component.
        this.state.onClose();
    }

    handleConfigured() {
        this.setState({ open: false });
        this.state.onClose();
        // Callback from parent, to update the peer overview.
        this.state.callBack();
    }


    handleSubmit(event) {
        event.preventDefault();

        if (this.state.needConnProf && this.state.tempForm["connProfile"]) {
            const connProfile = this.state.tempForm["connProfile"];
            if (connProfile.path && connProfile.content) {
                clientStorage.saveConnectionProfile({
                    name: connProfile.path,
                    path: connProfile.path,
                    content: connProfile.content
                });
            }
        }

        const formData = new FormData(event.target);
        if (this.state.needIdentity && this.state.tempForm["idCertificate"] && this.state.tempForm["idPrivateKey"]) {
            const idCert = this.state.tempForm["idCertificate"];
            const idPrvKey = this.state.tempForm["idPrivateKey"];
            if (idCert.path && idCert.content && idPrvKey.path && idPrvKey.content) {
                clientStorage.saveIdentity({
                    label: idCert.path,
                    MSPID: formData.get("MSPID"),
                    certPath: idCert.path,
                    certContent: idCert.content,
                    prvKeyPath: idPrvKey.path,
                    prvKeyContent: idPrvKey.content
                })
            }
        }

        this.handleConfigured();

        window.location.reload();
    }


    getConnectionProfileSetting(classes) {
        return (
            <Grid container spacing={2}>
                {
                    this.state.needConnProf ?
                        (<Grid item xs={12}>
                            <Typography className={classes.normalText}>{i18n("add_one_conn_profile")}</Typography>
                        </Grid>)
                        : null
                }

                <Grid item xs={5}>

                    <input
                        required
                        accept="*"
                        id="connProfile"
                        name="connProfile"
                        className={classes.hiddenInput}
                        onChange={this.onSelectFile}
                        type="file"
                    />
                    <label htmlFor="connProfile">
                        <Button variant="contained" component="span" fullWidth>
                            {i18n("connection_profile_upload")}
                        </Button>
                    </label>
                    <Typography className={classes.selectedFile} gutterBottom>{this.state.selectedFiles["connProfile"]}</Typography>
                </Grid>
                <Grid item xs={7}>
                    <Typography className={classes.normalText}>{i18n("example_from_document")}</Typography>
                </Grid>
            </Grid>
        );
    }

    getIdentitySetting(classes) {
        return (
            <Grid container spacing={2}>
                {
                    this.state.needIdentity ?
                        (<Grid item xs={12}>
                            <Typography className={classes.normalText}>{i18n("add_one_identity")}</Typography>
                        </Grid>)
                        : null
                }
                <Grid item xs={5}>
                    <TextField
                        label={i18n("msp_id")}
                        variant="outlined"
                        fullWidth
                        required
                        id="MSPID"
                        name="MSPID"
                        InputProps={{
                            classes: { input: classes.formField }
                        }}
                    />
                </Grid>
                <Grid item xs={7}></Grid>


                <Grid item xs={5}>
                    <input
                        required
                        accept="*"
                        id="idCertificate"
                        name="idCertificate"
                        className={classes.hiddenInput}
                        onChange={this.onSelectFile}
                        type="file"
                    />
                    <label htmlFor="idCertificate">
                        <Button variant="contained" component="span" fullWidth>
                            {i18n("id_certificate_upload")}
                        </Button>
                    </label>
                    <Typography className={classes.selectedFile} gutterBottom>{this.state.selectedFiles["idCertificate"]}</Typography>
                </Grid>
                <Grid item xs={7}>
                    <Typography className={classes.normalText}>{i18n("example_from_document")}</Typography>
                </Grid>

                <Grid item xs={5}>
                    <input
                        required
                        accept="*"
                        id="idPrivateKey"
                        name="idPrivateKey"
                        className={classes.hiddenInput}
                        onChange={this.onSelectFile}
                        type="file"
                    />
                    <label htmlFor="idPrivateKey">
                        <Button variant="contained" component="span" fullWidth>
                            {i18n("id_private_key_upload")}
                        </Button>
                    </label>
                    <Typography className={classes.selectedFile} gutterBottom>{this.state.selectedFiles["idPrivateKey"]}</Typography>
                </Grid>
                <Grid item xs={7}>
                    <Typography className={classes.normalText}>{i18n("example_from_document")}</Typography>
                </Grid>

            </Grid>);
    }

    render() {
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
                <DialogTitle id="dialog_title">{i18n("profile_setting")}</DialogTitle>
                <form onSubmit={this.handleSubmit}>
                    <DialogContent>
                        {
                            this.state.needConnProf ? this.getConnectionProfileSetting(classes) : null
                        }
                    </DialogContent>
                    <DialogContent>
                        {
                            this.state.needIdentity ? this.getIdentitySetting(classes) : null
                        }
                    </DialogContent>
                    <DialogActions>
                        <Grid
                            container
                            direction="row"
                            justify="flex-end"
                            alignItems="center"
                        >
                            <Grid>
                                <Button
                                    onClick={this.handleCancel}
                                    variant="contained"
                                    color="primary"
                                    disabled={this.state.loading}>
                                    {i18n("cancel")}
                                </Button>
                                &nbsp;
                                <Button
                                    type="submit"
                                    variant="contained"
                                    color="primary"
                                    style={{ marginLeft: "auto" }}
                                    disabled={this.state.loading}>
                                    {i18n("profile_setup")}
                                </Button>
                            </Grid>
                        </Grid>
                    </DialogActions>
                </form>
            </Dialog>
        );
    }
}

export default withStyles(styles)(Profile);