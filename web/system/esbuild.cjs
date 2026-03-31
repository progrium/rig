const esbuild = require('esbuild');
const production = process.argv.includes('--production');
const watch = process.argv.includes('--watch');

const shared = {
	bundle: true,
	minify: production,
	sourcemap: !production,
	sourcesContent: false,
	platform: 'browser',
	external: ['vscode', 'util', 'worker_threads'],
	logLevel: 'silent',
	loader: {
		'.html': 'text',
	},
	jsxFactory: 'm',
	jsxFragment: 'm.Fragment',
	define: {
		global: 'globalThis',
	},
};

async function main() {
	const extensionOpts = {
		...shared,
		entryPoints: ['src/web/extension.ts'],
		format: 'cjs',
		outfile: 'dist/web/extension.js',
	};
	const webviewOpts = {
		...shared,
		entryPoints: ['src/webview/webview.ts'],
		format: 'esm',
		outfile: 'dist/webview/webview.js',
	};

	if (watch) {
		const extensionCtx = await esbuild.context(extensionOpts);
		const webviewCtx = await esbuild.context(webviewOpts);
		await Promise.all([extensionCtx.watch(), webviewCtx.watch()]);
	} else {
		await esbuild.build(extensionOpts);
		await esbuild.build(webviewOpts);
	}
}

main().catch((e) => {
	console.error(e);
	process.exit(1);
});
