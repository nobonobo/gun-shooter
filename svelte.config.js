import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  kit: {
    adapter: process.argv.includes('dev') ? undefined : adapter(),
    alias: {
      '@/*': './src/lib/*',
    }
  }
};

export default config;
