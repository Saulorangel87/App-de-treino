import { expect, test, type Page } from '@playwright/test';

const API_URL = process.env.E2E_API_URL || 'http://localhost:8080';
const ORIGIN = process.env.E2E_BASE_URL || 'http://localhost:3000';

const password = 'senha-e2e-inicial-1';
const newPassword = 'senha-e2e-nova-22';

// O aviso de novidades aparece uma vez por conta e versão e cobre a página.
async function dismissUpdateNotice(page: Page) {
  await page
    .getByRole('button', { name: 'Fechar novidades' })
    .click({ timeout: 3_000 })
    .catch(() => undefined);
}

function uniqueEmail() {
  return `e2e-${Date.now()}-${Math.random().toString(36).slice(2, 8)}@example.invalid`;
}

// Remove a conta de teste pela própria API (mesmo caminho do botão "Encerrar conta").
async function deleteAccount(page: Page, currentPassword: string) {
  // Limpeza best-effort: nunca deve mascarar a falha real do teste.
  await page.request
    .delete(`${API_URL}/v1/auth/account`, {
      headers: { Origin: ORIGIN },
      data: { password: currentPassword, confirmation: 'ENCERRAR CONTA' },
      timeout: 10_000,
    })
    .catch(() => undefined);
}

test('cadastro, confirmação de e-mail, troca de senha e novo login', async ({ page }) => {
  const email = uniqueEmail();
  let activePassword = password;

  try {
    // Cadastro: em desenvolvimento a API devolve o link de confirmação.
    await page.goto('/entrar');
    // Espera a hidratação: antes dela o formulário ainda seria enviado pelo navegador.
    await page.waitForLoadState('networkidle');
    await page.getByLabel('Como podemos chamar você?').fill('Atleta E2E');
    await page.getByLabel('E-mail', { exact: true }).fill(email);
    await page.getByLabel('Senha', { exact: true }).fill(password);
    await page.getByRole('button', { name: 'Criar minha conta' }).click();
    await expect(page.getByRole('status')).toContainText('confirmar seu e-mail');

    await page.getByRole('link', { name: 'Abrir confirmação local' }).click();
    await expect(page.getByRole('heading', { name: 'E-mail confirmado.' })).toBeVisible();

    // Configurações mostra o e-mail como confirmado.
    await page.goto('/configuracoes');
    await page.waitForLoadState('networkidle');
    await dismissUpdateNotice(page);
    await expect(page.getByText(email)).toBeVisible();
    await expect(page.getByText('Confirmado')).toBeVisible();

    // Senha atual incorreta é recusada e nada muda.
    await page.getByLabel('Senha atual').first().fill('senha-errada-123');
    await page.getByLabel('Nova senha', { exact: true }).fill(newPassword);
    await page.getByLabel('Confirmar nova senha', { exact: true }).fill(newPassword);
    await page.getByRole('button', { name: 'Alterar senha' }).click();
    await expect(page.getByRole('alert')).toContainText('senha atual está incorreta');

    // Troca válida mantém a sessão atual.
    await page.getByLabel('Senha atual').first().fill(password);
    await page.getByRole('button', { name: 'Alterar senha' }).click();
    await expect(page.getByRole('status')).toContainText('Senha alterada');
    activePassword = newPassword;
    await page.goto('/configuracoes');
    await page.waitForLoadState('networkidle');
    await dismissUpdateNotice(page);
    await expect(page.getByText(email)).toBeVisible();

    // Encerrar outras sessões responde sem erro.
    await page.getByRole('button', { name: 'Sair dos outros dispositivos' }).click();
    await expect(page.getByRole('status')).toBeVisible();

    // Login com a senha antiga falha; com a nova, funciona.
    await page.context().clearCookies();
    await page.goto('/entrar');
    await page.waitForLoadState('networkidle');
    await page.getByRole('button', { name: 'Entrar' }).first().click();
    await page.getByLabel('E-mail', { exact: true }).fill(email);
    await page.getByLabel('Senha', { exact: true }).fill(password);
    await page.getByRole('button', { name: 'Entrar' }).last().click();
    await expect(page.getByRole('alert')).toContainText('E-mail ou senha inválidos');

    await page.getByLabel('Senha', { exact: true }).fill(newPassword);
    await page.getByRole('button', { name: 'Entrar' }).last().click();
    await page.waitForURL(`${ORIGIN}/`);
  } finally {
    await deleteAccount(page, activePassword);
  }
});

test('sessão expirada em uma rota protegida leva ao login', async ({ page }) => {
  await page.context().clearCookies();
  await page.goto('/plano');
  await page.waitForURL('**/entrar');
});
