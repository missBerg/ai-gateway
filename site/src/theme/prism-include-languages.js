/**
 * Swizzled from @docusaurus/theme-classic (official customization point for
 * Prism languages).
 *
 * Beyond loading `themeConfig.prism.additionalLanguages`, this extends the
 * bash grammar: stock Prism-bash only tokenizes strings/comments/builtins, so
 * real-world CLI snippets (`helm upgrade -i eg oci://… --version v1.0`)
 * render as a wall of plain text. We add tokens for command names, flags,
 * and URLs so shell blocks read with the same hierarchy as other languages.
 */
import siteConfig from '@generated/docusaurus.config';

export default function prismIncludeLanguages(PrismObject) {
  const {
    themeConfig: {prism},
  } = siteConfig;
  const {additionalLanguages} = prism;

  // Prism components work on the global namespace instance.
  globalThis.Prism = PrismObject;

  additionalLanguages.forEach((lang) => {
    // eslint-disable-next-line global-require, import/no-dynamic-require
    require(`prismjs/components/prism-${lang}`);
  });

  if (PrismObject.languages.bash) {
    PrismObject.languages.insertBefore('bash', 'function', {
      // https:// and oci:// references — themed like strings.
      url: {
        pattern: /\b(?:https?|oci):\/\/\S+/,
        alias: 'string',
      },
      // -f / --version style flags — themed like attribute names.
      parameter: {
        pattern: /(^|\s)--?[\w][\w-]*/m,
        lookbehind: true,
        alias: 'attr-name',
      },
      // Leading command name on a line (helm, kubectl, …) — themed like a
      // function. Continuation lines are indented, so they don't match.
      'command-name': {
        pattern: /^[a-zA-Z][\w.-]*(?=\s|$)/m,
        alias: 'function',
      },
    });
  }

  delete globalThis.Prism;
}
