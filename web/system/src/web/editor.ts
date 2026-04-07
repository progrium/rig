import * as vscode from 'vscode';
import * as duplex from "@progrium/duplex";
import { manifold } from '../webview/webview.js';

export async function activate(ctx: vscode.ExtensionContext, token: string, peer: duplex.Peer, realm: manifold.Realm) {
    ctx.subscriptions.push(
        vscode.commands.registerCommand('rig.open', async (id: string) => {
            // await vscode.commands.executeCommand('vscode.openWith', id, 'rig.manifoldEditor');
            EditorPanel.createOrShow(ctx.extensionUri, token, id);
        }),

    );    
}

export class EditorPanel {
	public static panels: Map<string, EditorPanel> = new Map();

	public static readonly viewType = 'rigEditor';

	private readonly panel: vscode.WebviewPanel;
	private readonly extensionUri: vscode.Uri;
    private readonly websocketURL: string;
    private nodeID: string;
	private disposables: vscode.Disposable[] = [];

	public static createOrShow(extensionUri: vscode.Uri, token: string, nodeID: string) {
		const column = vscode.window.activeTextEditor
			? vscode.window.activeTextEditor.viewColumn
			: undefined;

		// If we already have a panel, show it.
		if (EditorPanel.panels.get(nodeID)) {
			EditorPanel.panels.get(nodeID)?.panel.reveal(column);
			return;
		}

		// Otherwise, create a new panel.
		const panel = vscode.window.createWebviewPanel(
			EditorPanel.viewType,
			'Editor',
			column || vscode.ViewColumn.One,
			{
                enableScripts: true,
            },
		);

		EditorPanel.panels.set(nodeID, new EditorPanel(panel, extensionUri, "ws://localhost:8080/inspector/"+token, nodeID));
	}

	// public static revive(panel: vscode.WebviewPanel, extensionUri: vscode.Uri) {
	// 	EditorPanel.currentPanel = new EditorPanel(panel, extensionUri);
	// }

	constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri, websocketURL: string, nodeID: string) {
		this.panel = panel;
		this.extensionUri = extensionUri;
        this.nodeID = nodeID;
        this.websocketURL = websocketURL;

		// Set the webview's initial html content
		this._setup();

		// Listen for when the panel is disposed
		// This happens when the user closes the panel or when the panel is closed programmatically
		this.panel.onDidDispose(() => this.dispose(), null, this.disposables);

		// Update the content based on view changes
		this.panel.onDidChangeViewState(
			() => {
				if (this.panel.visible) {
					this._setup();
				}
			},
			null,
			this.disposables
		);

		// Handle messages from the webview
		this.panel.webview.onDidReceiveMessage(
			message => {
				switch (message.command) {
					case 'alert':
						vscode.window.showErrorMessage(message.text);
						return;
				}
			},
			null,
			this.disposables
		);
	}

	public dispose() {
		EditorPanel.panels.delete(this.nodeID);

		// Clean up our resources
		this.panel.dispose();

		while (this.disposables.length) {
			const x = this.disposables.pop();
			if (x) {
				x.dispose();
			}
		}
	}


    public selectNode(id: string) {
		this.nodeID = id;
		this.panel.webview.postMessage({
			manifold: "selectNode",
			id: this.nodeID,
		});
	}

	private _setup() {
        this.panel.iconPath = {
            light: vscode.Uri.joinPath(this.extensionUri, 'media', 'box-light.svg'),
            dark: vscode.Uri.joinPath(this.extensionUri, 'media', 'box-dark.svg'),
        };
        const nonce = getNonce();
        this.panel.webview.onDidReceiveMessage(data => {
			if (data.action) {
				vscode.commands.executeCommand(data.action, ...(data.args||[]))
				return;
			}
			if (data.ready) {
				this.selectNode("@main");
				return;
			}
		});
		this.panel.webview.html = `
<html>
<head>
	<meta http-equiv="Content-Security-Policy" content="
		default-src 'none';
		script-src 'nonce-${nonce}' http://localhost:8080;
		connect-src http://localhost:8080 ws://localhost:8080;
		style-src http://localhost:8080 'unsafe-inline';
		font-src http://localhost:8080;
	">
	<link rel="stylesheet" href="${this.extensionUri.with({path: "system/media/fontawesome/css/all.min.css"}).toString()}">
	<link rel="stylesheet" href="${this.extensionUri.with({path: "system/media/inspector.css"}).toString()}">
</head>
<body>
<script nonce="${nonce}" type="module">
import {
	m,
	manifold,
	duplex,
	util,
    inspector
} from "${this.extensionUri.with({path: "system/dist/webview/webview.js"}).toString()}";

  window.m = m;
  window.realm = new manifold.Realm();
  window.vscode = acquireVsCodeApi();
  
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
  
  peer.handle("update", duplex.HandlerFunc(async (r, c) => {
    const update = await c.receive();
    // console.log("update:", Object.keys(update).length);
    window.realm.update(update);
    m.redraw();
  }));
  peer.respond();

  var selectedID = "@main";
  var mounted = false;
  window.addEventListener("message", (e) => {
      if (!e.data.manifold) {
        return;
      }
      selectedID = e.data.id;
      if (!mounted) {
        m.mount(document.body, App);
        mounted = true;
      }
      m.redraw();
  });
  vscode.postMessage({ready: true});

  import {Editor} from "${this.extensionUri.with({path: "editors/rig/editor.js"}).toString()}";
  
  const App = {
    view: () => 
      	m("main", {style: {position: "absolute", left: "0", top: "0", left: "0", right: "0"}}, [
            m(Editor, {nodeID: selectedID, vscode, peer, realm, m})
		])
  }
</script>
</body>
</html>`;
		// Vary the webview's content based on where it is located in the editor.
		// switch (this._panel.viewColumn) {
		// 	case vscode.ViewColumn.Two:
		// 		this._updateForCat(webview, 'Compiling Cat');
		// 		return;

		// 	case vscode.ViewColumn.Three:
		// 		this._updateForCat(webview, 'Testing Cat');
		// 		return;

		// 	case vscode.ViewColumn.One:
		// 	default:
		// 		this._updateForCat(webview, 'Coding Cat');
		// 		return;
		// }
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