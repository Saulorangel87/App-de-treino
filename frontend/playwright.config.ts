import { defineConfig } from '@playwright/test';

// Testes ponta a ponta contra o ambiente LOCAL. Antes de rodar:
//   1. docker compose up -d postgres   (na raiz) e aplique as migrações;
//   2. pwsh -NoProfile -File scripts/run-api.ps1   (API em http://localhost:8080);
//   3. npm run e2e   (este arquivo sobe o frontend com `npm run dev`).
// Usa o Chrome instalado (channel: 'chrome'); não baixa navegadores do Playwright.
export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  retries: 0,
  workers: 1,
  reporter: [['list']],
  use: {
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:3000',
    channel: 'chrome',
    trace: 'retain-on-failure',
  },
  webServer: process.env.E2E_BASE_URL
    ? undefined
    : {
        command: 'npm run dev',
        url: 'http://localhost:3000/entrar',
        reuseExistingServer: true,
        timeout: 120_000,
      },
});
