import { render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
import Home from "@/app/page";

test("home page renders without errors", () => {
  render(<Home />);
  expect(screen.getByText(/to get started/i)).toBeDefined();
});
