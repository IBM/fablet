import React from 'react';
import clsx from 'clsx';
import { withStyles } from '@material-ui/core/styles';
import CssBaseline from '@material-ui/core/CssBaseline';
import Drawer from '@material-ui/core/Drawer';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import List from '@material-ui/core/List';
import Typography from '@material-ui/core/Typography';
import Divider from '@material-ui/core/Divider';
import IconButton from '@material-ui/core/IconButton';
import Container from '@material-ui/core/Container';
import MenuIcon from '@material-ui/icons/Menu';
import ChevronLeftIcon from '@material-ui/icons/ChevronLeft';
import { mainListItems } from './menu/list_items';
import ConnectionMenu from "./menu/connection";
import { BrowserRouter as Router, Switch, Route } from "react-router-dom";

import { clientStorage } from "../data/localstore";
import Profile from './network/profile';
import Discover from "./network/discover";
import { log } from "../common/log";

function Copyright() {
    return (
        <Typography variant="body2" color="textSecondary" align="center">
            {'Copyright © '}
            Fablet
            {' '}
            {new Date().getFullYear()}
            {'.'}
        </Typography>
    );
}

const drawerWidth = 240;

const styles = (theme) => ({

    root: {
        display: 'flex'
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
    appBarShift: {
        marginLeft: drawerWidth,
        width: `calc(100% - ${drawerWidth}px)`,
        transition: theme.transitions.create(['width', 'margin'], {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.enteringScreen,
        }),
    },
    menuButton: {
        marginRight: 36,
    },
    menuButtonHidden: {
        display: 'none',
    },
    title: {
        flexGrow: 1,
    },
    drawerPaper: {
        position: 'relative',
        whiteSpace: 'nowrap',
        width: drawerWidth,
        transition: theme.transitions.create('width', {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.enteringScreen,
        }),
    },
    drawerPaperClose: {
        overflowX: 'hidden',
        transition: theme.transitions.create('width', {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen,
        }),
        width: theme.spacing(7),
        [theme.breakpoints.up('sm')]: {
            width: theme.spacing(9),
        },
    },
    appBarSpacer: theme.mixins.toolbar,
    content: {
        flexGrow: 1,
        height: '100vh',
        overflow: 'auto'
    },
    container: {
        paddingTop: theme.spacing(4),
        paddingBottom: theme.spacing(4),
    },
    paper: {
        padding: theme.spacing(2),
        display: 'flex',
        overflow: 'auto',
        flexDirection: 'column',
    },
    fixedHeight: {
        height: 240,
    },
});


class Main extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            ...props,
            leftMenuOpen: true,
            profileOpen: false,
        };

        this.handleLeftMenuOpen = this.handleLeftMenuOpen.bind(this);
        this.handleLeftMenuClose = this.handleLeftMenuClose.bind(this);
        this.handleProfileOpen = this.handleProfileOpen.bind(this);
        this.handleProfileClose = this.handleProfileClose.bind(this);

        log.debug("Main: constructor");
    }

    componentDidMount() {
        const needConnProf = clientStorage.loadConnectionProfiles().length <= 0;
        const needIdentity = clientStorage.loadIdentities().length <= 0;

        this.setState({ 
            profileOpen: needConnProf || needIdentity,
            needConnProf: needConnProf,
            needIdentity: needIdentity
         });
    }

    componentDidUpdate() {
    }

    componentWillUnmount() {
    }

    handleLeftMenuOpen() {
        this.setState({ leftMenuOpen: true });
    }

    handleLeftMenuClose() {
        this.setState({ leftMenuOpen: false });
    }

    handleProfileOpen(needConnProf, needIdentity) {
        const main = this;
        // TODO UI refresh performance issues.
        // log.info(">>>>>>>>>>>>>>>>>>>>>> return the function self with ", needConnProf, needIdentity);
        return function () {
            // log.info("####################### execute the function self with ", needConnProf, needIdentity);
            main.setState({
                profileOpen: true,
                needConnProf: needConnProf,
                needIdentity: needIdentity
            });
        }
    }

    handleProfileClose() {
        this.setState({ profileOpen: false });
    }

    render() {
        const classes = this.props.classes;

        const connProfs = clientStorage.loadConnectionProfiles();
        const identities = clientStorage.loadIdentities();

        return (
            <div className={classes.root}>
                <CssBaseline />
                <AppBar position="absolute" className={clsx(classes.appBar, this.state.leftMenuOpen && classes.appBarShift)}>
                    <Toolbar className={classes.toolbar}>
                        <IconButton
                            edge="start"
                            color="inherit"
                            aria-label="open drawer"
                            onClick={this.handleLeftMenuOpen}
                            className={clsx(classes.menuButton, this.state.leftMenuOpen && classes.menuButtonHidden)}
                        >
                            <MenuIcon />
                        </IconButton>
                        <Typography component="h1" variant="h4" color="inherit" noWrap className={classes.title}>
                            Fablet
                        </Typography>
                        {/* <IconButton color="inherit">
                            <Badge badgeContent={4} color="secondary">
                                <NotificationsIcon />
                            </Badge>
                        </IconButton> */}
                    </Toolbar>
                </AppBar>

                <Router>

                    <Drawer
                        variant="permanent"
                        classes={{
                            paper: clsx(classes.drawerPaper, !this.state.leftMenuOpen && classes.drawerPaperClose),
                        }}
                        open={this.state.leftMenuOpen}
                    >
                        <div className={classes.toolbarIcon}>
                            <IconButton onClick={this.handleLeftMenuClose}>
                                <ChevronLeftIcon />
                            </IconButton>
                        </div>
                        <Divider />
                        <List>
                            {mainListItems}
                        </List>
                        <Divider />
                        <List>
                            <ConnectionMenu
                                onOpenProfile={this.handleProfileOpen}
                            />
                        </List>
                    </Drawer>

                    <main className={classes.content}>
                        <div className={classes.appBarSpacer} />
                        <Container className={classes.container} maxWidth={false}>
                            {
                                connProfs.length > 0 && identities.length > 0 ? (
                                    <Switch>
                                        <Route key={1} path={"/"} exact={true} children={<Discover/>} />
                                    </Switch>)
                                    :
                                    <div></div>
                            }
                        </Container>

                        <Profile key={"profile_" + this.state.profileOpen}
                            open={this.state.profileOpen}
                            needConnProf={this.state.needConnProf}
                            needIdentity={this.state.needIdentity}
                            callBack={this.handleProfileClose}
                            onClose={this.handleProfileClose}
                        />

                        <Copyright />
                    </main>

                </Router>


            </div>
        );
    }

}

export default withStyles(styles)(Main);