import { DuplicateConfigPayload } from "../../types/rest";

export async function duplicateConfig({
  existingName,
  newName,
}: {
  existingName: string;
  newName: string;
}): Promise<"created" | "conflict" | "error"> {
  const payload: DuplicateConfigPayload = {
    name: newName,
  };
  try {
    const resp = await fetch(`/v1/configurations/${existingName}/duplicate`, {
      method: "PUT",
      body: JSON.stringify(payload),
    });

    switch (resp.status) {
      case 201:
        return "created";
      case 409:
        return "conflict";
      default:
        return "error";
    }
  } catch (err) {
    return "error";
  }
}
