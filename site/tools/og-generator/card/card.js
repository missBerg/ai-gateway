/* Reads the card config from query params, populates the DOM, inlines the
   lucide icon, and auto-shrinks the title until it fits. Sets
   window.__ogReady = true when everything (fonts, images, layout) is settled —
   the screenshot step waits on that flag. */

const params = new URLSearchParams(window.location.search);

const config = {
  title: params.get('title') || 'Untitled post',
  eyebrow: params.get('eyebrow') || '',
  subtitle: params.get('subtitle') || '',
  icon: params.get('icon') || '',
  visual: params.get('visual') || 'icon', // icon | fade | frame | none
  image: params.get('image') || '',
  accent: params.get('accent') || 'brand', // brand | signature
  footer: params.get('footer') || 'aigateway.envoyproxy.io',
};

const card = document.getElementById('card');
const titleEl = document.getElementById('title');
const eyebrowEl = document.getElementById('eyebrow');
const subtitleEl = document.getElementById('subtitle');
const footerEl = document.getElementById('footer');

async function fetchIcon(name) {
  const res = await fetch(`/lucide/${encodeURIComponent(name)}.svg`);
  if (!res.ok) throw new Error(`unknown lucide icon: ${name}`);
  return res.text();
}

function loadImage(imgEl, src) {
  return new Promise((resolve, reject) => {
    imgEl.onload = () => resolve();
    imgEl.onerror = () => reject(new Error(`failed to load image: ${src}`));
    imgEl.src = src;
  });
}

/* Shrink the title font until the content block fits inside the card. */
function fitTitle() {
  const content = document.getElementById('content');
  let size = 72;
  const min = 40;
  titleEl.style.fontSize = `${size}px`;
  while (size > min && content.scrollHeight > content.clientHeight) {
    size -= 2;
    titleEl.style.fontSize = `${size}px`;
  }
}

async function init() {
  card.dataset.visual = config.visual;
  card.dataset.accent = config.accent;

  titleEl.textContent = config.title;
  footerEl.textContent = config.footer;

  if (config.subtitle) {
    subtitleEl.textContent = config.subtitle;
    subtitleEl.hidden = false;
  }

  const work = [];

  if (config.eyebrow) {
    if (config.icon) {
      const iconSpan = document.createElement('span');
      iconSpan.className = 'eyebrowIcon';
      eyebrowEl.appendChild(iconSpan);
      work.push(fetchIcon(config.icon).then((svg) => (iconSpan.innerHTML = svg)));
    }
    eyebrowEl.appendChild(document.createTextNode(config.eyebrow));
  } else {
    eyebrowEl.hidden = true;
  }

  if (config.visual === 'icon' && config.icon) {
    const bgIcon = document.getElementById('bgIcon');
    bgIcon.hidden = false;
    work.push(fetchIcon(config.icon).then((svg) => (bgIcon.innerHTML = svg)));
  } else if (config.visual === 'fade' && config.image) {
    document.getElementById('fadeImage').hidden = false;
    work.push(loadImage(document.getElementById('fadeImg'), config.image));
  } else if (config.visual === 'frame' && config.image) {
    document.getElementById('frameImage').hidden = false;
    work.push(loadImage(document.getElementById('frameImg'), config.image));
  }

  await Promise.all(work);
  await document.fonts.ready;
  fitTitle();
  window.__ogReady = true;
}

init().catch((err) => {
  window.__ogError = String(err);
  console.error(err);
});
