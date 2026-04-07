import * as vscode from 'vscode';

export async function activate(ctx: vscode.ExtensionContext, fsys: any) {
    [
        vscode.commands.registerCommand('rig.createTerminal', async () => {
            const term = vscode.window.createTerminal({ 
                name: 'Shell', 
                pty: await createTerminal(fsys)
            });
            term.show();
            ctx.subscriptions.push(term);
        })
    ].forEach(sub => ctx.subscriptions.push(sub));
}


async function createTerminal(fsys: any) {
	const rid = (await fsys.readText("#term/new")).trim();
	const writeEmitter = new vscode.EventEmitter<string>();
	const dec = new TextDecoder();
	const enc = new TextEncoder();
	const readable = await fsys.openReadable(`#term/${rid}/data`);
	const writable = (await fsys.openWritable(`#term/${rid}/data`)).getWriter();
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
			const winch = (await fsys.openWritable(`#term/${rid}/winch`)).getWriter();
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
