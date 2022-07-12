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

import mixins from "../../styles/mixins.module.scss";
import { withRequireLogin } from "../../contexts/RequireLogin";
import { withNavBar } from "../../components/NavBar";

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
  LINUX_AMD64 = "linux-amd64",
  DARWIN_AMD64 = "darwin-amd64",
  DARWIN_ARM64 = "darwin-arm64",
  WINDOWS_AMD64 = "windows-amd64",
}

const platforms: { name: string; value: Platform }[] = [
  {
    name: "Linux",
    value: Platform.LINUX_AMD64,
  },
  {
    name: "macOS (Intel)",
    value: Platform.DARWIN_AMD64,
  },
  {
    name: "macOS (Apple M1)",
    value: Platform.DARWIN_ARM64,
  },
  {
    name: "Windows",
    value: Platform.WINDOWS_AMD64,
  },
];

const InstallPageContent: React.FC = () => {
  const [platform, setPlatform] = useState<Platform>(Platform.LINUX_AMD64);
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
        <FormControl fullWidth margin="normal">
          <InputLabel id="platform-label">Choose OS</InputLabel>

          <Select
            labelId="platform-label"
            id="platform"
            label="Choose OS"
            onChange={(e: SelectChangeEvent<string>) => {
              setPlatform(e.target.value as Platform);
            }}
            value={platform}
          >
            {platforms.map((p) => (
              <MenuItem key={p.value} value={p.value}>
                {p.name}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

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
  platform: Platform | ""
): GetConfigurationNamesQuery["configurations"]["configurations"] {
  switch (platform) {
    case Platform.LINUX_AMD64:
      return configs.filter((c) => c.metadata.labels.platform === "linux");
    case Platform.DARWIN_AMD64:
    case Platform.DARWIN_ARM64:
      return configs.filter((c) => c.metadata.labels.platform === "macos");
    case Platform.WINDOWS_AMD64:
      return configs.filter((c) => c.metadata.labels.platform === "windows");
    default:
      return configs;
  }
}

export const InstallPage = withRequireLogin(withNavBar(InstallPageContent));
