import { gql } from "@apollo/client";
import { Card, CardContent, Stack, Typography } from "@mui/material";
import {
  ResourceConfiguration,
  useSourceTypeQuery,
} from "../../../graphql/generated";

import styles from "./configuration-page.module.scss";

gql`
  query SourceType($name: String!) {
    sourceType(name: $name) {
      metadata {
        name
        icon
        displayName
      }
    }
  }
`;

export const SourceCard: React.FC<{
  source: ResourceConfiguration;
  onClick?: () => void;
}> = ({ source, onClick }) => {
  const name = source.type ?? "";
  const { data } = useSourceTypeQuery({
    variables: { name },
  });
  const icon = data?.sourceType?.metadata.icon;
  const displayName = data?.sourceType?.metadata.displayName ?? "";
  const fontSize = displayName.length > 16 ? 14 : undefined;

  return (
    <Card className={styles["resource-card"]} onClick={onClick}>
      <CardContent>
        <Stack alignItems="center" textAlign={"center"}>
          <span
            className={styles.icon}
            style={{ backgroundImage: `url(${icon})` }}
          />
          <Typography component="div" fontWeight={600} fontSize={fontSize}>
            {displayName}
          </Typography>
        </Stack>
      </CardContent>
    </Card>
  );
};
