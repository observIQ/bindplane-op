import { Button, Divider, Typography } from "@mui/material";
import { ResourceConfiguration } from "../../graphql/generated";
import { PlusCircleIcon } from "../Icons";
import { InlineProcessorLabel } from "./InlineProcessorLabel";
import { DndProvider } from "react-dnd";
import { HTML5Backend } from "react-dnd-html5-backend";
import { useCallback, useState } from "react";
import { FormValues } from ".";

import mixins from "../../styles/mixins.module.scss";

interface Props {
  processors: ResourceConfiguration[];
  onAddProcessor: () => void;
  onEditProcessor: (index: number) => void;
  setFormValues: (value: React.SetStateAction<FormValues>) => void;
}

export const InlineProcessorContainer: React.FC<Props> = ({
  setFormValues,
  processors: processorsProp,
  onAddProcessor,
  onEditProcessor,
}) => {
  // Manage the processor order state internally in this component
  const [processors, setProcessors] = useState(processorsProp);

  function handleDrop() {
    setFormValues((prev: FormValues) => ({
      ...prev,
      processors,
    }));
  }

  const moveProcessor = useCallback(
    (dragIndex: number, hoverIndex: number) => {
      if (dragIndex === hoverIndex) {
        return;
      }

      const newProcessors = [...processors];

      const dragItem = newProcessors[dragIndex];
      const hoverItem = newProcessors[hoverIndex];

      // Swap places of dragItem and hoverItem in the array
      newProcessors[dragIndex] = hoverItem;
      newProcessors[hoverIndex] = dragItem;

      setProcessors(newProcessors);
    },
    [processors]
  );

  return (
    <>
      <DndProvider backend={HTML5Backend}>
        <Divider />
        <Typography fontWeight={600} marginTop={2}>
          Processors
        </Typography>

        <Typography variant="body2" marginBottom={2}>
          Processors are run on data after it&apos;s received and prior to being
          sent to a destination. They will be executed in the order they appear
          below.
        </Typography>
        {processors.map((p, ix) => {
          return (
            <InlineProcessorLabel
              moveProcessor={moveProcessor}
              key={`${p.name}-${ix}`}
              processor={p}
              onEdit={() => onEditProcessor(ix)}
              onDrop={handleDrop}
              index={ix}
            />
          );
        })}

        <Button
          variant="text"
          startIcon={<PlusCircleIcon />}
          classes={{ root: mixins["mb-2"] }}
          onClick={onAddProcessor}
        >
          Add processor
        </Button>
        <Divider />
      </DndProvider>
    </>
  );
};
