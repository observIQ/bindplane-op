import { gql } from "@apollo/client";
import { Card, IconButton, Stack, Typography } from "@mui/material";
import { useSnackbar } from "notistack";
import { useEffect, useRef } from "react";
import { useDrag, useDrop } from "react-dnd";
import {
  ResourceConfiguration,
  useGetProcessorTypeQuery,
} from "../../graphql/generated";
import { EditIcon, MenuIcon } from "../Icons";

import styles from "./inline-processor-label.module.scss";

interface Props {
  index: number;
  processor: ResourceConfiguration;
  onEdit: () => void;
  // Move processor should change the components order state
  moveProcessor: (dragIndex: number, dropIndex: number) => void;

  // onDrop should "save" the current order to the formValues and apply
  // the changed configuration.
  onDrop: () => void;
}

gql`
  query getProcessorType($type: String!) {
    processorType(name: $type) {
      metadata {
        displayName
        name
        description
      }
      spec {
        parameters {
          label
          name
          description
          required
          type
          default
          relevantIf {
            name
            operator
            value
          }
          validValues
        }
      }
    }
  }
`;

type Item = {
  index: number;
};

export const InlineProcessorLabel: React.FC<Props> = ({
  index,
  processor,
  onEdit,
  onDrop,
  moveProcessor,
}) => {
  // TODO (dsvanlani) handle loading and error
  const { data, error } = useGetProcessorTypeQuery({
    variables: { type: processor.type! },
  });

  const { enqueueSnackbar } = useSnackbar();

  useEffect(() => {
    if (error != null) {
      console.error(error);
      enqueueSnackbar("Error retrieving Processor Type", {
        variant: "error",
        key: "Error retrieving Processor Type",
      });
    }
  }, [enqueueSnackbar, error]);

  const [, dragRef] = useDrag({
    type: "inline-processor",
    item: { index },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
  });

  const [{ isHovered }, dropRef] = useDrop<
    Item,
    unknown,
    { isHovered: boolean }
  >({
    accept: "inline-processor",
    collect: (monitor) => ({
      isHovered: monitor.isOver(),
    }),
    hover: (item, monitor) => {
      if (ref.current == null) {
        return;
      }

      if (monitor == null) {
        return;
      }

      const dragIndex = item.index;
      const hoverIndex = index;

      moveProcessor(dragIndex, hoverIndex);
      item.index = hoverIndex;
    },

    // Save the order on drop
    drop: onDrop,
  });

  // Join the 2 refs together into one (both draggable and can be dropped on)
  const ref = useRef<HTMLDivElement>(null);

  const dragDropRef = dragRef(dropRef(ref)) as any;

  return (
    <Card
      variant="outlined"
      ref={dragDropRef}
      style={{
        border: isHovered ? "1px solid #4abaeb" : undefined,
      }}
      classes={{ root: styles.card }}
    >
      <Stack
        direction="row"
        alignItems={"center"}
        spacing={1}
        justifyContent={"space-between"}
      >
        <Stack direction={"row"} spacing={1}>
          <MenuIcon className={styles["hover-icon"]} />
          <Typography>{data?.processorType?.metadata.displayName}</Typography>
        </Stack>

        <IconButton onClick={onEdit}>
          <EditIcon width={15} height={15} style={{ float: "right" }} />
        </IconButton>
      </Stack>
    </Card>
  );
};
