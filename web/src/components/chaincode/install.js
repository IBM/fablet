import React from 'react';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import Typography from '@material-ui/core/Typography';
import { withStyles } from '@material-ui/core/styles';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import FormControl from '@material-ui/core/FormControl';
import InputLabel from '@material-ui/core/InputLabel';
import Grid from '@material-ui/core/Grid';
import Select from '@material-ui/core/Select';
import MenuItem from '@material-ui/core/MenuItem';
import Fade from '@material-ui/core/Fade';
import CircularProgress from '@material-ui/core/CircularProgress';
import { getCurrIDConn, outputServiceErrorResult } from "../../common/utils";
import { log } from "../../common/log";
import i18n from '../../i18n';
import * as CONST from "../../common/constants";

const styles = (theme) => ({
    formField: {
        fontSize: 13,
    },
});

class ChaincodeInstall extends React.Component {
    constructor(props) {
        super(props);
        this.state = { ...props, 
            // TODO to use target name if configured.
            ////////////////////////////////////////
            // callBack: props.callBack,
            // ccInstallOption
            //     name, version, type, path, target

            chaincodeType: (props.ccInstallOption && props.ccInstallOption.type) || "golang", 

            packageFile: "", 
            installError: "",
            loading: false,
        };

        this.handleCancel = this.handleCancel.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
        this.handleInstall = this.handleInstall.bind(this);
        this.onSelectFile = this.onSelectFile.bind(this);
        this.handleChangeChaicodeType = this.handleChangeChaicodeType.bind(this);
        this.tempForm = {};
        log.debug("ChaincodeInstall: constructor");
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
    handleInstall() {
        this.setState({ open: false });
        this.state.onClose();
        // Callback from parent, to update the peer overview.
        this.state.callBack();
    }

    onSelectFile(event) {
        // TODO To support Safari
        const fileReader = new FileReader();
        const tempForm = this.tempForm;
        const fieldName = event.target.name;
        const f = event.target.files[0];
        if (!f) {
            return;
        }

        this.setState({ packageFile: f.name });

        fileReader.readAsBinaryString(f);
        fileReader.onload = (e) => {
            tempForm[fieldName] = {
                path: f.name,
                // Encode the binary content to bse64.
                content: btoa(e.target.result)
            };

            log.debug("Selected file", tempForm[fieldName].path, tempForm[fieldName].content.length);
        };
    }

    handleChangeChaicodeType(event) {
        this.setState({ chaincodeType: event.target.value });
    }

    handleSubmit(event) {
        event.preventDefault();

        this.setState({
            loading: true,
            installError: ""
        });

        // To check id/conn validity.
        const currIDConn = getCurrIDConn();
        const tempForm = this.tempForm;
        const formData = new FormData(event.target);
        const reqBody = {
            connection: currIDConn,
            chaincode: {
                name: formData.get("name"),
                version: formData.get("version"),
                type: this.state.chaincodeType,
                path: formData.get("path"),
                package: tempForm["package"].content
            },
            targets: [(this.state.ccInstallOption.target||{}).URL]
        }

        let myHeaders = new Headers({
            'Accept': 'application/json',
            'Content-Type': 'application/octet-stream'
        });
        let request = new Request(CONST.getServiceURL("/chaincode/install"), {
            method: 'POST',
            mode: 'cors',
            //credentials: 'include',
            headers: myHeaders,
            body: JSON.stringify(reqBody),
        });

        fetch(request)
            .then(response => response.json())
            .then(result => {
                log.debug("ChaincodeInstall result", result);

                this.setState({
                    loading: false
                });

                if (result) {                    
                    if (result.resCode === 200) {
                        if (result.installRes) {
                            var allSuccess = true;
                            for (var peer in result.installRes) {
                                const peerRes = result.installRes[peer];
                                if (peerRes.code !== 0) {
                                    this.setState({ installError: this.state.installError + "  " + peerRes.message });
                                    allSuccess = false;
                                }
                            }
                            if (allSuccess) {
                                this.handleInstall();
                            }
                        }
                        else {
                            this.setState({ installError: i18n("chaincode_install_no_result") });
                        }
                    }
                    else {
                        this.setState({ installError: outputServiceErrorResult(result) });
                    }
                }
                else {
                    this.setState({ installError: i18n("no_result_error") });
                }
            })
            .catch(function(err) {
                log.error("Exception in fetching:", err);
                this.setState({
                    installError: String(err),
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
                open={this.state.open}
                onClose={this.handleCancel}
            >
                <DialogTitle id="dialog_title">{i18n("chaincode_install")}</DialogTitle>
                <form onSubmit={this.handleSubmit}>
                    <DialogContent>
                        <Typography className={classes.formField}>
                            {i18n("chaincode_install_to", (this.state.ccInstallOption.target||{}).URL)}
                        </Typography>
                    </DialogContent>
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
                                    defaultValue={(this.state.ccInstallOption && this.state.ccInstallOption.name) || ""}
                                    InputProps={{
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
                                    defaultValue={(this.state.ccInstallOption && this.state.ccInstallOption.version) || ""}
                                    InputProps={{
                                        classes: { input: classes.formField }
                                    }}
                                />
                            </Grid>

                            <Grid item xs={4}>
                                <FormControl className={classes.formControl}>
                                    <InputLabel id="type_label">{i18n("chaincode_type")}</InputLabel>
                                    <Select
                                        fullWidth
                                        id="type"
                                        value={this.state.chaincodeType}
                                        onChange={this.handleChangeChaicodeType}                                        
                                        style={{
                                            fontSize:13,
                                            width: 280
                                        }}
                                    >
                                        {
                                            CONST.CHAINCODE_TYPES.map(ccType=> {
                                                return (
                                                    <MenuItem value={ccType} className={classes.formField} key={ccType}>
                                                        {i18n(`chaincode_type_${ccType}`)}
                                                    </MenuItem>
                                                );
                                            })
                                        }
                                        
                                    </Select>
                                </FormControl>
                            </Grid>

                            <Grid item xs={3}>
                                <TextField
                                    fullWidth
                                    label={i18n("chaincode_path")}
                                    variant="outlined"
                                    id="path"
                                    name="path"
                                    defaultValue={(this.state.ccInstallOption && this.state.ccInstallOption.path) || ""}
                                    InputProps={{
                                        classes: { input: classes.formField }
                                    }}
                                />
                            </Grid>
                            <Grid item xs={9}>
                                <Typography display="block" color="textSecondary" className={classes.formField}>
                                    {i18n("chaincode_path_remark")}
                                </Typography>
                            </Grid>

                            <Grid item xs={3}>
                                <input
                                    required
                                    accept="*"
                                    id="package"
                                    name="package"
                                    onChange={this.onSelectFile}
                                    type="file"
                                    style={{ display: "none" }}
                                />
                                <label htmlFor="package" display="inline">
                                    <Button fullWidth variant="outlined" display="inline" component="span" className={classes.button}>
                                        {i18n("chaincode_upload")}
                                    </Button>
                                    <Typography color="textSecondary" className={classes.formField}>{this.state.packageFile}</Typography>
                                </label>
                            </Grid>
                            <Grid item xs={9}>
                                <Typography display="inline" color="textSecondary" className={classes.formField}>{i18n("chaincode_package_remark")}</Typography>
                            </Grid>


                        </Grid>
                    </DialogContent>
                    <DialogActions>
                        <Grid container spacing={2}>
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
                                <Typography color="error" style={{ marginRight: "auto", fontSize: 11 }}>{this.state.installError}</Typography>
                            </Grid>
                            <Grid item xs={4}>
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
                                    autoFocus
                                    variant="contained"
                                    color="primary"
                                    style={{ marginLeft: "auto" }}
                                    disabled={this.state.loading}>
                                    {i18n("chaincode_install")}
                                </Button>
                            </Grid>
                        </Grid>
                    </DialogActions>
                </form>
            </Dialog>
        );
    }
}

export default withStyles(styles)(ChaincodeInstall);