// tailwind.config.js（ルートに配置）
import { skeleton } from '@skeletonlabs/skeleton/tailwind';
import forms from '@tailwindcss/forms';
import typography from '@tailwindcss/typography';

export default {
  content: [
    './src/**/*.{html,js,svelte,ts}',
    './node_modules/@skeletonlabs/skeleton/**/*.{html,js,svelte,ts}'
  ],
  darkMode: 'class',
  plugins: [
    skeleton,  // これがないとCerberusボタン色が出ない！
    forms,
    typography
  ]
};
