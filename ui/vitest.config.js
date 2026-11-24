import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import jsconfigPaths from 'vite-jsconfig-paths';
import path from 'path';

export default defineConfig({
  plugins: [react(), jsconfigPaths()],
  resolve: {
    alias: [
      {
        find: 'lodash.debounce',
        replacement: path.resolve('./node_modules/lodash/debounce.js'),
      },
      {
        find: /^~(.+)/,
        replacement: path.join(process.cwd(), 'node_modules/$1'),
      },
      {
        find: /^src(.+)/,
        replacement: path.join(process.cwd(), 'src/$1'),
      },
    ],
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test-setup.js'],
  },
});
