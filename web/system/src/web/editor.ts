import * as vscode from 'vscode';
import * as duplex from "@progrium/duplex";
import { manifold } from '../webview/webview.js';

const titleize = (str: string): string =>
    str.replace(/\b\w+/g, w => w[0].toUpperCase() + w.slice(1).toLowerCase());

export async function activate(ctx: vscode.ExtensionContext, token: string, peer: duplex.Peer, realm: manifold.Realm) {
    ctx.subscriptions.push(
        vscode.commands.registerCommand('rig.open', async (id: string) => {
            // await vscode.commands.executeCommand('vscode.openWith', id, 'rig.manifoldEditor');
            EditorPanel.createOrShow(ctx.extensionUri, token, id);
        }),
        vscode.commands.registerCommand('rig.reload', async (id: string) => {
            for (const panel of EditorPanel.panels.values()) {
                panel.reload();
            }
        }),
        vscode.commands.registerCommand('rig.edit', async (id: string) => {
            const resp = await peer.call("listEditors", []);
            const editors = (resp.value as string[]).map(e => `${titleize(e)} Editor`);
            const sel = await vscode.window.showQuickPick(editors);
            if (!sel) {
                return;
            }
            console.log("editing", id, "with", sel);
            EditorPanel.createOrShow(ctx.extensionUri, token, id, sel.replace(" Editor", "").toLowerCase());
        }),
        vscode.commands.registerCommand('rig.createEditor', async (id: string) => {
            const name = await vscode.window.showInputBox({title: "Editor Name"});
            if (!name) {
                return;
            }

            const workspaceFolder = vscode.workspace.workspaceFolders?.[0].uri;
            if (!workspaceFolder) return;

            const dirUri = vscode.Uri.joinPath(workspaceFolder, "root/go/src/github.com/progrium/rig/web/editors", name);
            await vscode.workspace.fs.createDirectory(dirUri);

            const fileUri = vscode.Uri.joinPath(workspaceFolder, "root/go/src/github.com/progrium/rig/web/editors", name, "editor.jsx");
            const content = new TextEncoder().encode(`export const Editor = {
    view: ({attrs: {vscode,realm,nodeID}}) => {
        const node = realm.resolve(nodeID);
        if (node === null) {
            return <div>No node</div>;
        }
        return <div>Hello, world: {node.name}</div>;
    }
}`);
            await vscode.workspace.fs.writeFile(fileUri, content);
            
            const document = await vscode.workspace.openTextDocument(fileUri);
            await vscode.window.showTextDocument(document);
        }),
    );    
    const resp = await peer.call("watchFile", "/go/src/github.com/progrium/rig/web/editors");
    (async () => {
        while (true) {
            const path = await resp.receive();
            if (path === null) {
                break;
            }
            vscode.commands.executeCommand('rig.reload');
        }
    })();
}

export class EditorPanel {
	public static panels: Map<string, EditorPanel> = new Map();

	public static readonly viewType = 'rigEditor';

	private readonly panel: vscode.WebviewPanel;
	private readonly extensionUri: vscode.Uri;
    private readonly websocketURL: string;
    private readonly editor: string;
    private nodeID: string;
	private disposables: vscode.Disposable[] = [];

	public static createOrShow(extensionUri: vscode.Uri, token: string, nodeID: string, editor: string = "rig") {
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

		EditorPanel.panels.set(nodeID, new EditorPanel(panel, extensionUri, "ws://localhost:8080/inspector/"+token, editor, nodeID));
	}

	constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri, websocketURL: string, editor: string, nodeID: string) {
		this.panel = panel;
        this.editor = editor;
		this.extensionUri = extensionUri;
        this.nodeID = nodeID;
        this.websocketURL = websocketURL;

		// Set the webview's initial html content
		this.reload();

		// Listen for when the panel is disposed
		// This happens when the user closes the panel or when the panel is closed programmatically
		this.panel.onDidDispose(() => this.dispose(), null, this.disposables);

		// Update the content based on view changes
		// this.panel.onDidChangeViewState(
		// 	(e) => {
		// 		if (this.panel.visible) {
		// 			this.reload();
		// 		}
		// 	},
		// 	null,
		// 	this.disposables
		// );

        this.panel.iconPath = {
            light: vscode.Uri.joinPath(this.extensionUri, 'media', 'box-light.svg'),
            dark: vscode.Uri.joinPath(this.extensionUri, 'media', 'box-dark.svg'),
        };
        this.panel.webview.onDidReceiveMessage(data => {
			if (data.action) {
				vscode.commands.executeCommand(data.action, ...(data.args||[]))
				return;
			}
			if (data.ready) {
				this.selectNode(this.nodeID);
				return;
			}
		}, null, this.disposables);
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

	public reload() {
        const nonce = getNonce();
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
	<link rel="stylesheet" href="${this.extensionUri.with({path: "editors/editor.css"}).toString()}">
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
  window.realm = new manifold.Realm(peer);
  window.vscode = acquireVsCodeApi();
  
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

  import {Editor} from "${this.extensionUri.with({path: `editors/${this.editor}/editor.jsx`}).toString()}";
  
  const App = {view: () => m(Editor, {nodeID: selectedID, vscode, peer, realm, m})}
</script>
</body>
</html>`;
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