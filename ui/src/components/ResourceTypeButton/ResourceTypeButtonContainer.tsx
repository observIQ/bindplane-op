import React from "react";
import { InputAdornment, Stack, TextField } from "@mui/material";
import { SearchIcon } from "../Icons";

import styles from "./resource-button.module.scss";

interface Props {
  onSearchChange: (v: string) => void;
  placeholder?: string;
}

export const ResourceTypeButtonContainer: React.FC<Props> = ({
  children,
  onSearchChange,
  placeholder,
}) => {
  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    onSearchChange(e.target.value);
  }

  return (
    <>
      <TextField
        placeholder={placeholder ?? "Search for a technology..."}
        size="small"
        onChange={handleChange}
        type="search"
        fullWidth
        InputProps={{
          startAdornment: (
            <>
              <InputAdornment position="start">
                <SearchIcon />
              </InputAdornment>
            </>
          ),
        }}
      />
      <div className={styles.box}>
        <Stack spacing={1}>{children}</Stack>
      </div>
    </>
  );
};
