import { test, expect } from "@playwright/test";

// Full happy path against a RUNNING Go backend (task docker-up) with a
// pre-verified account. Register alone cannot log in (email verification is
// required — see research R-2), so the authenticated leg uses credentials for
// an account that was verified out-of-band.
//
// Provide them via env before running `pnpm e2e`:
//   E2E_EMAIL=verified@example.com E2E_PASSWORD=secret123 pnpm e2e
const email = process.env.E2E_EMAIL;
const password = process.env.E2E_PASSWORD;

test.describe("auth + resource happy path", () => {
  test.skip(
    !email || !password,
    "Set E2E_EMAIL and E2E_PASSWORD (a pre-verified account) with the Go backend running.",
  );

  test("register → login → create resource → logout → route protection", async ({ page }) => {
    // Register a fresh account → lands on /login with the verify-your-email notice.
    const fresh = `e2e-${Date.now()}@example.com`;
    await page.goto("/register");
    await page.getByLabel("Email").fill(fresh);
    await page.getByLabel("Password").fill("secret123");
    await page.getByRole("button", { name: "Create account" }).click();

    await expect(page).toHaveURL(/\/login/);
    await expect(page.getByRole("status")).toContainText(/verify/i);

    // Log in with the pre-verified account → land in the protected app.
    await page.getByLabel("Email").fill(email!);
    await page.getByLabel("Password").fill(password!);
    await page.getByRole("button", { name: "Sign in" }).click();

    await expect(page).toHaveURL(/\/dashboard/);

    // Create a resource and see it in the list.
    await page.goto("/resources");
    await page.getByRole("button", { name: "New resource" }).click();

    const title = `E2E resource ${Date.now()}`;
    await page.getByLabel("Title").fill(title);
    await page.getByLabel("Description").fill("Created by the Playwright happy-path test.");
    await page.getByRole("button", { name: "Create resource" }).click();

    await expect(page.getByText(title)).toBeVisible();

    // Log out → session revoked, redirected to /login.
    await page.getByRole("button", { name: "Sign out" }).click();
    await expect(page).toHaveURL(/\/login/);

    // Protected routes are now inaccessible.
    await page.goto("/resources");
    await expect(page).toHaveURL(/\/login/);
  });
});
