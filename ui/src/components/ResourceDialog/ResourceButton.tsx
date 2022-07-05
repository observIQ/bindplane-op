import { Card, Stack, Typography } from "@mui/material";
import { useState } from "react";
import { PipelineType } from "../../graphql/generated";
import { TelemetryChip } from "../Chips";

import styles from "./resource-button.module.scss";

interface ResourceButtonProps {
  icon: string;
  displayName: string;
  onSelect: () => void;
  telemetryTypes?: PipelineType[];
}

// Resource button is used to display a ResourceType, Source, or Destination as
// selectable button in a stack.
export const ResourceButton: React.FC<ResourceButtonProps> = ({
  icon,
  displayName,
  onSelect,
  telemetryTypes,
}) => {
  const [isHovered, setHovered] = useState(false);

  return (
    <Card
      role="button"
      onClick={onSelect}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      className={isHovered ? styles.hovered : styles.card}
    >
      <Stack
        direction="row"
        minHeight={55}
        alignItems="center"
        className={styles.pointer}
        justifyContent="space-between"
      >
        <Stack
          direction="row"
          alignItems={"center"}
          spacing={2}
          padding={1}
          marginLeft={1}
        >
          <span
            className={styles.icon}
            style={{ backgroundImage: `url(${icon})` }}
          />
          <Typography fontWeight={600} color={isHovered ? "#fff" : "inherit"}>
            {displayName}
          </Typography>
        </Stack>

        {telemetryTypes && (
          <Stack spacing={0.5} marginRight={1} direction="row">
            {telemetryTypes.map((t) => (
              <TelemetryChip key={t} hovered={isHovered} telemetryType={t} />
            ))}
          </Stack>
        )}
      </Stack>
    </Card>
  );
};
