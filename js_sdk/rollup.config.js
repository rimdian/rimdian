const { BUILD, ENVIRONMENT, BUILDALL } = process.env;

import resolve from '@rollup/plugin-node-resolve';
import babel from '@rollup/plugin-babel';
import typescript from '@rollup/plugin-typescript';
import commonjs from '@rollup/plugin-commonjs';
import json from '@rollup/plugin-json';
const extensions = ['.js', '.ts'];

const defaultOutputOptions = {
    name: 'Rimdian',
    strict: false,
    sourcemap: ENVIRONMENT !== 'prod' ? 'inline' : false,
};

const defaultBabel = babel({
    extensions,
    include: ['src/**/*'],
    babelHelpers: 'runtime',
});

const babelMinify = babel({
    extensions,
    include: ['src/**/*'],
    babelHelpers: 'runtime',
    babelrc: false,
    presets: [
        '@babel/preset-typescript',
        ['minify', { builtIns: false }],
        ['@babel/env', { modules: false }],
    ],
    plugins: [
        '@babel/proposal-class-properties',
        '@babel/proposal-object-rest-spread',
        '@babel/plugin-transform-runtime',
    ],
});

const builds = {
    prod: {
        input: 'src/sdk.ts',
        output: {
            ...defaultOutputOptions,
            file: 'dist/sdk.js',
            format: 'iife',
            name: 'Rimdian',
            globals: {
                'window': 'window',
                'document': 'document',
                'navigator': 'navigator',
            }
        },
        plugins: [
            babelMinify,
            resolve(),
            commonjs({
                include: 'node_modules/**',
            }),
            typescript({ tsconfig: './tsconfig.json' }),
            json(),
        ],
    },
    dev: {
        input: 'src/sdk.ts',
        output: {
            ...defaultOutputOptions,
            file: 'playground/sdk.js',
            format: 'iife',
            name: 'Rimdian',
            globals: {
                'window': 'window',
                'document': 'document',
                'navigator': 'navigator',
            }
        },
        plugins: [
            defaultBabel,
            resolve(),
            commonjs({
                include: 'node_modules/**',
            }),
            typescript({ tsconfig: './tsconfig.json' }),
            json(),
        ],
    }
};

let selectedBuilds = [];
if (BUILDALL) {
    for (let build of Object.keys(builds)) {
        selectedBuilds.push(builds[build]);
    }
} else {
    selectedBuilds.push(builds[BUILD]);
}

export default selectedBuilds;
