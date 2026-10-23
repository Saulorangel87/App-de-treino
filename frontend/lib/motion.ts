// Movimento do app com GSAP. As telas marcam elementos com `data-motion` e o
// `MotionRoot` (components/motion-root.tsx) anima cada um quando ele entra no
// DOM. O movimento é apoio visual: nunca carrega informação, nunca atrasa uma
// ação e não roda para quem pede menos movimento no sistema.
//
// Tipos aceitos em `data-motion`:
// - stagger: os filhos diretos sobem em sequência (abaixo da dobra, ao rolar).
// - cascade: igual a stagger, mais curto, para listas e faixas de dias.
// - rise:    o próprio elemento sobe e aparece.
// - fade:    o próprio elemento aparece (fundos de modal).
// - sheet:   painel ou modal entra de baixo.
// - bars-y:  as barras `<i>` crescem a partir da base (gráficos de colunas).
// - bars-x:  as barras `<i>` crescem da esquerda (barras de progresso).
// - route:   o mapa do treino se desenha (curvas de nível, trecho a trecho).
// - tab:     o traço vermelho do item ativo da navegação se estende.

import { gsap } from 'gsap';
import { DrawSVGPlugin } from 'gsap/DrawSVGPlugin';
import { ScrollTrigger } from 'gsap/ScrollTrigger';

export type MotionKind =
  | 'stagger'
  | 'cascade'
  | 'rise'
  | 'fade'
  | 'sheet'
  | 'bars-y'
  | 'bars-x'
  | 'route'
  | 'tab';

const EASE = 'power3.out';
// Ao terminar, o estilo inline sai e o CSS volta a mandar sozinho.
const CLEAR = 'transform,opacity';
// Elemento cujo topo passa deste ponto da janela só anima ao rolar até ele.
const FOLD = 0.92;

let registered = false;
const seen = new WeakSet<Element>();

function setup() {
  if (registered) return;
  gsap.registerPlugin(ScrollTrigger, DrawSVGPlugin);
  registered = true;
}

export function prefersReducedMotion() {
  return (
    typeof window === 'undefined' ||
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  );
}

function belowFold(el: Element) {
  return el.getBoundingClientRect().top > window.innerHeight * FOLD;
}

// Toca a animação (criada pausada, já no estado inicial) agora ou quando o
// elemento entrar na tela.
function playWhenVisible(el: Element, animation: gsap.core.Animation) {
  if (!belowFold(el)) {
    animation.play();
    return;
  }
  ScrollTrigger.create({
    trigger: el,
    start: `top ${FOLD * 100}%`,
    once: true,
    onEnter: () => animation.play(),
  });
}

function reveal(targets: Element[], y: number, stagger: number, duration: number) {
  if (!targets.length) return;
  const now = targets.filter((el) => !belowFold(el));
  const later = targets.filter(belowFold);
  if (now.length) {
    gsap.from(now, { y, opacity: 0, duration, ease: EASE, stagger, clearProps: CLEAR });
  }
  if (later.length) {
    gsap.set(later, { y, opacity: 0 });
    ScrollTrigger.batch(later, {
      start: `top ${FOLD * 100}%`,
      once: true,
      onEnter: (batch) =>
        gsap.to(batch, { y: 0, opacity: 1, duration, ease: EASE, stagger, clearProps: CLEAR }),
    });
  }
}

function bars(el: Element, axis: 'x' | 'y') {
  const items = el.querySelectorAll('i');
  if (!items.length) return;
  const tween = gsap.from(items, {
    paused: true,
    [axis === 'y' ? 'scaleY' : 'scaleX']: 0,
    transformOrigin: axis === 'y' ? '50% 100%' : '0% 50%',
    duration: 0.8,
    ease: EASE,
    stagger: 0.06,
    clearProps: 'transform',
  });
  playWhenVisible(el, tween);
}

function route(svg: Element) {
  const tl = gsap.timeline({ paused: true, defaults: { ease: EASE } });
  tl.from(svg.querySelectorAll('.route-contour'), {
    drawSVG: 0,
    duration: 1.2,
    stagger: 0.03,
    ease: 'power1.inOut',
  }, 0)
    .from(svg.querySelectorAll('.route-forest, .route-water'), { opacity: 0, duration: 0.8 }, 0)
    .from(svg.querySelectorAll('.route-start'), {
      scale: 0,
      transformOrigin: '50% 50%',
      duration: 0.35,
      ease: 'back.out(3)',
    }, 0.3);

  // Cada trecho é desenhado na ordem, com tempo proporcional ao comprimento.
  const lines = Array.from(svg.querySelectorAll<SVGGeometryElement>('.route-line'));
  const lengths = lines.map((line) => line.getTotalLength?.() || 1);
  const total = lengths.reduce((sum, length) => sum + length, 0) || 1;
  const drawTime = 1.2;
  let at = 0.45;
  tl.from(svg.querySelectorAll('.route-halo'), { drawSVG: 0, duration: drawTime, ease: 'none' }, at);
  lines.forEach((line, index) => {
    const duration = (lengths[index] / total) * drawTime;
    tl.from(line, { drawSVG: 0, duration, ease: 'none' }, at);
    at += duration;
  });
  tl.from(svg.querySelectorAll('.route-marker'), {
    scale: 0,
    transformOrigin: '50% 50%',
    duration: 0.35,
    ease: 'back.out(2.5)',
    stagger: 0.08,
  }, 0.6)
    .from(svg.querySelectorAll('.route-finish'), {
      scaleY: 0,
      transformOrigin: '0% 100%',
      duration: 0.4,
      ease: 'back.out(2)',
    }, at);
  playWhenVisible(svg, tl);
}

// O traço é um box-shadow inset; zera os deslocamentos e volta ao valor do CSS.
function tab(el: Element) {
  const shadow = getComputedStyle(el).boxShadow;
  if (!shadow || shadow === 'none') return;
  gsap.fromTo(
    el,
    { boxShadow: shadow.replace(/-?\d+(\.\d+)?px/g, '0px') },
    { boxShadow: shadow, duration: 0.45, ease: EASE, clearProps: 'boxShadow' },
  );
}

export function animateMotion(el: Element) {
  if (seen.has(el)) return;
  seen.add(el);
  if (prefersReducedMotion()) return;
  setup();
  const kind = el.getAttribute('data-motion') as MotionKind | null;
  switch (kind) {
    case 'stagger':
      reveal(Array.from(el.children), 24, 0.08, 0.6);
      break;
    case 'cascade':
      reveal(Array.from(el.children), 12, 0.04, 0.45);
      break;
    case 'rise':
      reveal([el], 24, 0, 0.6);
      break;
    case 'fade':
      gsap.from(el, { opacity: 0, duration: 0.3, ease: 'power1.out', clearProps: 'opacity' });
      break;
    case 'sheet':
      gsap.from(el, { y: 48, opacity: 0, duration: 0.45, ease: EASE, clearProps: CLEAR });
      break;
    case 'bars-y':
      bars(el, 'y');
      break;
    case 'bars-x':
      bars(el, 'x');
      break;
    case 'route':
      route(el);
      break;
    case 'tab':
      tab(el);
      break;
  }
}

// Repete a entrada de um elemento que já estava no DOM e só passou a aparecer
// (painel do treino no celular, aberto por classe).
export function replayMotion(el: Element | null) {
  if (!el || prefersReducedMotion()) return;
  setup();
  gsap.from(el, { y: 48, opacity: 0, duration: 0.45, ease: EASE, clearProps: CLEAR });
}

// Remove gatilhos de rolagem de elementos que saíram da tela (troca de rota).
export function cleanupMotion() {
  if (!registered) return;
  ScrollTrigger.getAll().forEach((trigger) => {
    const target = trigger.trigger;
    if (target && !target.isConnected) trigger.kill();
  });
}

export function refreshMotion() {
  if (registered && ScrollTrigger.getAll().length) ScrollTrigger.refresh();
}
