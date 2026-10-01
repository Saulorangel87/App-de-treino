import { describe, expect, it } from 'vitest';
import {
  LEGAL_CONTACT_EMAIL,
  LEGAL_VERSION,
  PRIVACY_POLICY,
  TERMS_OF_USE,
} from './legal';

describe('textos jurídicos', () => {
  for (const document of [PRIVACY_POLICY, TERMS_OF_USE]) {
    it(`${document.title}: seções com id único e conteúdo`, () => {
      const ids = document.sections.map((section) => section.id);
      expect(new Set(ids).size).toBe(ids.length);
      for (const section of document.sections) {
        const content =
          (section.paragraphs?.length ?? 0) + (section.items?.length ?? 0);
        expect(content, section.id).toBeGreaterThan(0);
      }
    });
  }

  it('a política cobre o que a LGPD exige e nomeia os operadores reais', () => {
    const text = JSON.stringify(PRIVACY_POLICY);
    for (const required of [
      'controlador',
      'Resend',
      'Cloudflare',
      'Baixar minha planilha',
      'Encerrar conta',
      'ANPD',
      LEGAL_CONTACT_EMAIL,
    ]) {
      expect(text, required).toContain(required);
    }
  });

  it('a versão é uma data ISO', () => {
    expect(LEGAL_VERSION).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });
});
