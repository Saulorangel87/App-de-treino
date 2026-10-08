import { describe, expect, it } from 'vitest';
import {
  LEGAL_CONTACT_EMAIL,
  LEGAL_VERSION,
  PRIVACY_POLICY,
  TERMS_OF_USE,
} from './legal';
import { PRIVACY_POLICY_EN, TERMS_OF_USE_EN } from './legal.en';

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

  it('a tradução em inglês tem as mesmas seções, na mesma ordem e com o mesmo tamanho', () => {
    for (const [pt, en] of [[PRIVACY_POLICY, PRIVACY_POLICY_EN], [TERMS_OF_USE, TERMS_OF_USE_EN]] as const) {
      expect(en.sections.map((section) => section.id)).toEqual(pt.sections.map((section) => section.id));
      pt.sections.forEach((section, index) => {
        const translated = en.sections[index];
        expect(translated.paragraphs?.length, section.id).toBe(section.paragraphs?.length);
        expect(translated.items?.length, section.id).toBe(section.items?.length);
        expect(translated.after?.length, section.id).toBe(section.after?.length);
      });
    }
  });

  it('a tradução da política nomeia os mesmos operadores e o contato', () => {
    const text = JSON.stringify(PRIVACY_POLICY_EN);
    for (const required of ['controller', 'Resend', 'Cloudflare', 'Download my spreadsheet', 'Delete account', 'ANPD', LEGAL_CONTACT_EMAIL]) {
      expect(text, required).toContain(required);
    }
  });
});
