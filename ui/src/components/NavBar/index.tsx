import { AppBar, IconButton, Menu, MenuItem, Toolbar } from "@mui/material";
import React, { SyntheticEvent, useEffect, useState } from "react";
import { Link, NavLink, useNavigate } from "react-router-dom";
import {
  EmailIcon,
  GridIcon,
  HelpCircleIcon,
  LogoutIcon,
  SettingsIcon,
  SlackIcon,
  SlidersIcon,
  SquareIcon,
} from "../Icons";
import { BindPlaneOPLogo } from "../Logos";
import { classes } from "../../utils/styles";

import styles from "./nav-bar.module.scss";

export const NavBar: React.FC = () => {
  const [settingsAnchorEl, setAnchorEl] = useState<Element | null>(null);
  const settingsOpen = Boolean(settingsAnchorEl);

  const navigate = useNavigate();

  // make navigate available to the global window
  // to let us use it outside of components.
  useEffect(() => {
    window.navigate = navigate;
  }, [navigate]);

  const handleSettingsClick = (event: SyntheticEvent) => {
    setAnchorEl(event.currentTarget);
  };

  const handleSettingsClose = () => {
    setAnchorEl(null);
  };

  async function handleLogout() {
    await fetch("/logout", {
      method: "PUT",
    });

    localStorage.removeItem("user");
    navigate("/login");
  }

  return (
    <>
      <AppBar position="static" classes={{ root: styles["app-bar-root"] }}>
        <Toolbar classes={{ root: styles.toolbar }}>
          <Link to="/">
            <BindPlaneOPLogo
              className={styles.logo}
              aria-label="bindplane-logo"
            />
          </Link>

          <div className={styles["main-nav"]}>
            <NavLink
              className={({ isActive }) =>
                isActive
                  ? classes([styles["nav-link"], styles["active"]])
                  : styles["nav-link"]
              }
              to="/agents"
            >
              <GridIcon className={styles.icon} />
              Agents
            </NavLink>

            <NavLink
              className={({ isActive }) =>
                isActive
                  ? classes([styles["nav-link"], styles["active"]])
                  : styles["nav-link"]
              }
              to="/configurations"
            >
              <SlidersIcon className={styles.icon} />
              Configs
            </NavLink>

            <NavLink
              className={({ isActive }) =>
                isActive
                  ? classes([styles["nav-link"], styles["active"]])
                  : styles["nav-link"]
              }
              to="/components"
            >
              <SquareIcon className={styles.icon} />
              Components
            </NavLink>
          </div>

          <div className={styles["sub-nav"]}>
            <IconButton
              className={styles.button}
              target="_blank"
              color="inherit"
              data-testid="doc-link"
              href="https://docs.bindplane.observiq.com/docs"
            >
              <HelpCircleIcon className={styles.icon} />
            </IconButton>
            <IconButton
              className={styles.button}
              target="_blank"
              color="inherit"
              data-testid="support-link"
              href="mailto:support@observiq.com"
            >
              <EmailIcon className={styles.icon} />
            </IconButton>
            <IconButton
              className={styles.button}
              target="_blank"
              color="inherit"
              data-testid="slack-link"
              href="https://observiq.com/support-bindplaneop/"
            >
              <SlackIcon className={styles.icon} />
            </IconButton>
            <IconButton
              className={styles.button}
              aria-controls={settingsOpen ? "settings-menu" : undefined}
              aria-haspopup="true"
              aria-expanded={settingsOpen ? "true" : undefined}
              color="inherit"
              data-testid="settings-button"
              onClick={handleSettingsClick}
            >
              <SettingsIcon className={styles.icon} />
            </IconButton>
            <Menu
              anchorEl={settingsAnchorEl}
              open={settingsOpen}
              onClose={handleSettingsClose}
              anchorOrigin={{
                vertical: "bottom",
                horizontal: "center",
              }}
              transformOrigin={{
                vertical: "top",
                horizontal: "right",
              }}
              MenuListProps={{
                "aria-labelledby": "settings-button   ",
              }}
            >
              <MenuItem onClick={handleLogout}>
                <LogoutIcon className={styles["settings-icon"]} />
                Logout
              </MenuItem>
            </Menu>
          </div>
        </Toolbar>
      </AppBar>
    </>
  );
};

export function withNavBar(FC: React.FC): React.FC {
  return () => (
    <>
      <NavBar />
      <div className="content">
        <FC />
      </div>
    </>
  );
}
