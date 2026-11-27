import { render, screen } from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router";
import { expect, it } from "vitest";
import { routes } from "./routes";

it("renders the Dashboard page inside the shell with section links", async () => {
  render(<RouterProvider router={createMemoryRouter(routes)} />);

  expect(
    await screen.findByRole("heading", { name: "Dashboard" }),
  ).toBeInTheDocument();
  expect(screen.getByRole("link", { name: "Payment methods" })).toHaveAttribute(
    "href",
    "/payment-methods",
  );
});
