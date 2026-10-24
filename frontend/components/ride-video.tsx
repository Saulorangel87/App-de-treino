'use client';

import { useEffect, useRef } from 'react';
import { gsap } from 'gsap';
import { DrawSVGPlugin } from 'gsap/DrawSVGPlugin';
import { prefersReducedMotion } from '@/lib/motion';

// Cena da tela de acesso: vídeo original gerado por IA (ciclista fictício, sem
// pessoa real nem marcas) em loop, com a identidade do Cadência desenhada por
// cima pelo GSAP: curvas de nível no céu e a trilha vermelha correndo pela
// estrada. O quadro inteiro do vídeo aparece no centro; nas laterais, a mesma
// imagem desfocada preenche a faixa larga do computador. Com movimento reduzido
// o vídeo fica parado no primeiro quadro e nada se desenha.

const SRC = '/media/ride.mp4';

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
    const reveal = () => host.classList.add('ready');
    videos[0].addEventListener('loadeddata', reveal, { once: true });
    if (videos[0].readyState >= 2) reveal();
    if (still) return;

    gsap.registerPlugin(DrawSVGPlugin);
    const q = gsap.utils.selector(host);
    const ctx = gsap.context(() => {
      gsap
        .timeline({ delay: 0.4, defaults: { ease: 'power2.inOut' } })
        .from(q('.ride-contour'), { drawSVG: 0, duration: 2.2, stagger: 0.12 })
        .from(q('.ride-trail'), { drawSVG: 0, duration: 1.6 }, 0.6)
        .from(q('.ride-trail-dash'), { opacity: 0, duration: 0.6 }, 2);
      // A trilha tracejada corre na velocidade aparente da estrada.
      gsap.to(q('.ride-trail-dash'), { strokeDashoffset: -120, duration: 0.9, ease: 'none', repeat: -1 });
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
      videos[0].removeEventListener('loadeddata', reveal);
    };
  }, []);

  return (
    <div ref={hostRef} className="ride-scene ride-video" role="img" aria-label={label}>
      <video className="ride-video-fill" src={SRC} muted loop playsInline preload="auto" aria-hidden="true" />
      <video className="ride-video-main" src={SRC} muted loop playsInline preload="auto" aria-hidden="true" />
      <svg className="ride-video-overlay" viewBox="0 0 1000 400" preserveAspectRatio="xMidYMid slice" aria-hidden="true">
        <g className="ride-contours" fill="none" stroke="#fff" strokeOpacity="0.55">
          {RINGS.map((d, i) => (
            <path key={d} className="ride-contour" d={d} strokeWidth={i === 5 ? 1.4 : 0.8} />
          ))}
        </g>
        <path className="ride-trail" d="M-20,372 C240,366 520,370 1020,362" fill="none" stroke="#c3321f" strokeWidth="3" strokeOpacity="0.85" />
        <path
          className="ride-trail-dash"
          d="M-20,372 C240,366 520,370 1020,362"
          fill="none"
          stroke="#fff"
          strokeWidth="1.5"
          strokeDasharray="40 80"
          strokeOpacity="0.7"
        />
      </svg>
    </div>
  );
}
