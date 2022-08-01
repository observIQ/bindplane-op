import { isEmpty } from "lodash";

type ResourceWithMetadata = {
  metadata: {
    name: string;
    displayName?: string | null;
  };
};

export function metadataSatisfiesSubstring(
  rt: ResourceWithMetadata,
  substring: string
) {
  return isEmpty(substring)
    ? true
    : rt.metadata.name.includes(substring) ||
        rt.metadata.displayName?.includes(substring) ||
        rt.metadata.displayName?.toLowerCase().includes(substring);
}
