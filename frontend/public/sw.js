const CACHE_NAME = 'cadencia-static-v3';
// Guarda, por instante, o arquivo recebido pelo menu "Compartilhar" do Android
// (Web Share Target), até a página /atividades/importar buscá-lo. Cache
// separado do CACHE_NAME para não ser limpo pela troca de versão dos
// recursos estáticos.
const SHARE_CACHE_NAME = 'cadencia-shared-file';
const SHARE_TARGET_PATH = '/atividades/compartilhar';
const SHARED_FILE_URL = '/__shared-activity';
// Dependendo do servidor (vinext start ou Cloudflare), a página offline responde
// em /offline ou somente em /offline.html; guardamos as que existirem.
const OFFLINE_URLS = ['/offline', '/offline.html'];
const PRECACHE = [
  '/app.webmanifest',
  '/favicon.svg',
  '/icons/icon-192.png',
  '/icons/icon-512.png',
  '/icons/icon-512-maskable.png',
  '/icons/apple-touch-icon.png',
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches
      .open(CACHE_NAME)
      .then(async (cache) => {
        // addAll falha por inteiro se uma URL falhar e impediria a instalação do
        // service worker; cada recurso é guardado de forma independente.
        await Promise.all([...OFFLINE_URLS, ...PRECACHE].map((url) => cache.add(url).catch(() => undefined)));
      })
      .then(() => self.skipWaiting()),
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys()
      .then((keys) => Promise.all(
        keys.filter((key) => key !== CACHE_NAME && key !== SHARE_CACHE_NAME).map((key) => caches.delete(key)),
      ))
      .then(() => self.clients.claim()),
  );
});

self.addEventListener('fetch', (event) => {
  const request = event.request;

  if (request.method === 'POST' && new URL(request.url).pathname === SHARE_TARGET_PATH) {
    event.respondWith(handleSharedActivity(request));
    return;
  }
  if (request.method !== 'GET') return;

  const url = new URL(request.url);
  if (url.origin !== self.location.origin || url.pathname.startsWith('/v1/')) return;

  if (request.mode === 'navigate') {
    event.respondWith(fetch(request).catch(async () => (await caches.match(OFFLINE_URLS[0])) || (await caches.match(OFFLINE_URLS[1]))));
    return;
  }

  if (['style', 'script'].includes(request.destination)) {
    event.respondWith(
      fetch(request)
        .then((response) => {
          if (response.ok) {
            const copy = response.clone();
            void caches.open(CACHE_NAME).then((cache) => cache.put(request, copy));
          }
          return response;
        })
        .catch(() => caches.match(request)),
    );
    return;
  }

  if (!['font', 'image'].includes(request.destination)) return;
  event.respondWith(
    caches.match(request).then((cached) => cached || fetch(request).then((response) => {
      if (response.ok) {
        const copy = response.clone();
        void caches.open(CACHE_NAME).then((cache) => cache.put(request, copy));
      }
      return response;
    })),
  );
});

// Recebe o POST do menu "Compartilhar" do Android (Web Share Target),
// definido em public/app.webmanifest. O navegador entrega o arquivo
// compartilhado aqui, antes de qualquer navegação; guardamos o arquivo num
// cache próprio e redirecionamos para a tela de importação, que busca o
// arquivo e completa o envio como se o atleta tivesse escolhido pelo seletor
// de arquivos. Nada aqui fala com a API: é só a ponte até a tela normal de
// upload, que passa pelas mesmas validações de sempre.
async function handleSharedActivity(request) {
  try {
    const form = await request.formData();
    const file = form.get('file');
    if (!(file instanceof File)) {
      return Response.redirect('/atividades/importar?compartilhado=erro', 303);
    }
    const cache = await caches.open(SHARE_CACHE_NAME);
    await cache.put(
      SHARED_FILE_URL,
      new Response(file, { headers: { 'X-Shared-Filename': encodeURIComponent(file.name) } }),
    );
  } catch {
    return Response.redirect('/atividades/importar?compartilhado=erro', 303);
  }
  return Response.redirect('/atividades/importar?compartilhado=1', 303);
}
