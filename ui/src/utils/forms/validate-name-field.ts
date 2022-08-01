const ALPHA_NUMERIC_START_END_REGEX = /(([A-Za-z0-9].*)?[A-Za-z0-9])?/;
const LABEL_REGEX = /(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?/;

/** validateNameField makes sure that the BindPlane Resource name is a valid selector.  ie.
 * 1. Matches label regex
 * 2. Contains no spaces
 * 3. Must be 63 characters or less
 * 4. If existingResources is passed, it validates that the name is not taken.
 */
export function validateNameField(
  name?: string | null,
  kind?: "configuration" | "source" | "destination" | "processor",
  existingNames?: string[]
): string | null {
  if (name == null || name === "") {
    return "Required.";
  } else {
    // name must be valid label

    // Label Regex
    const match0 = ALPHA_NUMERIC_START_END_REGEX.exec(name);
    if (match0 != null && match0[0] !== name) {
      return "Must begin and end with an alphanumeric character.";
    }

    const match = LABEL_REGEX.exec(name);
    if (match != null && match[0] !== name) {
      return "Invalid character. Can contain alphanumeric characters, dashes ( - ), underscores ( _ ), and dots ( . ).";
    }

    // No Spaces
    if (name.includes(" ")) {
      return "Must not contain spaces.";
    }

    // Length
    if (name.length > 63) {
      return "Must be 63 characters or less.";
    }

    // Verify name does not exist already.
    if (existingNames != null) {
      const r = existingNames.find((existingName) => existingName === name);
      if (r != null) {
        return `A ${kind} named ${name} already exists.`;
      }
    }
  }
  return null;
}
