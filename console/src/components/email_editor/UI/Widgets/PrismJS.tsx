import React from 'react';
import Prismjs from 'prismjs';

import 'prismjs/themes/prism.css'; /* or your own custom theme */
import 'prismjs/plugins/line-numbers/prism-line-numbers.css' /* add plugin css */

// Require all needed languages
require('prismjs/components/prism-xml-doc')
// require('prismjs/components/prism-typescript');
// require('prismjs/components/prism-jsx');
// require('prismjs/components/prism-tsx');

// Require all needed plugins
require('prismjs/plugins/line-numbers/prism-line-numbers');

export function usePrismjs<T extends HTMLElement>(
  target: React.RefObject<T>,
  plugins: string[] = []
) {
  React.useLayoutEffect(() => {
    if (target.current) {
      if (plugins.length > 0) {
        target.current.classList.add(...plugins);
      }
      // Highlight all <pre><code>...</code></pre> blocks contained by this element
      Prismjs.highlightAllUnder(target.current);
    }
  }, [target, plugins]);
}
