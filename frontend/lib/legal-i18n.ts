import type { Locale } from './i18n';
import { LEGAL_UPDATED_LABEL, PRIVACY_POLICY, TERMS_OF_USE, type LegalDocument } from './legal';
import { PRIVACY_POLICY_EN, TERMS_OF_USE_EN } from './legal.en';

// Escolhe o texto jurídico pelo idioma. O português é o que vale; o inglês é
// tradução de cortesia (legal.en.ts).

const UPDATED_LABEL_EN = 'October 1, 2026';

export function legalUpdatedLabel(locale: Locale): string {
  return locale === 'en' ? UPDATED_LABEL_EN : LEGAL_UPDATED_LABEL;
}

export function privacyPolicy(locale: Locale): LegalDocument {
  return locale === 'en' ? PRIVACY_POLICY_EN : PRIVACY_POLICY;
}

export function termsOfUse(locale: Locale): LegalDocument {
  return locale === 'en' ? TERMS_OF_USE_EN : TERMS_OF_USE;
}
