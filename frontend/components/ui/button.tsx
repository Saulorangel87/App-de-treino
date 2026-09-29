import { Button as ButtonPrimitive } from '@base-ui/react/button';

import { cn } from '@/lib/utils';

// Os estilos ficam em app/styles/base.css (.btn, .btn-primary, .btn-outline…).
const variantClasses = {
  default: 'btn-primary',
  outline: 'btn-outline',
  secondary: 'btn-outline',
  ghost: 'btn-ghost',
  destructive: 'btn-danger',
  link: 'btn-link',
} as const;

const sizeClasses = {
  default: '',
  xs: 'btn-sm',
  sm: 'btn-sm',
  lg: 'btn-lg',
  icon: 'btn-icon',
  'icon-xs': 'btn-icon',
  'icon-sm': 'btn-icon',
  'icon-lg': 'btn-icon',
} as const;

type ButtonProps = ButtonPrimitive.Props & {
  variant?: keyof typeof variantClasses;
  size?: keyof typeof sizeClasses;
};

function Button({ className, variant = 'default', size = 'default', ...props }: ButtonProps) {
  return (
    <ButtonPrimitive
      data-slot="button"
      className={cn('btn', variantClasses[variant], sizeClasses[size], className)}
      {...props}
    />
  );
}

export { Button };
