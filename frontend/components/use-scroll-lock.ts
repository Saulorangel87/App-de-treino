import { useEffect } from 'react';

// Trava a rolagem da página enquanto um modal está aberto, compensando a
// largura da barra de rolagem para o conteúdo não "pular" no desktop.
export function useScrollLock(active: boolean) {
  useEffect(() => {
    if (!active) return;

    const root = document.documentElement;
    const body = document.body;
    const previousRootOverflow = root.style.overflow;
    const previousBodyOverflow = body.style.overflow;
    const previousBodyPaddingRight = body.style.paddingRight;
    const scrollbarWidth = window.innerWidth - root.clientWidth;

    root.style.overflow = 'hidden';
    body.style.overflow = 'hidden';
    if (scrollbarWidth > 0) {
      body.style.paddingRight = `${scrollbarWidth}px`;
    }

    return () => {
      root.style.overflow = previousRootOverflow;
      body.style.overflow = previousBodyOverflow;
      body.style.paddingRight = previousBodyPaddingRight;
    };
  }, [active]);
}
