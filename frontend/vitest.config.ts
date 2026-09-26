import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vitest/config';

// Configuração separada de vite.config.ts, que carrega os plugins do
// Cloudflare/vinext e não é necessária para testes unitários.
export default defineConfig({
  resolve: { alias: { '@': fileURLToPath(new URL('./', import.meta.url)) } },
  test: {
    environment: 'node',
    include: ['**/*.test.ts'],
    exclude: ['node_modules/**', 'dist/**', 'e2e/**'],
  },
});
