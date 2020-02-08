import React from 'react';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import DashboardIcon from '@material-ui/icons/Dashboard';
import { Link } from "react-router-dom";
import i18n from "../../i18n";
import "../../css/common.css";
import { Typography } from '@material-ui/core';

export const mainListItems = (
    <div>
        <Link to="/">
            <ListItem button>
                <ListItemIcon>
                    <DashboardIcon />
                </ListItemIcon>
                <ListItemText >
                    <Typography style={{ fontSize: 13 }}>{i18n("network")}</Typography>
                </ListItemText>
            </ListItem>
        </Link>

        {/* <Link to="/orders">
            <ListItem button>
                <ListItemIcon>
                    <DashboardIcon />
                </ListItemIcon>
                <ListItemText >
                    <Typography style={{ fontSize: 13 }}>{i18n("channel")}</Typography>
                </ListItemText>
            </ListItem>
        </Link>

        <Link to="/chaincode">
            <ListItem button>
                <ListItemIcon>
                    <LayersIcon />
                </ListItemIcon>
                <ListItemText >
                    <Typography style={{ fontSize: 13 }}>{i18n("chaincode")}</Typography>
                </ListItemText>
            </ListItem>
        </Link> */}
    </div>
);
