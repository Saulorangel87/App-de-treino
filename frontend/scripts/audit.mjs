#!/usr/bin/env node

/**
 * Auditoria de dependências de produção do frontend.
 *
 * Executa `npm audit --omit=dev --audit-level=high --json` e falha se houver
 * vulnerabilidades altas ou críticas, permitindo apenas exceções documentadas
 * que ainda não possuem correção oficial disponível no ecossistema npm.
 */

import { execSync } from 'node:child_process';

const ALLOWLIST = new Map([
  [
    'GHSA-vfj7-8cjw-p6xm',
    'braces <= 3.0.3: DoS por exaustão de pilha em padrões aninhados. Sem versão corrigida lançada pelos mantenedores. Dependência transitiva de vinext (vite-plugin-commonjs -> fast-glob -> micromatch -> braces), sem caminhos expostos a entrada de usuário em produção.',
  ],
]);

let stdout = '';
try {
  stdout = execSync('npm audit --omit=dev --audit-level=high --json', {
    encoding: 'utf8',
    stdio: ['pipe', 'pipe', 'pipe'],
  });
  console.log(
    'Auditoria de dependências de produção: 0 vulnerabilidades altas ou críticas encontradas.',
  );
  process.exit(0);
} catch (err) {
  stdout = err.stdout?.toString() || '';
}

let report;
try {
  report = JSON.parse(stdout);
} catch {
  console.error('Falha ao processar relatório do npm audit:\n', stdout);
  process.exit(1);
}

const vulns = report.vulnerabilities || {};
const unallowed = [];
const allowlistedFound = new Set();

function getAdvisoryIds(name, visited = new Set()) {
  if (visited.has(name)) return [];
  visited.add(name);
  const item = vulns[name];
  if (!item) return [];

  const ids = [];
  for (const v of item.via || []) {
    if (typeof v === 'object' && v.url) {
      const id = v.url.split('/').pop();
      if (id)
        ids.push({
          id,
          url: v.url,
          title: v.title,
          name: item.name,
          severity: item.severity,
        });
    } else if (typeof v === 'string') {
      ids.push(...getAdvisoryIds(v, visited));
    }
  }
  return ids;
}

for (const name of Object.keys(vulns)) {
  const item = vulns[name];
  if (item.severity !== 'high' && item.severity !== 'critical') {
    continue;
  }

  const advisories = getAdvisoryIds(name);
  for (const adv of advisories) {
    if (ALLOWLIST.has(adv.id)) {
      allowlistedFound.add(adv.id);
    } else {
      unallowed.push(adv);
    }
  }
}

if (unallowed.length > 0) {
  console.error(
    '\n[ERRO] Vulnerabilidades de produção não permitidas encontradas:',
  );
  for (const item of unallowed) {
    console.error(
      `- [${item.severity.toUpperCase()}] ${item.name}: ${item.title} (${item.url})`,
    );
  }
  process.exit(1);
}

if (allowlistedFound.size > 0) {
  console.log(
    '\n[AVISO] Auditoria de produção aprovada com exceções documentadas sem correção upstream:',
  );
  for (const id of allowlistedFound) {
    console.log(`- ${id}: ${ALLOWLIST.get(id)}`);
  }
}

process.exit(0);
