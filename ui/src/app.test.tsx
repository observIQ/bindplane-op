import { render, screen } from "@testing-library/react";
import { App } from "./app";

describe("App", () => {
  it("unauthenticated will render login page", () => {
    render(<App />);
    screen.getByTestId("login-page");
  });
});
