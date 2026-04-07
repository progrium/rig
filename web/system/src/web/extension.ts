
import * as vscode from 'vscode';
import * as duplex from "@progrium/duplex";
import { WanixBridge } from './bridge.js';
import { manifold, util } from '../webview/webview.js';

import { activate as activateTreeview } from './treeview.js';
import { activate as activateTerminal } from './terminal.js';
import { activate as activateInspector } from './inspector.js';
import { activate as activateCommands } from './commands.js';
import { activate as activateTempFS } from './tempfs.js';
import { activate as activateDocument } from './document.js';
import { activate as activateEditor } from './editor.js';

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
	
	const fsys = await bridge.ready;
	const token = (await fsys.readText("root/etc/token")).trim();
	const realm = new manifold.Realm();

	let peer: any;
	peer = await util.connectWithRetry("ws://localhost:8080/inspector/"+token, (conn: Conn) => {
		const sess = new duplex.Session(conn);
		if (!peer) {
			peer = new duplex.Peer(sess, new duplex.CBORCodec());
		} else {
			peer.session = sess;
			peer.caller = new duplex.Client(sess, new duplex.CBORCodec());
		}
		peer.respond();
		return peer;
	});

	// peer.handle("signaled", duplex.HandlerFunc(async (r: duplex.Responder, c: duplex.Call) => {
	// 	const [id] = await c.receive();
	// 	signals.port2.postMessage({type: "signaled", id: id});
	// 	treeView._onDidChangeTreeData.fire([id]);
	// 	inspector.reload();
	// }));
	peer.handle("update", duplex.HandlerFunc(async (r: duplex.Responder, c: duplex.Call) => {
		const update = await c.receive();
		// console.log("update:", Object.keys(update).length);
		realm.update(update);
		// signals.port2.postMessage({});
		// treeView._onDidChangeTreeData.fire([id]);
		// inspector.reload();
	}));
	peer.handle("execCommand", duplex.HandlerFunc(async (r: duplex.Responder, c: duplex.Call) => {
		const [cmd, args] = await c.receive();
		vscode.commands.executeCommand(cmd, ...args);
	}));
	
	activateTerminal(context, fsys);	
	activateTempFS(context, fsys, peer);
	activateCommands(context, fsys, peer);
	activateInspector(context, fsys, peer, realm);
	activateTreeview(context, fsys, peer, realm);
	activateEditor(context, token, peer, realm);
	// activateDocument(context, peer, realm);
	
	vscode.commands.executeCommand(`rig.createTerminal`);
	vscode.commands.executeCommand('manifold.hierarchy.focus');
	vscode.commands.executeCommand('manifold-sidebar.inspector.focus');

	console.log('System extension activated');
}

const connect = async (url: string, onconnect: (conn: Conn) => any, attempt = 0) => {
	console.log("CONNECT", url, attempt);
	return new Promise((resolve, reject) => {
		try {
			const ws = new WebSocket(url);
			ws.onclose = ({ wasClean }) => {
				if (wasClean) return;
				console.log("CLOSE");
				const backoff = Math.random() * Math.min(1000 * 2 ** attempt, 30000);
				setTimeout(() => {
					try {
						connect(url, onconnect, attempt + 1)
					} catch (err) {}
				}, backoff);
			};
			ws.onerror = reject;
			ws.onopen = () => resolve(onconnect(new Conn(ws)));
		} catch (err) {
			// ignore, ws.onclose will handle it
		}
	});
};


class Conn {
	ws: WebSocket
	waiters: Array<() => void>
	chunks: Array<Uint8Array>;
	isClosed: boolean
  
	constructor(ws: WebSocket) {
	  this.isClosed = false;
	  this.waiters = [];
	  this.chunks = [];
	  this.ws = ws;
	  this.ws.binaryType = "arraybuffer";
	  this.ws.onmessage = (event) => {
		const chunk = new Uint8Array(event.data);
		this.chunks.push(chunk);
		if (this.waiters.length > 0) {
		  const waiter = this.waiters.shift();
		  if (waiter) waiter();
		}
	  };
	  const onclose = this.ws.onclose;
	  this.ws.onclose = (e: CloseEvent) => {
		if (onclose) onclose.bind(this.ws)(e);
		this.close();
	  }
	}
  
	read(p: Uint8Array): Promise<number | null> {
	  return new Promise((resolve) => {
		var tryRead = () => {
		  if (this.isClosed) {
			resolve(null);
			return;
		  }
		  if (this.chunks.length === 0) {
			this.waiters.push(tryRead);
			return;
		  }
		  let written = 0;
		  while (written < p.length) {
			const chunk = this.chunks.shift();
			if (chunk === null || chunk === undefined) {
			  resolve(written);
			  return;
			}
			const buf = chunk.slice(0, p.length-written);
			p.set(buf, written)
			written += buf.length;
			if (chunk.length > buf.length) {
			  const restchunk = chunk.slice(buf.length);
			  this.chunks.unshift(restchunk);
			}
		  }
		  resolve(written);
		  return;
		}
		tryRead();
	  });
	}
  
	write(p: Uint8Array): Promise<number> {
	  this.ws.send(p);
	  return Promise.resolve(p.byteLength);
	}
  
	close() {
	  if (this.isClosed) return;
	  this.isClosed = true;
	  this.waiters.forEach(waiter => waiter());
	  this.ws.close();
	}
  }