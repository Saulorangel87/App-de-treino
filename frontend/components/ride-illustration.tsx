'use client';

import { useEffect, useRef } from 'react';
import { gsap } from 'gsap';
import { prefersReducedMotion } from '@/lib/motion';

// Ilustração animada da tela de acesso, alternativa em SVG à cena 3D
// (components/ride-scene.tsx): um ciclista do Cadência de perfil, pedalando
// numa estrada com a faixa vermelha, diante de serras desenhadas como curvas de
// nível. Cada plano corre numa velocidade (paralaxe), as rodas e o pedivela
// giram juntos e as pernas seguem os pedais por cinemática inversa. Tudo num
// timeline do GSAP, que para fora da tela e não roda com movimento reduzido.

const W = 800; // período horizontal das camadas que se repetem
const H = 260;
const SPEED = 170; // px/s da estrada
const WHEEL_R = 44;
const CRANK_R = 23;
const GEAR = 1.6;

const REAR = { x: 352, y: 184 };
const FRONT = { x: 480, y: 184 };
const BB = { x: 412, y: 190 };
const SEAT = { x: 390, y: 124 };
const HEAD_TOP = { x: 466, y: 118 };
const HEAD_LOW = { x: 472, y: 142 };
const BAR = { x: 482, y: 110 };
const HIP = { x: 384, y: 104 };
const SHOULDER = { x: 452, y: 74 };
const HAND = { x: 499, y: 109 };
const THIGH = 68;
const SHIN = 66;
const UPPER_ARM = 36;
const FOREARM = 37;

type Point = { x: number; y: number };

// Perfil de serra que se repete a cada W px (ciclos inteiros por período).
function ridge(x: number, base: number, amp: number, waves: Array<[number, number, number]>) {
  return base - amp * waves.reduce((sum, [k, a, phase]) => sum + a * Math.sin((2 * Math.PI * k * x) / W + phase), 0);
}

function ridgePath(base: number, amp: number, waves: Array<[number, number, number]>, closed: boolean) {
  const points: string[] = [];
  for (let x = 0; x <= W * 2; x += 8) points.push(`${x},${ridge(x, base, amp, waves).toFixed(1)}`);
  return closed ? `M0,${H} L${points.join(' L')} L${W * 2},${H} Z` : `M${points.join(' L')}`;
}

const LAYERS = [
  {
    key: 'far',
    speed: 0.06,
    fill: 'var(--ride-far)',
    line: 'var(--ride-far-line)',
    base: 150,
    amp: 26,
    waves: [
      [1, 0.6, 0.4],
      [2, 0.5, 2.1],
      [5, 0.18, 0.9],
    ] as Array<[number, number, number]>,
  },
  {
    key: 'mid',
    speed: 0.18,
    fill: 'var(--ride-mid)',
    line: 'var(--ride-mid-line)',
    base: 186,
    amp: 20,
    waves: [
      [2, 0.6, 1.2],
      [3, 0.4, 0.2],
      [7, 0.12, 2.4],
    ] as Array<[number, number, number]>,
  },
  {
    key: 'near',
    speed: 0.45,
    fill: 'var(--ride-near)',
    line: 'var(--ride-near-line)',
    base: 214,
    amp: 8,
    waves: [
      [3, 0.6, 0.5],
      [8, 0.3, 1.7],
    ] as Array<[number, number, number]>,
  },
];

// Dois segmentos de a até b; `bend` escolhe o lado da articulação.
function solve(a: Point, b: Point, first: number, second: number, bend: 1 | -1): Point {
  const dx = b.x - a.x;
  const dy = b.y - a.y;
  const dist = Math.min(Math.hypot(dx, dy), first + second - 0.01);
  const cos = (first * first + dist * dist - second * second) / (2 * first * dist);
  const angle = Math.atan2(dy, dx) + bend * Math.acos(Math.max(-1, Math.min(1, cos)));
  return { x: a.x + Math.cos(angle) * first, y: a.y + Math.sin(angle) * first };
}

function setLine(line: SVGLineElement | null, a: Point, b: Point) {
  if (!line) return;
  line.setAttribute('x1', a.x.toFixed(2));
  line.setAttribute('y1', a.y.toFixed(2));
  line.setAttribute('x2', b.x.toFixed(2));
  line.setAttribute('y2', b.y.toFixed(2));
}

const SPOKES = Array.from({ length: 12 }, (_, i) => (i / 12) * Math.PI * 2);

function Wheel({ at }: { at: Point }) {
  return (
    <g className="ride-wheel" transform={`translate(${at.x} ${at.y})`}>
      <g data-spin>
        <circle r={WHEEL_R - 2} fill="none" stroke="var(--ride-ink)" strokeWidth="5" />
        <circle r={WHEEL_R - 7} fill="none" stroke="var(--ride-ink)" strokeWidth="6" />
        <path
          d={`M0,${-(WHEEL_R - 7)} A${WHEEL_R - 7},${WHEEL_R - 7} 0 0 1 ${WHEEL_R - 7},0`}
          fill="none"
          stroke="var(--ride-red)"
          strokeWidth="6"
        />
        {SPOKES.map((angle) => (
          <line
            key={angle}
            x1="0"
            y1="0"
            x2={(Math.cos(angle) * (WHEEL_R - 10)).toFixed(2)}
            y2={(Math.sin(angle) * (WHEEL_R - 10)).toFixed(2)}
            stroke="var(--ride-spoke)"
            strokeWidth="0.9"
          />
        ))}
      </g>
      <circle r="4" fill="var(--ride-metal)" />
    </g>
  );
}

function Limb({ name }: { name: 'near' | 'far' }) {
  return (
    <g className={`ride-leg ride-leg-${name}`}>
      <line data-part="thigh" stroke="var(--ride-bib)" strokeWidth="19" strokeLinecap="round" />
      <line data-part="shin" stroke="var(--ride-skin)" strokeWidth="12" strokeLinecap="round" />
      <line data-part="sock" stroke="var(--ride-white)" strokeWidth="11" strokeLinecap="round" />
      <g data-part="shoe">
        <path d="M-12,-5 Q-12,-11 -4,-11 L8,-8 Q17,-6 17,0 L17,2 L-12,2 Z" fill="var(--ride-white)" />
        <rect x="-12" y="1" width="29" height="3" rx="1.5" fill="var(--ride-ink)" />
        <circle cx="1" cy="-7" r="2.2" fill="var(--ride-red)" />
      </g>
    </g>
  );
}

function Arm({ name }: { name: 'near' | 'far' }) {
  return (
    <g className={`ride-arm ride-arm-${name}`}>
      <line data-part="upper" stroke="var(--ride-red)" strokeWidth="13" strokeLinecap="round" />
      <line data-part="fore" stroke="var(--ride-skin)" strokeWidth="9" strokeLinecap="round" />
      <circle data-part="glove" r="6" fill="var(--ride-ink)" />
    </g>
  );
}

export function RideIllustration({ label }: { label: string }) {
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    const svg = svgRef.current;
    if (!svg) return;
    const q = gsap.utils.selector(svg);
    const legs = (['near', 'far'] as const).map((name) => {
      const root = svg.querySelector(`.ride-leg-${name}`)!;
      return {
        thigh: root.querySelector<SVGLineElement>('[data-part="thigh"]'),
        shin: root.querySelector<SVGLineElement>('[data-part="shin"]'),
        sock: root.querySelector<SVGLineElement>('[data-part="sock"]'),
        shoe: root.querySelector<SVGGElement>('[data-part="shoe"]'),
      };
    });
    const arms = (['near', 'far'] as const).map((name) => {
      const root = svg.querySelector(`.ride-arm-${name}`)!;
      return {
        upper: root.querySelector<SVGLineElement>('[data-part="upper"]'),
        fore: root.querySelector<SVGLineElement>('[data-part="fore"]'),
        glove: root.querySelector<SVGCircleElement>('[data-part="glove"]'),
      };
    });
    const crankArms = q('[data-crank]');
    const pedals = q('[data-pedal]');
    const body = q('.ride-body');

    // Pose do ciclista para um ângulo de pedivela (radianos).
    const state = { crank: 0.4 };
    function pose() {
      const angle = state.crank;
      const bob = Math.sin(angle * 2) * 0.8;
      gsap.set(body, { y: bob });
      legs.forEach((leg, index) => {
        const a = angle + index * Math.PI;
        const foot = { x: BB.x + Math.cos(a) * CRANK_R, y: BB.y + Math.sin(a) * CRANK_R };
        const ankle = { x: foot.x - 5, y: foot.y - 9 };
        const hip = { x: HIP.x + (index ? -3 : 0), y: HIP.y + bob };
        const knee = solve(hip, ankle, THIGH, SHIN, -1);
        setLine(leg.thigh, hip, knee);
        setLine(leg.shin, knee, ankle);
        setLine(leg.sock, { x: ankle.x + (knee.x - ankle.x) * 0.12, y: ankle.y + (knee.y - ankle.y) * 0.12 }, ankle);
        leg.shoe?.setAttribute(
          'transform',
          `translate(${foot.x.toFixed(2)} ${(foot.y - 1).toFixed(2)}) rotate(${(Math.sin(a) * 12 - 6).toFixed(1)})`,
        );
        crankArms[index]?.setAttribute('x2', foot.x.toFixed(2));
        crankArms[index]?.setAttribute('y2', foot.y.toFixed(2));
        pedals[index]?.setAttribute('transform', `translate(${foot.x.toFixed(2)} ${foot.y.toFixed(2)})`);
      });
      arms.forEach((arm, index) => {
        const shoulder = { x: SHOULDER.x - index * 4, y: SHOULDER.y + bob };
        const hand = { x: HAND.x - index * 3, y: HAND.y };
        const elbow = solve(shoulder, hand, UPPER_ARM, FOREARM, 1);
        setLine(arm.upper, shoulder, elbow);
        setLine(arm.fore, elbow, hand);
        arm.glove?.setAttribute('cx', hand.x.toFixed(2));
        arm.glove?.setAttribute('cy', hand.y.toFixed(2));
      });
    }
    pose();
    if (prefersReducedMotion()) return;

    // Uma volta do pedivela por ciclo; o resto se repete em múltiplos dele.
    const wheelTurn = (2 * Math.PI * WHEEL_R) / SPEED; // s por volta da roda
    const ctx = gsap.context(() => {
      const loop = gsap.timeline({ repeat: -1, defaults: { ease: 'none' } });
      loop.to(state, { crank: state.crank + Math.PI * 2, duration: wheelTurn * GEAR, onUpdate: pose, repeat: -1 }, 0);
      loop.to(q('[data-spin]'), { rotation: 360, transformOrigin: '50% 50%', duration: wheelTurn, repeat: -1 }, 0);
      LAYERS.forEach((layer) => {
        loop.to(q(`.ride-layer-${layer.key}`), { x: -W, duration: W / (SPEED * layer.speed), repeat: -1 }, 0);
      });
      loop.to(q('.ride-road-paint'), { x: -W, duration: W / SPEED, repeat: -1 }, 0);
      loop.to(q('.ride-grass'), { x: -W, duration: W / (SPEED * 1.35), repeat: -1 }, 0);
      loop.to(q('.ride-clouds'), { x: -W, duration: 140, repeat: -1 }, 0);

      // Entrada: as serras sobem em sequência e o ciclista chega pela esquerda.
      gsap
        .timeline({ defaults: { ease: 'power3.out' } })
        .from(q('.ride-depth'), { y: 40, opacity: 0, duration: 1, stagger: 0.12 })
        .from(q('.ride-rider'), { x: -260, duration: 1.4, ease: 'power2.out' }, 0.2);

      const io = new IntersectionObserver(([entry]) => {
        if (entry.isIntersecting && !document.hidden) loop.play();
        else loop.pause();
      });
      io.observe(svg);
      const onVisibility = () => (document.hidden ? loop.pause() : loop.play());
      document.addEventListener('visibilitychange', onVisibility);
      return () => {
        io.disconnect();
        document.removeEventListener('visibilitychange', onVisibility);
      };
    }, svg);
    return () => ctx.revert();
  }, []);

  return (
    <div className="ride-scene ride-illustration ready" role="img" aria-label={label}>
      <svg ref={svgRef} viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="xMidYMax slice" aria-hidden="true">
        <defs>
          <linearGradient id="ride-sky" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0" stopColor="var(--ride-sky-top)" />
            <stop offset="1" stopColor="var(--ride-sky-low)" />
          </linearGradient>
          <radialGradient id="ride-sun">
            <stop offset="0" stopColor="var(--ride-sun)" stopOpacity="0.9" />
            <stop offset="0.45" stopColor="var(--ride-sun)" stopOpacity="0.55" />
            <stop offset="1" stopColor="var(--ride-sun)" stopOpacity="0" />
          </radialGradient>
        </defs>

        <rect width={W} height={H} fill="url(#ride-sky)" />
        <circle cx="610" cy="70" r="70" fill="url(#ride-sun)" />
        <circle cx="610" cy="70" r="22" fill="var(--ride-sun)" />

        <g className="ride-clouds" fill="var(--ride-cloud)">
          {[0, W].map((shift) => (
            <g key={shift} transform={`translate(${shift} 0)`}>
              <path d="M90,62 q10,-18 30,-10 q12,-14 30,-2 q18,0 16,12 z" />
              <path d="M420,40 q8,-14 24,-8 q10,-12 26,-1 q14,1 12,9 z" />
              <path d="M700,78 q10,-16 28,-9 q12,-12 26,0 q14,1 12,9 z" />
            </g>
          ))}
        </g>

        {LAYERS.map((layer) => (
          <g key={layer.key} className="ride-depth">
            <g className={`ride-layer-${layer.key}`}>
              <path d={ridgePath(layer.base, layer.amp, layer.waves, true)} fill={layer.fill} />
              {/* Curvas de nível: a mesma serra, mais baixa e mais suave a cada linha. */}
              {[1, 2, 3, 4, 5].map((step) => (
                <path
                  key={step}
                  d={ridgePath(layer.base + step * 9, layer.amp * (1 - step * 0.14), layer.waves, false)}
                  fill="none"
                  stroke={layer.line}
                  strokeWidth={step === 5 ? 1.4 : 0.8}
                />
              ))}
            </g>
          </g>
        ))}

        {/* Capim na beira da estrada, atrás do asfalto. */}
        <g className="ride-grass" stroke="var(--ride-grass)" strokeWidth="2" strokeLinecap="round" fill="none">
          {Array.from({ length: 36 }, (_, i) => {
            // 18 tufos por período, repetidos uma vez para o laço não emendar.
            const j = i % 18;
            const x = j * 44.4 + ((j * 7) % 13) + (i >= 18 ? W : 0);
            return (
              <path key={i} d={`M${x},224 q-2,-10 -6,-${14 + (j % 4) * 4} M${x + 4},224 q1,-9 5,-${12 + (j % 3) * 5}`} />
            );
          })}
        </g>
        {/* Estrada */}
        <g className="ride-depth">
          <rect y="222" width={W} height="38" fill="var(--ride-road)" />
          <rect y="222" width={W} height="2.5" fill="var(--ride-white)" opacity="0.85" />
          <rect y="226" width={W} height="34" fill="var(--ride-road-shade)" opacity="0.35" />
          <g className="ride-road-paint">
            <line x1="0" y1="242" x2={W * 2} y2="242" stroke="var(--ride-red)" strokeWidth="3.5" strokeDasharray="44 36" />
            <line x1="0" y1="257" x2={W * 2} y2="257" stroke="var(--ride-white)" strokeWidth="2" opacity="0.8" />
          </g>
        </g>

        <g className="ride-rider">
          <ellipse cx="416" cy="230" rx="105" ry="5" fill="var(--ride-ink)" opacity="0.22" />

          {/* Lado de lá: perna e braço mais escuros, atrás do quadro. */}
          <g className="ride-far">
            <line data-crank x1={BB.x} y1={BB.y} stroke="var(--ride-metal-dark)" strokeWidth="5" strokeLinecap="round" />
            <rect data-pedal x="-7" y="-2" width="14" height="4" rx="1.5" fill="var(--ride-ink)" />
          </g>
          <g className="ride-body">
            <g className="ride-shade">
              <Limb name="far" />
              <Arm name="far" />
            </g>
          </g>

          <Wheel at={REAR} />
          <Wheel at={FRONT} />

          {/* Quadro */}
          <g fill="none" strokeLinecap="round" strokeLinejoin="round">
            <path d={`M${REAR.x},${REAR.y} L${BB.x},${BB.y} L${SEAT.x},${SEAT.y} Z`} stroke="var(--ride-ink)" strokeWidth="5" />
            <path d={`M${SEAT.x},${SEAT.y} L${HEAD_TOP.x},${HEAD_TOP.y}`} stroke="var(--ride-red)" strokeWidth="7" />
            <path d={`M${BB.x},${BB.y} L${HEAD_LOW.x},${HEAD_LOW.y}`} stroke="var(--ride-ink)" strokeWidth="8" />
            <path d={`M${HEAD_TOP.x},${HEAD_TOP.y} L${HEAD_LOW.x},${HEAD_LOW.y}`} stroke="var(--ride-red)" strokeWidth="8" />
            <path d={`M${HEAD_LOW.x},${HEAD_LOW.y} Q${HEAD_LOW.x + 4},${HEAD_LOW.y + 22} ${FRONT.x},${FRONT.y}`} stroke="var(--ride-ink)" strokeWidth="5" />
            <path d={`M${SEAT.x},${SEAT.y} L${SEAT.x - 4},${SEAT.y - 14}`} stroke="var(--ride-ink)" strokeWidth="4" />
            <path d={`M${HEAD_TOP.x},${HEAD_TOP.y} L${BAR.x},${BAR.y}`} stroke="var(--ride-ink)" strokeWidth="5" />
            <path
              d={`M${BAR.x},${BAR.y} q16,-1 18,10 q2,12 -10,16`}
              stroke="var(--ride-ink)"
              strokeWidth="4.5"
            />
            <path d={`M${BB.x + 4},${BB.y - 26} L${HEAD_LOW.x - 20},${HEAD_LOW.y - 5}`} stroke="var(--ride-bottle)" strokeWidth="9" />
          </g>
          <path d={`M${SEAT.x - 22},${SEAT.y - 16} q14,-6 34,1 l-2,3 q-16,-2 -32,0 z`} fill="var(--ride-ink)" />
          <circle cx={BB.x} cy={BB.y} r="12" fill="none" stroke="var(--ride-metal-dark)" strokeWidth="3" />
          <path
            d={`M${BB.x},${BB.y - 12} L${REAR.x},${REAR.y - 7} M${BB.x},${BB.y + 12} L${REAR.x},${REAR.y + 7}`}
            stroke="var(--ride-metal-dark)"
            strokeWidth="1.6"
          />

          <g className="ride-body">
            {/* Tronco arqueado: bermuda, camisa vermelha com a faixa branca. */}
            <path
              d="M372,112 C374,92 392,76 418,68 C434,63 452,62 462,68 C466,76 462,86 452,90 C436,96 418,104 404,114 C394,121 378,124 372,112 Z"
              fill="var(--ride-red)"
            />
            <path d="M372,112 C373,104 378,96 386,91 L401,114 C392,121 378,124 372,112 Z" fill="var(--ride-bib)" />
            <path d="M440,64 C446,72 448,82 446,92" fill="none" stroke="var(--ride-white)" strokeWidth="6" />
            <path d="M398,80 C410,74 424,70 434,68" fill="none" stroke="var(--ride-red-dark)" strokeWidth="2" opacity="0.6" />
            {/* Cabeça, capacete e óculos. */}
            <path d="M458,70 L468,62" stroke="var(--ride-skin)" strokeWidth="9" strokeLinecap="round" />
            <circle cx="476" cy="57" r="12" fill="var(--ride-skin)" />
            <path d="M486,60 l4,3 l-4,1 z" fill="var(--ride-skin)" />
            <path d="M458,54 C460,40 476,36 488,44 C492,47 492,51 489,52 L462,58 C458,58 457,56 458,54 Z" fill="var(--ride-white)" />
            <path d="M464,45 C472,40 482,40 487,45" fill="none" stroke="var(--ride-red)" strokeWidth="2.5" />
            <path d="M478,55 l12,-1 q1,4 -3,5 l-9,0 z" fill="var(--ride-ink)" />
            <path d="M470,58 L482,68" stroke="var(--ride-ink)" strokeWidth="1.2" />
            <Limb name="near" />
            <Arm name="near" />
          </g>

          <g>
            <line data-crank x1={BB.x} y1={BB.y} stroke="var(--ride-metal)" strokeWidth="5" strokeLinecap="round" />
            <rect data-pedal x="-7" y="-2" width="14" height="4" rx="1.5" fill="var(--ride-ink)" />
            <circle cx={BB.x} cy={BB.y} r="5" fill="var(--ride-metal)" />
          </g>
        </g>

      </svg>
    </div>
  );
}
