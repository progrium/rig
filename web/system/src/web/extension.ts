
import * as vscode from 'vscode';
import { WanixBridge } from './bridge.js';
import { activate as activateTreeview } from './treeview.js';

declare const navigator: unknown;

export async function activate(context: vscode.ExtensionContext) {
	if (typeof navigator !== 'object') {	// do not run under node.js
		console.error("not running in browser");
		return;
	}
	
	const channel = new MessageChannel();
	const bridge = new WanixBridge(channel.port2, "");
	context.subscriptions.push(bridge);

	// get our message port from config
	const port = (context as any).messagePassingProtocol;
	// send the bridge channel port to env to have a wanix port sent over it
	// wanix port -> bridge port -> message passing port
	port.postMessage({type: "_port", port: channel.port1}, [channel.port1]);

	bridge.ready.then(async (wfsys) => {

		vscode.commands.registerCommand('rig.createTerminal', async () => {
			const term = vscode.window.createTerminal({ name: 'Shell', pty: await createTerminal(wfsys)});
			context.subscriptions.push(term);
			term.show();
		});

		vscode.commands.executeCommand(`rig.createTerminal`);
		
		activateTreeview(context, wfsys);
		
		vscode.commands.executeCommand('manifold.hierarchy.focus');
		vscode.commands.executeCommand('manifold-sidebar.inspector.focus');

	});

	
	console.log('System extension activated');
}

async function createTerminal(wx: any) {
	const rid = (await wx.readText("#term/new")).trim();
	const writeEmitter = new vscode.EventEmitter<string>();
	const dec = new TextDecoder();
	const enc = new TextEncoder();
	const readable = await wx.openReadable(`#term/${rid}/data`);
	const writable = (await wx.openWritable(`#term/${rid}/data`)).getWriter();
	return {
		onDidWrite: writeEmitter.event,
		open: () => {
			(async () => {
				for await (const chunk of readable) {
					writeEmitter.fire(dec.decode(chunk));
				}
			})();
		},
		close: () => {
			writable.close();
		},
		handleInput: async (data: string) => {
			await writable.write(enc.encode(data));
		},
		setDimensions: async (dimensions: vscode.TerminalDimensions) => {
			const winch = (await wx.openWritable(`#term/${rid}/winch`)).getWriter();
			await winch.write(enc.encode(`${dimensions.columns} ${dimensions.rows}\n`));
			await winch.close();
		}
	};
}


// @ts-ignore
// polyfill for ReadableStream.prototype[Symbol.asyncIterator] on safari
if (!ReadableStream.prototype[Symbol.asyncIterator]) {
	// @ts-ignore
    ReadableStream.prototype[Symbol.asyncIterator] = async function* () {
        const reader = this.getReader();
        try {
            while (true) {
                const { done, value } = await reader.read();
                if (done) return;
                yield value;
            }
        } finally {
            reader.releaseLock();
        }
    };
}
