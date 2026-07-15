import { readFileSync, existsSync } from 'node:fs';

const required = ['site/index.html', 'site/robots.txt', 'site/sitemap.xml', 'site/llms.txt', 'site/social-card.svg'];
for (const file of required) {
  if (!existsSync(file)) throw new Error(`Missing ${file}`);
}
const html = readFileSync('site/index.html', 'utf8');
for (const marker of ['<title>', 'name="description"', 'rel="canonical"', 'application/ld+json', 'FAQPage', 'SoftwareApplication']) {
  if (!html.includes(marker)) throw new Error(`Missing SEO/GEO marker: ${marker}`);
}
const json = html.match(/<script type="application\/ld\+json">\s*([\s\S]*?)\s*<\/script>/)?.[1];
if (!json) throw new Error('Missing JSON-LD');
JSON.parse(json);
console.log('site validation passed');
