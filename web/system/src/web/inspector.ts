
import * as vscode from 'vscode';
import { manifold } from '../webview/webview.js';

export async function activate(ctx: vscode.ExtensionContext, fsys: any, peer: any, realm: manifold.Realm) {
    const token = (await fsys.readText("root/etc/token")).trim();
    const inspector = new InspectorViewProvider(ctx.extensionUri, "ws://localhost:8080/inspector/"+token, peer);
	
	ctx.subscriptions.push(
		vscode.window.registerWebviewViewProvider(InspectorViewProvider.viewType, inspector),
		vscode.commands.registerCommand('manifold.inspect', (id: string) => {
            inspector.selectNode(id);
        })
	);

	realm.addEventListener("change", (event) => {
		inspector.reload();
	});
    
}

export class InspectorViewProvider implements vscode.WebviewViewProvider {

	public static readonly viewType = 'manifold-sidebar.inspector';

	private _view?: vscode.WebviewView;
	private _selectedID?: string;

	constructor(
		private readonly _extensionUri: vscode.Uri,
		private readonly websocketURL: string,
		public readonly peer: any
	) { }

	public selectNode(id: string) {
		this._selectedID = id;
		this.reload();
	}

	public async reload() {
		if (!this._view || !this._selectedID) {
			return;
		}
		const fields: any[] = [];
		const resp = await this.peer.call("fields", this._selectedID);
		while (true) {
			const field = await resp.receive();
			if (field === null) {
				break;
			}
			fields.push(field);
		}
		this._view.webview.postMessage({
			manifold: "selectNode",
			id: this._selectedID,
			fields: JSON.parse(JSON.stringify(fields, (key, value) =>
				typeof value === "bigint" ? Number(value) : value,
			))
		});
	}

	public resolveWebviewView(
		webviewView: vscode.WebviewView,
		context: vscode.WebviewViewResolveContext,
		_token: vscode.CancellationToken,
	) {
		this._view = webviewView;

		webviewView.webview.options = {
			enableScripts: true,
			localResourceRoots: [
				this._extensionUri
			]
		};

		const nonce = getNonce();
		webviewView.webview.html = `
<html>
<head>
	<meta http-equiv="Content-Security-Policy" content="
		default-src 'none';
		script-src 'nonce-${nonce}' http://localhost:8080;
		connect-src http://localhost:8080 ws://localhost:8080;
		style-src http://localhost:8080 'unsafe-inline';
		font-src http://localhost:8080;
	">
	<link rel="stylesheet" href="${this._extensionUri.with({path: "system/media/fontawesome/css/all.min.css"}).toString()}">
	<link rel="stylesheet" href="${this._extensionUri.with({path: "system/media/inspector.css"}).toString()}">
</head>
<body>
<script nonce="${nonce}" type="module">
import {
	m,
	manifold,
	duplex,
	util,
	inspector
} from "${this._extensionUri.with({path: "system/dist/webview/webview.js"}).toString()}";

  
  var peer = undefined;
  peer = await util.connectWithRetry("${this.websocketURL}", (conn) => {
    const sess = new duplex.Session(conn);
    if (!peer) {
      peer = new duplex.Peer(sess, new duplex.CBORCodec());
    } else {
      peer.session = sess;
      peer.caller = new duplex.Client(sess, new duplex.CBORCodec());
      peer.respond();
    }
	return peer;
  });

  window.m = m;
  window.nodes = new manifold.Realm(peer);
  window.vscode = acquireVsCodeApi();
  
  peer.handle("update", duplex.HandlerFunc(async (r, c) => {
    const update = await c.receive();
    // console.log("update:", Object.keys(update).length);
    window.nodes.update(update);
    m.redraw();
  }));
  peer.respond();


  class Backend {
    setValue(field, value) {
      peer.call("setValue", {Type: field.TypeName, Selector: field.ID, Value: value });
    }

    appendValue(field, value) {
      peer.call("appendValue", {Type: field.TypeName, Selector: field.ID, Value: value});
    }
    
    removeKey(field, key) {
      peer.call("unsetValue", {Type: field.TypeName, Selector: field.ID+"/"+key});
    }
    
    setKey(field, key, value) {
      peer.call("setValue", {Type: field.TypeName, Selector: field.ID+"/"+key, Value: value });
    }
  }
  window.store = new Backend();

  var selectedID = "@main";
  var fields = undefined;
  var mounted = false;
  window.addEventListener("message", (e) => {
      if (!e.data.manifold) {
        return;
      }
      selectedID = e.data.id;
      fields = e.data.fields;
      if (!mounted) {
        m.mount(document.body, App);
        mounted = true;
      }
      m.redraw();
  });

  const App = {
    view: () => 
      	m("main", {style: {position: "absolute", left: "0", top: "0", left: "0", right: "0"}}, [
        	m(inspector.Inspector, {selectedID, fields, store})
		])
  }
</script>
</body>
</html>`;

	// <button style="display: none;" id="menu-more" data-vscode-context='{"webviewSection": "more",  "preventDefaultContextMenuItems": true}'></button>
	// <button style="display: none;" id="menu-ref" data-vscode-context='{"webviewSection": "ref",  "preventDefaultContextMenuItems": true}'></button>
	// <script type="text/javascript">
	// const vscode = acquireVsCodeApi();
	// window.addEventListener("message", e => {
	// 	if (e.data.manifold) {
	// 		frame.contentWindow.postMessage(e.data, "${this.iframeURL}");
	// 		return;
	// 	}
	// 	if (e.data.menu) {
	// 		const button = document.getElementById(e.data.menu);
	// 		const ctx = JSON.parse(button.dataset.vscodeContext);
	// 		Object.assign(ctx, e.data.params);
	// 		button.dataset.vscodeContext = JSON.stringify(ctx);
	// 		button.dispatchEvent(new MouseEvent('contextmenu', {
    // 			bubbles: true, 
    // 			cancelable: true, 
    // 			view: window,
	// 			clientX: e.data.params.clientX, 
	// 			clientY: e.data.params.clientY,
	// 		}));
	// 		return;
	// 	}
	// 	if (e.data.action) {
	// 		vscode.postMessage(e.data);
	// 	}
	// });
	// vscode.postMessage({ready: true});
	// </script>

		webviewView.webview.onDidReceiveMessage(data => {
			if (data.action) {
				vscode.commands.executeCommand(data.action, ...(data.args||[]))
				return;
			}
			if (data.ready) {
				this.selectNode("@main");
				return;
			}
		});
	}
}

function getNonce() {
	let text = '';
	const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
	for (let i = 0; i < 32; i++) {
	  text += possible.charAt(Math.floor(Math.random() * possible.length));
	}
	return text;
}