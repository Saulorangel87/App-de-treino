'use client';

import * as React from 'react';
import { Eye, EyeOff } from 'lucide-react';
import { defineMessages } from '@/lib/i18n';
import { cn } from '@/lib/utils';
import { useMessages } from '../locale-provider';

const messages = defineMessages({
  pt: { show: 'Mostrar senha', hide: 'Ocultar senha' },
  en: { show: 'Show password', hide: 'Hide password' },
});

// Campo de senha com o botão de olho para conferir o que foi digitado. A senha
// volta a ficar oculta quando o formulário é enviado, para não ser salva nem
// exibida por engano na tela seguinte.
function PasswordInput({ className, ...props }: Omit<React.ComponentProps<'input'>, 'type'>) {
  const [visible, setVisible] = React.useState(false);
  const t = useMessages(messages);
  const ref = React.useRef<HTMLInputElement>(null);

  React.useEffect(() => {
    const form = ref.current?.form;
    if (!form) return;
    const hide = () => setVisible(false);
    form.addEventListener('submit', hide);
    return () => form.removeEventListener('submit', hide);
  }, []);

  return (
    <span className="password-field">
      {/* Visível, o campo vira texto e o teclado do celular passaria a pôr
          maiúscula, corrigir e sugerir palavras, alterando a senha digitada. */}
      <input
        ref={ref}
        type={visible ? 'text' : 'password'}
        className={cn(className)}
        autoCapitalize="none"
        autoCorrect="off"
        spellCheck={false}
        {...props}
      />
      <button
        type="button"
        className="password-toggle"
        onClick={() => setVisible((value) => !value)}
        aria-label={visible ? t.hide : t.show}
        aria-pressed={visible}
        aria-controls={props.id}
      >
        {visible ? <EyeOff size={18} aria-hidden="true" /> : <Eye size={18} aria-hidden="true" />}
      </button>
    </span>
  );
}

export { PasswordInput };
