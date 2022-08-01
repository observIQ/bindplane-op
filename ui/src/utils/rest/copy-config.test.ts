import nock from "nock";
import { copyConfig } from "./copy-config";

describe("duplicateConfig", () => {
  const existingName = "name";
  const newName = "new-name";
  const endpoint = "/v1/configurations/name/copy";

  it("correct endpoint and payload", async () => {
    let endpointCalled = false;
    let payload: any;
    nock("http://localhost:80")
      .post(endpoint, (body) => {
        endpointCalled = true;
        payload = body;
        return true;
      })
      .reply(201, {
        name: newName,
      });

    await copyConfig({ existingName, newName });

    expect(endpointCalled).toEqual(true);
    expect(payload).toEqual({ name: newName });
  });

  it("created", async () => {
    nock("http://localhost:80")
      .post("/v1/configurations/name/copy", (body) => {
        return true;
      })
      .reply(201, {
        name: "new-name",
      });

    const status = await copyConfig({ existingName, newName });
    expect(status).toEqual("created");
  });
  it("conflict", async () => {
    nock("http://localhost:80")
      .post("/v1/configurations/name/copy", (body) => {
        return true;
      })
      .reply(409, {});

    const status = await copyConfig({ existingName, newName });
    expect(status).toEqual("conflict");
  });
  it("error", async () => {
    nock("http://localhost:80")
      .post("/v1/configurations/name/copy", (body) => {
        return true;
      })
      .reply(500, {});

    const status = await copyConfig({ existingName, newName });
    expect(status).toEqual("error");
  });
});
