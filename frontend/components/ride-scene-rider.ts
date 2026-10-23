// Ciclista e bicicleta da cena de acesso (components/ride-scene.tsx), modelados
// em código para o Cadência: bicicleta de estrada em carbono com a pintura da
// marca e um ciclista de proporções reais (membros torneados por perfil de
// revolução), com materiais físicos — verniz, tecido com brilho, borracha e
// metal. Recebe o módulo do Three.js já carregado, para que o pacote só seja
// baixado na tela de acesso.
//
// Eixos locais: +x é a frente da bicicleta, +y para cima, z para os lados.

type Three = typeof import('three');
type Vec3 = import('three').Vector3;
type Obj3D = import('three').Object3D;
type Geometry = import('three').BufferGeometry;
type Material = import('three').Material;

export const WHEEL_R = 0.34;
export const CRANK_R = 0.17;

const THIGH = 0.45;
const SHIN = 0.44;
const UPPER_ARM = 0.29;
const FOREARM = 0.28;

export type Rider = {
  root: Obj3D;
  pose: (crankAngle: number, wheelAngle: number) => void;
  dispose: () => void;
};

export function buildRider(THREE: Three): Rider {
  const disposables: Array<{ dispose(): void }> = [];
  const keep = <T extends { dispose(): void }>(item: T) => {
    disposables.push(item);
    return item;
  };

  // ---------- Materiais ----------
  const physical = (params: ConstructorParameters<typeof THREE.MeshPhysicalMaterial>[0]) =>
    keep(new THREE.MeshPhysicalMaterial(params));
  const carbon = physical({ color: '#16191c', roughness: 0.34, metalness: 0.2, clearcoat: 1, clearcoatRoughness: 0.06 });
  const paint = physical({ color: '#c3321f', roughness: 0.3, metalness: 0.35, clearcoat: 1, clearcoatRoughness: 0.04 });
  const alloy = physical({ color: '#b8bcc0', roughness: 0.26, metalness: 1 });
  const darkAlloy = physical({ color: '#3a3e42', roughness: 0.35, metalness: 1 });
  const rubber = physical({ color: '#1c1c1c', roughness: 0.9 });
  const tape = physical({ color: '#202326', roughness: 0.75 });
  const jersey = physical({ color: '#c3321f', roughness: 0.62, sheen: 1, sheenColor: '#ff9a80', sheenRoughness: 0.45 });
  const jerseyLight = physical({ color: '#f4f5f2', roughness: 0.6, sheen: 1, sheenColor: '#ffffff', sheenRoughness: 0.4 });
  const bib = physical({ color: '#15171a', roughness: 0.42, sheen: 0.9, sheenColor: '#6d7884', sheenRoughness: 0.35 });
  const skin = physical({ color: '#b98762', roughness: 0.55, sheen: 0.35, sheenColor: '#ffd4b5', sheenRoughness: 0.6 });
  const sock = physical({ color: '#f1f2ef', roughness: 0.85 });
  const shell = physical({ color: '#f4f5f2', roughness: 0.25, clearcoat: 1, clearcoatRoughness: 0.08 });
  const lens = physical({ color: '#1a1d22', roughness: 0.04, metalness: 0.9, iridescence: 1, iridescenceIOR: 1.7 });
  const bottleBlue = physical({ color: '#2f6f9e', roughness: 0.4, clearcoat: 0.6 });

  const v = (x: number, y: number, z = 0) => new THREE.Vector3(x, y, z);
  const up = v(0, 1, 0);

  function mesh(geometry: Geometry, material: Material) {
    const m = new THREE.Mesh(keep(geometry), material);
    m.castShadow = true;
    m.receiveShadow = true;
    return m;
  }
  function add<T extends Obj3D>(parent: Obj3D, part: T) {
    parent.add(part);
    return part;
  }
  function placeAt<T extends Obj3D>(parent: Obj3D, part: T, point: Vec3) {
    part.position.copy(point);
    parent.add(part);
    return part;
  }
  // Peça que nasce em y = 0 e cresce em +y: orienta de a para b.
  function aim(part: Obj3D, a: Vec3, b: Vec3) {
    part.position.copy(a);
    part.quaternion.setFromUnitVectors(up, new THREE.Vector3().subVectors(b, a).normalize());
    return part;
  }
  // Tubo cônico entre dois pontos, achatado para o perfil aerodinâmico.
  function rod(parent: Obj3D, a: Vec3, b: Vec3, r1: number, r2: number, material: Material, aero = 1) {
    const length = a.distanceTo(b);
    const geometry = new THREE.CylinderGeometry(r2, r1, length, 20);
    geometry.translate(0, length / 2, 0);
    const part = mesh(geometry, material);
    part.scale.set(aero, 1, 1);
    const holder = new THREE.Group();
    holder.add(part);
    parent.add(aim(holder, a, b));
    return holder;
  }
  function ellipsoid(parent: Obj3D, at: Vec3, sx: number, sy: number, sz: number, material: Material) {
    const part = placeAt(parent, mesh(new THREE.SphereGeometry(1, 28, 20), material), at);
    part.scale.set(sx, sy, sz);
    return part;
  }
  // Membro torneado: perfil [raio, fração do comprimento], girado em torno de y.
  function lathe(profile: Array<[number, number]>, length: number, material: Material, squash = 1) {
    const points = profile.map(([r, t]) => new THREE.Vector2(r, t * length));
    const part = mesh(new THREE.LatheGeometry(points, 28), material);
    part.scale.set(squash, 1, 1);
    const holder = new THREE.Group();
    holder.add(part);
    return holder;
  }

  const root = new THREE.Group();
  const bike = add(root, new THREE.Group());

  const rear = v(-0.5, WHEEL_R);
  const front = v(0.53, WHEEL_R);
  const bb = v(0, 0.29);
  const seatCluster = v(-0.14, 0.78);
  const headTop = v(0.39, 0.84);
  const headLow = v(0.43, 0.69);
  const bar = v(0.46, 0.92);

  // ---------- Rodas: carbono de perfil alto, raios e pneu ----------
  const SPOKES = 20;
  const spokeLength = WHEEL_R - 0.085;
  const spokeGeometry = keep(new THREE.CylinderGeometry(0.0035, 0.0035, spokeLength, 5));
  spokeGeometry.translate(0, spokeLength / 2, 0);
  const wheels = [rear, front].map((center, wheelIndex) => {
    const wheel = placeAt(bike, new THREE.Group(), center);
    wheel.add(mesh(new THREE.TorusGeometry(WHEEL_R - 0.012, 0.024, 16, 96), rubber));
    const rim = mesh(new THREE.TorusGeometry(WHEEL_R - 0.06, 0.042, 12, 96), carbon);
    rim.scale.z = 0.42;
    wheel.add(rim);
    // Faixa vermelha no aro: a marca do Cadência nas rodas.
    const stripe = mesh(new THREE.TorusGeometry(WHEEL_R - 0.058, 0.0435, 8, 96, Math.PI * 0.5), paint);
    stripe.scale.z = 0.43;
    stripe.rotation.z = wheelIndex * 1.3;
    wheel.add(stripe);
    const spokes = new THREE.InstancedMesh(spokeGeometry, darkAlloy, SPOKES);
    const matrix = new THREE.Matrix4();
    const q = new THREE.Quaternion();
    for (let i = 0; i < SPOKES; i += 1) {
      const angle = (i / SPOKES) * Math.PI * 2;
      q.setFromAxisAngle(v(0, 0, 1), angle);
      matrix.compose(v(0, 0, i % 2 ? 0.022 : -0.022), q, v(1, 1, 1));
      spokes.setMatrixAt(i, matrix);
    }
    spokes.castShadow = true;
    wheel.add(spokes);
    const hub = mesh(new THREE.CylinderGeometry(0.03, 0.03, 0.1, 20), alloy);
    hub.rotation.x = Math.PI / 2;
    wheel.add(hub);
    // Disco de freio.
    const rotor = mesh(new THREE.CylinderGeometry(0.075, 0.075, 0.004, 32), alloy);
    rotor.rotation.x = Math.PI / 2;
    rotor.position.z = -0.045;
    wheel.add(rotor);
    return wheel;
  });

  // Cassete atrás (fica parado em relação ao quadro, gira com a roda visualmente).
  for (let i = 0; i < 6; i += 1) {
    const cog = mesh(new THREE.CylinderGeometry(0.06 - i * 0.006, 0.06 - i * 0.006, 0.004, 24), alloy);
    cog.rotation.x = Math.PI / 2;
    cog.position.set(rear.x, rear.y, 0.05 + i * 0.006);
    bike.add(cog);
  }

  // ---------- Quadro: carbono com a parte de cima pintada ----------
  rod(bike, bb, seatCluster, 0.026, 0.021, carbon, 0.8);
  rod(bike, seatCluster, headTop, 0.022, 0.024, paint, 0.75);
  rod(bike, bb, headLow, 0.032, 0.028, carbon, 0.7);
  rod(bike, headLow, headTop, 0.03, 0.028, paint);
  for (const side of [-0.055, 0.055]) {
    rod(bike, v(rear.x, rear.y, side), v(seatCluster.x - 0.01, seatCluster.y - 0.05, side * 0.3), 0.011, 0.014, carbon);
    rod(bike, v(rear.x, rear.y, side), v(bb.x - 0.03, bb.y, side * 0.5), 0.012, 0.016, carbon);
    rod(bike, v(headLow.x, headLow.y, side * 0.45), v(front.x, front.y, side), 0.017, 0.011, carbon, 0.7);
  }
  const bbShell = mesh(new THREE.CylinderGeometry(0.036, 0.036, 0.09, 24), carbon);
  bbShell.rotation.x = Math.PI / 2;
  placeAt(bike, bbShell, bb);

  // Canote e selim esculpido.
  const seatTop = v(-0.16, 0.9);
  rod(bike, seatCluster, seatTop, 0.014, 0.014, carbon);
  const saddleShape = new THREE.Shape();
  saddleShape.moveTo(-0.13, 0);
  saddleShape.bezierCurveTo(-0.13, 0.075, -0.05, 0.07, 0.0, 0.04);
  saddleShape.bezierCurveTo(0.05, 0.025, 0.12, 0.022, 0.135, 0);
  saddleShape.bezierCurveTo(0.12, -0.022, 0.05, -0.025, 0.0, -0.04);
  saddleShape.bezierCurveTo(-0.05, -0.07, -0.13, -0.075, -0.13, 0);
  const saddleGeometry = new THREE.ExtrudeGeometry(saddleShape, {
    depth: 0.018,
    bevelEnabled: true,
    bevelThickness: 0.012,
    bevelSize: 0.012,
    bevelSegments: 4,
    curveSegments: 24,
  });
  saddleGeometry.rotateX(Math.PI / 2);
  const saddle = placeAt(bike, mesh(saddleGeometry, tape), v(-0.165, 0.935));
  saddle.rotation.z = 0.03;

  // ---------- Cockpit: mesa, guidão drop com fita e manetes ----------
  rod(bike, headTop, bar, 0.017, 0.016, carbon);
  rod(bike, v(bar.x, bar.y, -0.2), v(bar.x, bar.y, 0.2), 0.014, 0.014, carbon);
  for (const side of [-0.2, 0.2]) {
    const curve = new THREE.CatmullRomCurve3([
      v(bar.x, bar.y, side),
      v(bar.x + 0.1, bar.y - 0.005, side),
      v(bar.x + 0.135, bar.y - 0.08, side),
      v(bar.x + 0.08, bar.y - 0.15, side),
      v(bar.x - 0.02, bar.y - 0.155, side),
    ]);
    bike.add(mesh(new THREE.TubeGeometry(curve, 40, 0.0135, 12), tape));
    const hood = ellipsoid(bike, v(bar.x + 0.115, bar.y + 0.022, side), 0.045, 0.03, 0.02, rubber);
    hood.rotation.z = -0.35;
    rod(bike, v(bar.x + 0.13, bar.y + 0.01, side), v(bar.x + 0.15, bar.y - 0.11, side), 0.008, 0.006, darkAlloy);
  }

  // Caramanhola no tubo inferior.
  const bottle = placeAt(bike, lathe([[0, 0], [0.034, 0.02], [0.034, 0.8], [0.02, 0.92], [0.012, 1], [0, 1]], 0.2, bottleBlue), v(0.08, 0.42, 0));
  bottle.rotation.z = -0.95;

  // ---------- Transmissão ----------
  const ring = mesh(new THREE.TorusGeometry(0.1, 0.008, 8, 64), darkAlloy);
  ring.position.set(bb.x, bb.y, 0.075);
  bike.add(ring);
  const ring2 = mesh(new THREE.TorusGeometry(0.08, 0.006, 8, 64), darkAlloy);
  ring2.position.set(bb.x, bb.y, 0.067);
  bike.add(ring2);
  const chain = new THREE.CatmullRomCurve3(
    [
      v(bb.x + 0.02, bb.y + 0.1, 0.075),
      v(rear.x, rear.y + 0.055, 0.065),
      v(rear.x - 0.05, rear.y, 0.065),
      v(rear.x, rear.y - 0.055, 0.065),
      v(rear.x + 0.2, rear.y - 0.09, 0.07),
      v(bb.x, bb.y - 0.1, 0.075),
      v(bb.x + 0.1, bb.y, 0.075),
    ],
    true,
  );
  bike.add(mesh(new THREE.TubeGeometry(chain, 120, 0.005, 6, true), darkAlloy));
  const crank = placeAt(bike, new THREE.Group(), bb);
  [0.09, -0.09].forEach((z, index) => {
    const arm = new THREE.Group();
    arm.position.z = z;
    arm.rotation.z = index * Math.PI;
    rod(arm, v(0, 0), v(CRANK_R, 0), 0.016, 0.012, carbon, 0.6);
    crank.add(arm);
  });

  // ---------- Ciclista ----------
  const rider = add(bike, new THREE.Group());
  const hip = v(-0.17, 0.98);
  const shoulder = v(0.19, 1.27);

  // Tronco: perfil de revolução achatado (mais largo nos ombros).
  const torso = add(
    rider,
    lathe(
      [
        [0, 0],
        [0.1, 0.02],
        [0.118, 0.14],
        [0.104, 0.36],
        [0.122, 0.62],
        [0.132, 0.82],
        [0.1, 0.97],
        [0, 1],
      ],
      hip.distanceTo(shoulder) + 0.05,
      jersey,
      0.78,
    ),
  );
  aim(torso, hip.clone().add(v(-0.02, -0.03)), shoulder);
  torso.children[0].scale.z = 1.32;
  // Faixa branca no peito e a bermuda (bib) até a cintura.
  const band = add(rider, lathe([[0.125, 0], [0.127, 0.5], [0.125, 1]], 0.06, jerseyLight, 0.79));
  aim(band, hip.clone().lerp(shoulder, 0.62), shoulder);
  band.children[0].scale.z = 1.34;
  const shorts = ellipsoid(rider, hip.clone().add(v(0.01, -0.01)), 0.12, 0.1, 0.15, bib);
  shorts.rotation.z = -0.5;

  // Cabeça, pescoço, capacete aero e óculos.
  const head = v(0.34, 1.385);
  rod(rider, v(shoulder.x + 0.03, shoulder.y), v(head.x - 0.04, head.y - 0.06), 0.042, 0.036, skin);
  ellipsoid(rider, head, 0.075, 0.088, 0.07, skin);
  const helmet = ellipsoid(rider, v(head.x - 0.035, head.y + 0.045), 0.135, 0.075, 0.098, shell);
  helmet.rotation.z = 0.22;
  const vent = ellipsoid(rider, v(head.x - 0.02, head.y + 0.105), 0.09, 0.012, 0.03, paint);
  vent.rotation.z = 0.22;
  const glasses = ellipsoid(rider, v(head.x + 0.066, head.y + 0.01), 0.018, 0.026, 0.085, lens);
  glasses.rotation.y = 0.02;

  // Pernas: coxa na bermuda, panturrilha torneada, meia e sapatilha.
  const legs = [0.1, -0.1].map((z) => {
    const thigh = add(rider, lathe([[0, 0], [0.072, 0.04], [0.078, 0.3], [0.064, 0.75], [0.05, 0.97], [0, 1]], THIGH + 0.03, bib));
    const shin = add(
      rider,
      lathe([[0, 0], [0.046, 0.03], [0.056, 0.28], [0.046, 0.6], [0.031, 0.86], [0.03, 1], [0, 1]], SHIN, skin),
    );
    const cuff = add(rider, lathe([[0.033, 0], [0.034, 1]], 0.1, sock));
    const kneeCap = ellipsoid(rider, v(0, 0), 0.05, 0.05, 0.05, skin);
    const shoe = add(rider, new THREE.Group());
    const sole = ellipsoid(shoe, v(0.03, -0.006), 0.1, 0.032, 0.044, shell);
    sole.rotation.z = 0.05;
    ellipsoid(shoe, v(0.0, 0.012), 0.07, 0.03, 0.04, carbon);
    return { z, thigh, shin, cuff, kneeCap, shoe };
  });
  const pedals = legs.map(() => {
    const pedal = mesh(new THREE.BoxGeometry(0.1, 0.012, 0.06), carbon);
    bike.add(pedal);
    return pedal;
  });

  // Braços: manga vermelha, antebraço, luva.
  const arms = [0.16, -0.16].map((z) => ({
    z,
    upper: add(rider, lathe([[0, 0], [0.05, 0.05], [0.048, 0.45], [0.04, 1], [0, 1]], UPPER_ARM + 0.02, jersey)),
    cuff: add(rider, lathe([[0.041, 0], [0.042, 1]], 0.03, jerseyLight)),
    fore: add(rider, lathe([[0, 0], [0.038, 0.06], [0.04, 0.3], [0.028, 0.92], [0, 1]], FOREARM, skin)),
    glove: ellipsoid(rider, v(0, 0), 0.042, 0.032, 0.035, carbon),
  }));
  // Deltoides: unem a manga ao tronco sem formar uma bola.
  for (const side of [0.14, -0.14]) {
    const deltoid = ellipsoid(rider, shoulder.clone().add(v(-0.01, -0.03, side)), 0.065, 0.05, 0.05, jersey);
    deltoid.rotation.z = -0.6;
  }

  // Dois segmentos entre a e b, com a articulação para o lado indicado
  // (joelho para a frente, cotovelo para baixo), no plano xy.
  const joint = new THREE.Vector3();
  function solve(a: Vec3, b: Vec3, first: number, second: number, bend: 1 | -1) {
    const dx = b.x - a.x;
    const dy = b.y - a.y;
    const dist = Math.min(Math.hypot(dx, dy), first + second - 0.001);
    const base = Math.atan2(dy, dx);
    const cos = (first * first + dist * dist - second * second) / (2 * first * dist);
    const angle = base + bend * Math.acos(Math.max(-1, Math.min(1, cos)));
    return joint.set(a.x + Math.cos(angle) * first, a.y + Math.sin(angle) * first, (a.z + b.z) / 2);
  }

  function pose(crankAngle: number, wheelAngle: number) {
    wheels.forEach((wheel) => {
      wheel.rotation.z = -wheelAngle;
    });
    crank.rotation.z = -crankAngle;
    // O corpo acompanha a pedalada: quadril balança e o tronco respira.
    const rock = Math.sin(crankAngle * 2);
    rider.rotation.x = rock * 0.022;
    legs.forEach((leg, index) => {
      const angle = -crankAngle + index * Math.PI;
      const foot = v(bb.x + Math.cos(angle) * CRANK_R, bb.y + Math.sin(angle) * CRANK_R, leg.z);
      pedals[index].position.copy(foot);
      const hipPoint = v(hip.x, hip.y + rock * 0.006, leg.z * 0.85);
      const ankle = foot.clone().add(v(-0.03, 0.05));
      const knee = solve(hipPoint, ankle, THIGH, SHIN, 1).clone();
      aim(leg.thigh, hipPoint.clone().add(v(0, 0.02)), knee);
      aim(leg.shin, knee, ankle);
      aim(leg.cuff, ankle, knee);
      leg.kneeCap.position.copy(knee);
      leg.shoe.position.copy(foot).add(v(0.0, 0.03));
      leg.shoe.rotation.z = Math.sin(angle) * 0.18 - 0.05;
    });
    arms.forEach((arm) => {
      const from = v(shoulder.x, shoulder.y - 0.02 + rock * 0.004, arm.z);
      const hand = v(bar.x + 0.11, bar.y + 0.035, arm.z * 1.25);
      const elbow = solve(from, hand, UPPER_ARM, FOREARM, -1).clone();
      elbow.z = arm.z * 1.18;
      aim(arm.upper, from, elbow);
      aim(arm.cuff, from.clone().lerp(elbow, 0.92), elbow);
      aim(arm.fore, elbow, hand);
      arm.glove.position.copy(hand);
    });
  }

  return {
    root,
    pose,
    dispose: () => disposables.forEach((item) => item.dispose()),
  };
}
