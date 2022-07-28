import { gql } from "@apollo/client";
import {
  Box,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  SelectChangeEvent,
  Typography,
} from "@mui/material";
import React, { useEffect, useState } from "react";
import { CardContainer } from "../../components/CardContainer";
import { CodeBlock } from "../../components/CodeBlock";
import {
  GetConfigurationNamesQuery,
  useGetConfigurationNamesQuery,
} from "../../graphql/generated";
import { InstallCommandResponse } from "../../types/rest";
import { withRequireLogin } from "../../contexts/RequireLogin";
import { withNavBar } from "../../components/NavBar";
import { PlatformSelect } from "../../components/PlatformSelect";

import mixins from "../../styles/mixins.module.scss";

gql`
  query GetConfigurationNames {
    configurations {
      configurations {
        metadata {
          name
          labels
        }
      }
    }
  }
`;

enum Platform {
  Linux = "linux",
  macOS = "macos",
  Windows = "windows",
}

const InstallPageContent: React.FC = () => {
  const [platform, setPlatform] = useState<string>("linux");
  const [installCommand, setCommand] = useState("");
  const [configs, setConfigs] = useState<string[]>([]);
  const [selectedConfig, setSelectedConfig] = useState<string>("");
  const { data } = useGetConfigurationNamesQuery();

  useEffect(() => {
    if (data) {
      // First filter the configs to match the platform
      const filtered = filterConfigurationsByPlatform(
        data.configurations.configurations,
        platform
      );

      const configNames = filtered.map((c) => c.metadata.name);

      setConfigs(configNames);
    }
  }, [data, platform, setConfigs]);

  useEffect(() => {
    async function fetchInstallText() {
      const url = installCommandUrl({
        platform,
        configuration: selectedConfig,
      });
      const resp = await fetch(url);
      const { command } = (await resp.json()) as InstallCommandResponse;
      if (resp.status === 200) {
        setCommand(command);
      }
    }

    fetchInstallText();
  }, [platform, selectedConfig]);

  return (
    <CardContainer>
      <Typography variant="h5" classes={{ root: mixins["mb-5"] }}>
        Agent Installation
      </Typography>

      <Box
        component="form"
        className={`${mixins["form-width"]} ${mixins["mb-3"]}`}
      >
        <PlatformSelect
          value={platform}
          onPlatformSelected={(v) => setPlatform(v)}
        />

        {configs.length > 0 && (
          <>
            <FormControl fullWidth margin="normal">
              <InputLabel id="config-label">
                Select Configuration (optional)
              </InputLabel>

              <Select
                labelId="config-label"
                id="configuration"
                label="Select Configuration (optional)"
                onChange={(e: SelectChangeEvent<string>) => {
                  setSelectedConfig(e.target.value);
                }}
                value={selectedConfig}
              >
                <MenuItem value="">
                  <em>None</em>
                </MenuItem>
                {configs.map((p) => (
                  <MenuItem key={p} value={p}>
                    {p}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </>
        )}
      </Box>

      <CodeBlock value={installCommand} />
    </CardContainer>
  );
};

function installCommandUrl(params: {
  platform: string;
  configuration?: string;
}): string {
  const url = new URL(window.location.href);
  url.pathname = "/v1/agent-versions/latest/install-command";

  const searchParams: { platform: string; labels?: string } = {
    platform: params.platform,
  };

  if (params.configuration) {
    searchParams.labels = encodeURI(`configuration=${params.configuration}`);
  }

  url.search = new URLSearchParams(searchParams).toString();
  return url.href;
}

function filterConfigurationsByPlatform(
  configs: GetConfigurationNamesQuery["configurations"]["configurations"],
  platform: string
): GetConfigurationNamesQuery["configurations"]["configurations"] {
  switch (platform) {
    case Platform.Linux:
      return configs.filter((c) => c.metadata.labels.platform === "linux");
    case Platform.macOS:
      return configs.filter((c) => c.metadata.labels.platform === "macos");
    case Platform.Windows:
      return configs.filter((c) => c.metadata.labels.platform === "windows");
    default:
      return configs;
  }
}

export const InstallPage = withRequireLogin(withNavBar(InstallPageContent));
