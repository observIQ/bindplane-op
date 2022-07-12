import {
  Button,
  Card,
  CardContent,
  FormControl,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import React, { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { BindPlaneOPLogo } from "../../components/Logos";

import styles from "./login.module.scss";

export const LoginPage: React.FC = () => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [invalidCreds, setInvalidCreds] = useState(false);
  const formRef = useRef<HTMLFormElement | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    if (localStorage.getItem("user") != null) {
      navigate("/agents");
    }
  }, [navigate]);

  async function handleLogin(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (formRef.current == null) {
      return;
    }

    const data = new FormData();
    data.append("username", username);
    data.append("password", password);
    const resp = await fetch("/login", {
      method: "POST",
      body: data,
    });

    if (resp.status === 401) {
      setInvalidCreds(true);
      return;
    }

    // TODO (auth) check for correct return status
    localStorage.setItem("user", username);

    navigate("/agents");
  }

  return (
    <div className={styles["login-page"]} data-testid="login-page">
      <Stack alignItems={"center"} justifyContent={"center"}>
        <BindPlaneOPLogo width={225} height={60} className={styles.logo} />
        <Card classes={{ root: styles.card }}>
          <CardContent>
            <Typography variant="h5" fontWeight={600}>
              Sign In
            </Typography>
            <form
              action="/login"
              method="POST"
              ref={formRef}
              onSubmit={(e) => {
                handleLogin(e);
              }}
            >
              <Stack>
                <TextField
                  type="text"
                  label="Username"
                  name="username"
                  value={username}
                  onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                    setUsername(event.target.value);
                  }}
                  size="small"
                  margin="normal"
                />
                <TextField
                  type="password"
                  label="Password"
                  name="password"
                  value={password}
                  onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                    setPassword(event.target.value);
                  }}
                  size="small"
                  margin="normal"
                />
              </Stack>

              <FormControl margin="normal" fullWidth>
                {invalidCreds && (
                  <Typography variant="body2" color="error">
                    Invalid username or password.
                  </Typography>
                )}
              </FormControl>

              <FormControl margin="normal" fullWidth>
                <Button
                  type="submit"
                  variant="contained"
                  disabled={username === "" || password === ""}
                >
                  Sign In
                </Button>
              </FormControl>
            </form>
          </CardContent>
        </Card>
      </Stack>
    </div>
  );
};
