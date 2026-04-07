import * as vscode from 'vscode';
import * as duplex from "@progrium/duplex";
import { manifold } from '../webview/webview.js';

export async function activate(ctx: vscode.ExtensionContext, peer: duplex.Peer, realm: manifold.Realm) {
    ctx.subscriptions.push(
        vscode.commands.registerCommand('rig.open', async (id: string) => {
            await vscode.commands.executeCommand('vscode.openWith', rigURI(id), 'rig.manifoldEditor');
        }),
        ManifoldEditorProvider.register(ctx, peer, realm),
        URIProvider.register(ctx),
    );    
}

export function rigURI(nodeID: string): vscode.Uri {
    return vscode.Uri.from({
      scheme: 'rig',
      path: `/${nodeID}.rig`,
    });
  }

export class ManifoldView implements vscode.CustomDocument {
  constructor(
    readonly uri: vscode.Uri, 
    private readonly nodeID: string, 
    private readonly peer: duplex.Peer, 
    private readonly realm: manifold.Realm) {}
  dispose() {}
}

export class ManifoldEditorProvider implements vscode.CustomEditorProvider<ManifoldView> {

  static register(context: vscode.ExtensionContext, peer: duplex.Peer, realm: manifold.Realm) {
    return vscode.window.registerCustomEditorProvider(
      'rig.manifoldEditor',
      new ManifoldEditorProvider(peer, realm),
    );
  }

  constructor(private readonly peer: duplex.Peer, private readonly realm: manifold.Realm) {}

  openCustomDocument(uri: vscode.Uri): ManifoldView {
    const nodeID = uri.path.replace(".rig", "").slice(1); // remove the leading '/'
    return new ManifoldView(uri, nodeID, this.peer, this.realm);
  }

  resolveCustomEditor(document: ManifoldView, panel: vscode.WebviewPanel): void {
        panel.webview.options = {
            enableScripts: true,
        };
        panel.title = "EDIT";
        panel.webview.html = `<!DOCTYPE html>
        <html>
            <body>Opened: ${document.uri.toString()}</body>
        </html>`;
        console.log("resolveCustomEditor");
  }

  private readonly _onDidChangeCustomDocument = new vscode.EventEmitter<vscode.CustomDocumentEditEvent<ManifoldView>>();
  readonly onDidChangeCustomDocument = this._onDidChangeCustomDocument.event;

  saveCustomDocument(document: ManifoldView, cancellation: vscode.CancellationToken): Thenable<void> {
    return Promise.resolve();
  }

  saveCustomDocumentAs(document: ManifoldView, destination: vscode.Uri, cancellation: vscode.CancellationToken): Thenable<void> {
    return Promise.resolve();
  }

  revertCustomDocument(document: ManifoldView, cancellation: vscode.CancellationToken): Thenable<void> {
    return Promise.resolve();
  }

  backupCustomDocument(document: ManifoldView, context: vscode.CustomDocumentBackupContext, cancellation: vscode.CancellationToken): Thenable<vscode.CustomDocumentBackup> {
    return Promise.resolve({ id: context.destination.toString(), delete: () => {} });
  }
}

export class URIProvider implements vscode.FileSystemProvider {

    private readonly _onDidChangeFile = new vscode.EventEmitter<vscode.FileChangeEvent[]>();
    readonly onDidChangeFile = this._onDidChangeFile.event;
  
    static register(context: vscode.ExtensionContext) {
      return vscode.workspace.registerFileSystemProvider(
        'rig',
        new URIProvider(),
        { isCaseSensitive: true, isReadonly: true }
      );
    }
  
    stat(): vscode.FileStat {
      return { type: vscode.FileType.File, ctime: 0, mtime: 0, size: 0 };
    }
  
    readFile(): Uint8Array { return new Uint8Array(); }
    writeFile(): void {}
    readDirectory(): never { throw vscode.FileSystemError.NoPermissions(); }
    createDirectory(): never { throw vscode.FileSystemError.NoPermissions(); }
    delete(): never { throw vscode.FileSystemError.NoPermissions(); }
    rename(): never { throw vscode.FileSystemError.NoPermissions(); }
    watch(): vscode.Disposable { return new vscode.Disposable(() => {}); }
}