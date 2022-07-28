import { isEmpty } from "lodash";

export function validateStringsField(
  value: string[],
  required?: boolean
): string | null {
  if (required && isEmpty(value)) {
    return "Required.";
  }

  return null;
}
