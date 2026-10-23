'use client';

import { useEffect } from 'react';
import { animateMotion, cleanupMotion, refreshMotion } from '@/lib/motion';

// Observa o documento e anima cada elemento com `data-motion` quando ele é
// inserido (troca de tela, fim do carregamento, modal aberto). O que já veio
// renderizado do servidor não anima, para não piscar depois da hidratação.
export function MotionRoot() {
  useEffect(() => {
    const observer = new MutationObserver((records) => {
      const found: Element[] = [];
      let removed = false;
      for (const record of records) {
        if (record.removedNodes.length) removed = true;
        record.addedNodes.forEach((node) => {
          if (!(node instanceof Element)) return;
          if (node.matches('[data-motion]')) found.push(node);
          node.querySelectorAll('[data-motion]').forEach((el) => found.push(el));
        });
      }
      if (removed) cleanupMotion();
      found.forEach(animateMotion);
      if (found.length || removed) refreshMotion();
    });
    observer.observe(document.body, { childList: true, subtree: true });
    return () => {
      observer.disconnect();
      cleanupMotion();
    };
  }, []);

  return null;
}
