import { validateMapField, validateStringsField } from "./validation-functions";

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

describe("validateMapField", () => {
  it("{}, required => error", () => {
    const error = validateMapField({}, true);
    expect(error).not.toBeNull();
  });

  it(`{"":""}, required => error`, () => {
    const error = validateMapField({ "": "" }, true);
    expect(error).not.toBeNull();
  });
});
