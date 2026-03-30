
import * as vscode from 'vscode';


export class InspectorViewProvider implements vscode.WebviewViewProvider {

	public static readonly viewType = 'manifold-sidebar.inspector';

	private _view?: vscode.WebviewView;
	private _selectedID?: string;

	constructor(
		private readonly _extensionUri: vscode.Uri,
		private readonly iframeURL: string = "about:none",
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
		const resp = await this.peer.call("Fields", this._selectedID);
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

		webviewView.webview.html = `
    <html>
    <head>
      <style>
        html {
          height: 100%;
        }
        body {
          width: 100%;
          height: 100%;
          margin: 0;
          padding: 0;
		  display: flex;
		  flex-direction: column;
        }
        iframe {
          width: 100%;
          height: 100%;
          border: 0;
        }
      </style>
    </head>
    <body>
    <iframe id="frame" src="${this.iframeURL}">
    </iframe>
	<button style="display: none;" id="menu-more" data-vscode-context='{"webviewSection": "more",  "preventDefaultContextMenuItems": true}'></button>
	<button style="display: none;" id="menu-ref" data-vscode-context='{"webviewSection": "ref",  "preventDefaultContextMenuItems": true}'></button>
	<script type="text/javascript">
	const vscode = acquireVsCodeApi();
	window.addEventListener("message", e => {
		if (e.data.manifold) {
			frame.contentWindow.postMessage(e.data, "${this.iframeURL}");
			return;
		}
		if (e.data.menu) {
			const button = document.getElementById(e.data.menu);
			const ctx = JSON.parse(button.dataset.vscodeContext);
			Object.assign(ctx, e.data.params);
			button.dataset.vscodeContext = JSON.stringify(ctx);
			button.dispatchEvent(new MouseEvent('contextmenu', {
    			bubbles: true, 
    			cancelable: true, 
    			view: window,
				clientX: e.data.params.clientX, 
				clientY: e.data.params.clientY,
			}));
			return;
		}
		if (e.data.action) {
			vscode.postMessage(e.data);
		}
	});
	vscode.postMessage({ready: true});
	</script>
    </body>
    </html>`;

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