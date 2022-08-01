import { Stack } from "@mui/material";

interface ButtonFooterProps {
  backButton: JSX.Element;
  primaryButton: JSX.Element;
  secondaryButton: JSX.Element;
}

export const ButtonFooter: React.FC<ButtonFooterProps> = ({
  backButton,
  primaryButton,
  secondaryButton,
}) => {
  return (
    <Stack direction={"row"} justifyContent="space-between" marginTop={2}>
      {backButton}

      <div>
        {secondaryButton && secondaryButton}

        {primaryButton}
      </div>
    </Stack>
  );
};
