import { CopyConfigPayload } from "../../types/rest";

export async function copyConfig({
  existingName,
  newName,
}: {
  existingName: string;
  newName: string;
}): Promise<"created" | "conflict" | "error"> {
  const payload: CopyConfigPayload = {
    name: newName,
  };
  try {
    const resp = await fetch(`/v1/configurations/${existingName}/copy`, {
      method: "POST",
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
