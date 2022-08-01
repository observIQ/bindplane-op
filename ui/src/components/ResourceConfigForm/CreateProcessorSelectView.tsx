import { gql } from "@apollo/client";
import { Box, Button, CircularProgress } from "@mui/material";
import { useSnackbar } from "notistack";
import { useEffect, useState } from "react";
import { ButtonFooter, FormTitle, ProcessorType } from ".";
import { useGetProcessorTypesQuery } from "../../graphql/generated";
import { metadataSatisfiesSubstring } from "../../utils/metadata-satisfies-substring";
import {
  ResourceTypeButton,
  ResourceTypeButtonContainer,
} from "../ResourceTypeButton";

gql`
  query getProcessorTypes {
    processorTypes {
      metadata {
        displayName
        description
        name
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

interface CreateProcessorSelectViewProps {
  title: string;
  onBack: () => void;
  onSelect: (pt: ProcessorType) => void;
}

export const CreateProcessorSelectView: React.FC<CreateProcessorSelectViewProps> =
  ({ title, onBack, onSelect }) => {
    const { data, loading, error } = useGetProcessorTypesQuery();
    const [search, setSearch] = useState("");
    const { enqueueSnackbar } = useSnackbar();

    useEffect(() => {
      if (error != null) {
        enqueueSnackbar("Error retrieving data for Processor Type.", {
          variant: "error",
          key: "Error retrieving data for Processor Type.",
        });
      }
    }, [enqueueSnackbar, error]);

    const backButton: JSX.Element = (
      <Button variant="contained" color="secondary" onClick={onBack}>
        Back
      </Button>
    );

    return (
      <>
        <FormTitle
          title={title}
          crumbs={["Add a processor"]}
          description={"Select a processor type to configure."}
        />

        <ResourceTypeButtonContainer
          onSearchChange={(v: string) => setSearch(v)}
          placeholder={"Search for a processor..."}
        >
          {loading && (
            <Box display="flex" justifyContent={"center"} marginTop={2}>
              <CircularProgress />
            </Box>
          )}
          {data?.processorTypes
            .filter((pt) => metadataSatisfiesSubstring(pt, search))
            .map((p) => (
              <ResourceTypeButton
                hideIcon
                key={`${p.metadata.name}`}
                displayName={p.metadata.displayName!}
                onSelect={() => {
                  onSelect(p);
                }}
              />
            ))}
        </ResourceTypeButtonContainer>
        <ButtonFooter
          primaryButton={<></>}
          secondaryButton={<></>}
          backButton={backButton}
        />
      </>
    );
  };
