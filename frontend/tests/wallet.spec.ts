import { expect, test } from "@playwright/test";

test("registers, signs in, edits a wallet and records money", async ({ page, context }) => {
  const email = "test-" + crypto.randomUUID() + "@example.com";
  const password = "budget-test-password";
  await page.goto("/en/register");
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password").fill(password);
  await page.getByRole("button", { name: "Create account" }).click();
  await expect(page).toHaveURL(/\/en\/login\?registered=1$/);
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password").fill(password);
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page.getByRole("heading", { name: "Your wallets" })).toBeVisible();
  expect((await context.cookies()).find((cookie) => cookie.name === "fuchs_session")?.httpOnly).toBe(true);
  await page.getByRole("link", { name: "New wallet" }).click();
  await page.getByLabel("Wallet name").fill("Cash");
  await page.getByRole("button", { name: "Save" }).click();
  await page.getByRole("link", { name: "Edit wallet" }).click();
  await page.getByLabel("Wallet name").fill("Pocket");
  await page.getByRole("button", { name: "Save" }).click();
  await expect(page.getByRole("heading", { name: "Pocket" })).toBeVisible();

  const deposit = page.getByRole("region", { name: "Deposit" });
  await deposit.getByLabel(/Amount/).fill("12.35");
  await deposit.getByLabel("Description").fill("Payday");
  await deposit.getByRole("button", { name: "Deposit" }).click();
  await expect(page.getByTestId("wallet-balance")).toContainText("12.35");
  const withdraw = page.getByRole("region", { name: "Withdraw" });
  await withdraw.getByLabel(/Amount/).fill("2.10");
  await withdraw.getByLabel("Description").fill("Transport");
  await withdraw.getByRole("button", { name: "Withdraw" }).click();
  await expect(page.getByTestId("wallet-balance")).toContainText("10.25");
  await expect(page.getByText("Transport")).toBeVisible();
  await expect(page.getByText("Payday")).toBeVisible();

  const walletPath = new URL(page.url()).pathname;
  await page.goto("/en");
  await expect(async () => {
    await page.reload();
    const report = page.getByRole("region", { name: "Income and expenses" });
    await expect(report.getByRole("cell", { name: /RUB\s+12\.35/ })).toBeVisible({ timeout: 500 });
    await expect(report.getByRole("cell", { name: /RUB\s+2\.10/ })).toBeVisible({ timeout: 500 });
  }).toPass({ timeout: 25000 });
  await page.getByRole("link", { name: "Exchange rates" }).click();
  await expect(page.getByRole("heading", { name: "Exchange rates" })).toBeVisible();
  await page.getByLabel("Currency code").fill("TST");
  await page.getByLabel("Fractional digits").selectOption("0");
  await page.getByRole("button", { name: "Add currency" }).click();
  await expect(page.getByLabel("Base currency").getByRole("option", { name: "TST" })).toBeAttached();
  await page.getByRole("link", { name: "Dashboard" }).click();
  await page.getByRole("link", { name: "New wallet" }).click();
  await page.getByLabel("Wallet name").fill("Whole units");
  await page.getByLabel("Currency").selectOption("TST");
  await page.getByRole("button", { name: "Save" }).click();
  await page.getByRole("article").filter({ hasText: "Whole units" }).getByRole("link", { name: "Edit wallet" }).click();
  await page.getByRole("region", { name: "Deposit" }).getByLabel(/Amount/).fill("1.01");
  await page.getByRole("region", { name: "Deposit" }).getByRole("button", { name: "Deposit" }).click();
  await expect(page.getByText("Couldn't save your changes. Check the details and try again.")).toBeVisible();
  await page.goto(walletPath);

  await page.getByRole("link", { name: "RU", exact: true }).click();
  await expect(page).toHaveURL(/\/ru\/wallets\/.+\/edit$/);
  await expect(page.getByTestId("wallet-balance")).toContainText("10,25");
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);

  const protectedURL = page.url();
  await context.clearCookies();
  await page.goto(protectedURL);
  await expect(page).toHaveURL(/\/ru\/login$/);
});
