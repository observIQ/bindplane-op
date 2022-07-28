import React, { ChangeEvent, useState } from "react";
import "@uiw/react-textarea-code-editor/dist.css";
import CodeEditor from "@uiw/react-textarea-code-editor";
import { ChevronDown, ChevronUp } from "../Icons";
import { classes } from "../../utils/styles";
import { Button } from "@mui/material";
import { useRef } from "react";
import { useEffect } from "react";

import styles from "./yaml-editor.module.scss";

interface YamlEditorProps
  extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  readOnly?: boolean;
  limitHeight?: boolean;
  minHeight?: number;
  inputRef?: React.RefObject<HTMLTextAreaElement>;
  onValueChange?: (e: ChangeEvent<HTMLTextAreaElement>) => void;
  value: string;
}

export const YamlEditor: React.FC<YamlEditorProps> = ({
  inputRef,
  limitHeight = false,
  readOnly,
  value,
  onValueChange,
  ...rest
}) => {
  // We are only using light theme right now.  This overrides the styling if a user
  // has dark mode as a browser preference.
  document.documentElement.setAttribute("data-color-mode", "light");
  const textRef = useRef<HTMLTextAreaElement | null>(null);
  const ref = inputRef ?? textRef;

  const [expanded, setExpanded] = useState(false);
  const [expandable, setExpandable] = useState(false);

  useEffect(() => {
    if (ref.current && ref.current.scrollHeight > 300) {
      setExpandable(true);
      return;
    }
    setExpandable(false);
  }, [ref]);

  const classNames = [styles["code-container"]];
  if (expanded || !limitHeight || !readOnly) {
    classNames.push(styles.expanded);
  }

  return (
    <div>
      <CodeEditor
        {...rest}
        data-testid="yaml-editor"
        className={classes(classNames)}
        readOnly={readOnly}
        value={value}
        ref={ref}
        language="yaml"
        onChange={onValueChange}
        padding={15}
        style={{
          backgroundColor: readOnly ? "#f5f5f5" : "#fff",
          fontFamily:
            "ui-monospace,SFMono-Regular,SF Mono,Consolas,Liberation Mono,Menlo,monospace",
          fontSize: 12,
          border: readOnly ? "1px solid #f5f5f5 " : "1px solid #aaa",
        }}
      />

      {/* Allow expand and collapse in readOnly mode, if height is over 300 and the limitHeight prop was passed */}
      {expandable && readOnly && limitHeight && (
        <span className={styles["button-row"]}>
          {!expanded && (
            <Button
              variant="contained"
              color="secondary"
              size="small"
              onClick={() => setExpanded(true)}
              endIcon={<ChevronDown />}
            >
              Show more
            </Button>
          )}
          {expanded && (
            <Button
              variant="contained"
              color="secondary"
              size="small"
              onClick={() => setExpanded(false)}
              endIcon={<ChevronUp />}
            >
              Show less
            </Button>
          )}
        </span>
      )}
    </div>
  );
};
