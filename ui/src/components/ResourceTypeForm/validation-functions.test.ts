import { validateStringsField } from "./validation-functions";

describe("validateStringsField", () => {
  it("[], required", () => {
    const error = validateStringsField([], true);
    expect(error).not.toBeNull();
  });

  it("[], not required", () => {
    const error = validateStringsField([], false);
    expect(error).toBeNull();
  });
});
