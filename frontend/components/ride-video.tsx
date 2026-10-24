'use client';

import { useEffect, useRef } from 'react';
import { gsap } from 'gsap';
import { DrawSVGPlugin } from 'gsap/DrawSVGPlugin';
import { prefersReducedMotion } from '@/lib/motion';

// Cena da tela de acesso: vídeo original gerado por IA (ciclista fictício, sem
// pessoa real nem marcas) em loop, com a identidade do Cadência desenhada por
// cima pelo GSAP: curvas de nível que se desenham no céu. O quadro inteiro do
// vídeo aparece no centro; nas laterais, a mesma imagem desfocada preenche a
// faixa larga do computador. No celular vem uma versão menor do arquivo. Com
// movimento reduzido o vídeo fica parado no primeiro quadro e nada se desenha.

const SRC = '/media/ride.mp4';
const SRC_MOBILE = '/media/ride-mobile.mp4';

// Curvas de nível: anéis irregulares em volta de um cume, no canto do céu.
function contour(cx: number, cy: number, r: number, wobble: number, phase: number) {
  const points: string[] = [];
  for (let i = 0; i <= 48; i += 1) {
    const a = (i / 48) * Math.PI * 2;
    const k = r * (1 + wobble * Math.sin(a * 3 + phase) + wobble * 0.6 * Math.cos(a * 5 - phase));
    points.push(`${(cx + Math.cos(a) * k * 1.6).toFixed(1)},${(cy + Math.sin(a) * k).toFixed(1)}`);
  }
  return `M${points.join(' L')} Z`;
}

const RINGS = [18, 34, 50, 66, 82, 98].map((r, i) => contour(860, 70, r, 0.08 + i * 0.012, i * 0.7));

export function RideVideo({ label }: { label: string }) {
  const hostRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const host = hostRef.current;
    if (!host) return;
    const videos = Array.from(host.querySelectorAll('video'));
    const still = prefersReducedMotion();
    // No celular o vídeo cobre a faixa toda: o fundo desfocado nem carrega.
    const mobile = window.matchMedia('(max-width: 959px)').matches;
    const [fill, main] = videos;
    main.src = mobile ? SRC_MOBILE : SRC;
    if (mobile) videos.shift();
    else fill.src = SRC;
    const reveal = () => host.classList.add('ready');
    main.addEventListener('loadeddata', reveal, { once: true });
    if (main.readyState >= 2) reveal();
    if (still) return;

    gsap.registerPlugin(DrawSVGPlugin);
    const q = gsap.utils.selector(host);
    const ctx = gsap.context(() => {
      gsap
        .timeline({ delay: 0.4, defaults: { ease: 'power2.inOut' } })
        .from(q('.ride-contour'), { drawSVG: 0, duration: 2.2, stagger: 0.12 });
      gsap.to(q('.ride-contours'), { x: -30, duration: 20, ease: 'sine.inOut', yoyo: true, repeat: -1 });
    }, host);

    // Fora da tela ou com a aba escondida, o vídeo para e não gasta bateria.
    const play = () => videos.forEach((video) => void video.play().catch(() => undefined));
    const pause = () => videos.forEach((video) => video.pause());
    let visible = true;
    const io = new IntersectionObserver(([entry]) => {
      visible = entry.isIntersecting;
      if (visible && !document.hidden) play();
      else pause();
    });
    io.observe(host);
    const onVisibility = () => (document.hidden || !visible ? pause() : play());
    document.addEventListener('visibilitychange', onVisibility);

    return () => {
      ctx.revert();
      io.disconnect();
      document.removeEventListener('visibilitychange', onVisibility);
      main.removeEventListener('loadeddata', reveal);
    };
  }, []);

  return (
    <div ref={hostRef} className="ride-scene ride-video" role="img" aria-label={label}>
      <video className="ride-video-fill" muted loop playsInline preload="auto" aria-hidden="true" />
      <video className="ride-video-main" muted loop playsInline preload="auto" aria-hidden="true" />
      <svg className="ride-video-overlay" viewBox="0 0 1000 400" preserveAspectRatio="xMidYMid slice" aria-hidden="true">
        <g className="ride-contours" fill="none" stroke="#fff" strokeOpacity="0.55">
          {RINGS.map((d, i) => (
            <path key={d} className="ride-contour" d={d} strokeWidth={i === 5 ? 1.4 : 0.8} />
          ))}
        </g>
      </svg>
    </div>
  );
}
