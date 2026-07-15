import { readFileSync, existsSync } from 'node:fs';

const required = ['site/index.html', 'site/404.html', 'site/robots.txt', 'site/sitemap.xml', 'site/llms.txt', 'site/social-card.svg', 'site/favicon.svg', 'site/site.webmanifest'];
for (const file of required) {
  if (!existsSync(file)) throw new Error(`Missing ${file}`);
}
const html = readFileSync('site/index.html', 'utf8');
for (const marker of ['<title>', 'name="description"', 'rel="canonical"', 'application/ld+json', 'FAQPage', 'SoftwareApplication', 'WebSite', 'class="skip-link"', 'id="main-content"', 'prefers-reduced-motion', 'og:image:alt', 'twitter:image:alt']) {
  if (!html.includes(marker)) throw new Error(`Missing SEO/GEO marker: ${marker}`);
}
if (html.includes('Â') || html.includes('â€') || html.includes('â†')) throw new Error('Found likely character-encoding corruption in index.html');
const json = html.match(/<script type="application\/ld\+json">\s*([\s\S]*?)\s*<\/script>/)?.[1];
if (!json) throw new Error('Missing JSON-LD');
JSON.parse(json);
const notFound = readFileSync('site/404.html', 'utf8');
for (const marker of ['lang="en"', 'name="robots" content="noindex, follow"', '/ai-credit-scrub/']) {
  if (!notFound.includes(marker)) throw new Error(`Missing accessible 404 marker: ${marker}`);
}
console.log('site validation passed');
