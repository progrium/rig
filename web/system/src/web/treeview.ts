
import * as vscode from 'vscode';
import * as duplex from "@progrium/duplex";

import { TempFileSystemProvider } from './bufferfs';
import { InspectorViewProvider } from './inspector';

var peer: duplex.Peer;

export async function activate(context: vscode.ExtensionContext, fsys: any) {

    context.subscriptions.push(
      vscode.commands.registerCommand('manifold.openBrowser', (url?: string) => {
          const panel = vscode.window.createWebviewPanel(
              'embeddedBrowser', // Identifies the type of the webview. Used internally
              'Preview', // Title of the panel displayed to the user
              vscode.ViewColumn.One, // Editor column to show the new webview panel in.
              {
                  enableScripts: true
              }
          );

          panel.webview.html = getWebviewContent(url);

          // Handle messages from the webview
          panel.webview.onDidReceiveMessage(
              message => {
                  switch (message.command) {
                      case 'updateIframe':
                          panel.webview.html = getWebviewContent(message.url);
                          return;
                  }
              },
              undefined,
              context.subscriptions
          );
      })
  );

    // const logs = vscode.window.createOutputChannel("Manifold");
    // context.subscriptions.push(logs);
    

    var inspector: InspectorViewProvider;
    const conn: File = await connectWithRetry(fsys, "/var/run/manifold.sock", (reconn: File) => {
      console.warn("reconnected, not implemented!")
      // const sess = new duplex.Session(new SocketConn(reconn));
      // if (!peer) {
      //   peer = new duplex.Peer(sess, new duplex.CBORCodec());
      // } else {
      //   console.log("updating session:", sess);
      //   peer.session = sess;
      //   peer.caller = new duplex.Client(sess, new duplex.CBORCodec());
      //   peer.call("session", []);
      //   // startOutput(logs, peer);
      //   if (inspector) {
      //     inspector.reload();
      //   }
      //   peer.respond();
      // }
    });

    const sess = new duplex.Session(conn);
	  peer = new duplex.Peer(sess, new duplex.CBORCodec());
    peer.call("session", []);
    // startOutput(logs, peer);

    inspector = new InspectorViewProvider(context.extensionUri, "http://localhost:8080/-/inspector/inspector.html", peer);
    context.subscriptions.push(
      vscode.window.registerWebviewViewProvider(InspectorViewProvider.viewType, inspector));
    context.subscriptions.push(
        vscode.commands.registerCommand('manifold.inspect', (id: string) => {
            inspector.selectNode(id);
        })
    );

    const treeView = new TreeView(peer);
    const view = vscode.window.createTreeView('manifold.hierarchy', { treeDataProvider: treeView, showCollapseAll: true, canSelectMany: true});
	  context.subscriptions.push(view);

    peer.handle("signaled", duplex.HandlerFunc(async (r: duplex.Responder, c: duplex.Call) => {
      const [id] = await c.receive();
      treeView._onDidChangeTreeData.fire([id]);
      inspector.reload();
    }));
    peer.handle("execCommand", duplex.HandlerFunc(async (r: duplex.Responder, c: duplex.Call) => {
      const [cmd, args] = await c.receive();
      vscode.commands.executeCommand(cmd, ...args);
    }));
    peer.respond();

    
    
    
    view.onDidChangeCheckboxState(async (event) => {
      event.items.forEach(async (item) => {
        await peer.call("Toggle", [item[0], item[1] === 1]);
        treeView._onDidChangeTreeData.fire([item[0]]);
      });
    });
    view.onDidChangeSelection(async (event) => {
      const selected = event.selection[0];
      if (selected) {
        inspector.selectNode(selected);
        peer.call("Select", [selected]);
      }
    });

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.activate', async (id: string) => {
        await peer.call("Toggle", [id, true]);
        treeView._onDidChangeTreeData.fire([id]);
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.activateAll', async (id: string) => {
        await peer.call("Toggle", ["", true]);
        treeView._onDidChangeTreeData.fire(undefined);  
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.viewSource', async (id: string) => {
        if (!id) return;
        const resp = await peer.call("GetTreeItem", [id]);
        if (!resp.value.description) {
          return;
        }
        openComponentSymbol(resp.value.description);
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.fullReload', async () => {
        await peer.call("Toggle", ["", false]);
        vscode.commands.executeCommand('workbench.action.reloadWindow');
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.deactivate', async (id: string) => {
            await peer.call("Toggle", [id, false]);
            treeView._onDidChangeTreeData.fire([id]);
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.deactivateAll', async (id: string) => {
        await peer.call("Toggle", ["", false]);
        treeView._onDidChangeTreeData.fire(undefined);  
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.callMethod', async (id: string) => {
        const resp = await peer.call("Methods", [id]);
        if (!resp.value) {
          return;
        }
        const sel = await vscode.window.showQuickPick(resp.value as vscode.QuickPickItem[]);
        if (!sel) {
          return;
        }
        await peer.call("Call", [(sel as any)["id"]]);
        treeView._onDidChangeTreeData.fire(undefined);
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.switchView', async (id: string) => {
            const resp = await peer.call("Views", [id]);
            const sel = await vscode.window.showQuickPick(resp.value);
            await peer.call("SwitchView", [id, sel]);
            treeView._onDidChangeTreeData.fire([id])
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.add', async (id: string) => {
            const resp = await peer.call("GetAddItems", [id]);
            const items = resp.value.map((item: string) => {
              if (item.startsWith("--")) {
                return {kind: -1, label: item.slice(2)};
              }
              const parts = item.split("/");
              return {label: parts[parts.length-1], detail: item};
            }) as vscode.QuickPickItem[];
            const sel = await vscode.window.showQuickPick(items);
            if (!sel) {
              return;
            }
            const tmpName = sel.label.slice(sel.label.indexOf(".")+1).replace("()", "");
            const name = await vscode.window.showInputBox({title: "Name", value: tmpName});
            await peer.call("AddItem", [id, sel.detail, name]);
            treeView._onDidChangeTreeData.fire([id]);
	    }
    ));

    

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.addComponent', async (id: any) => {
            const resp = await peer.call("GetAddComponents", [id]);
            const items = resp.value.map((item: string) => {
              if (item.startsWith("--")) {
                return {kind: -1, label: item.slice(2)};
              }
              if (item.startsWith("main.")) {
                return {label: item, id: item};
              }
              const parts = item.split("/");
              return {label: parts[parts.length-1], detail: item, id: item};
            }) as (vscode.QuickPickItem & {id: string})[];
            const sel = await vscode.window.showQuickPick(items);
            if (!sel) {
              return;
            }
            await peer.call("AddComponent", [id, sel.id]);
            treeView._onDidChangeTreeData.fire([id]);
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.editField', async (id: string) => {
            const resp = await peer.call("PreEditField", [id]);
            if (!resp.value["OK"]) {
              return;
            }
            const newval = await vscode.window.showInputBox({title: resp.value["Name"] as string, value: resp.value["Value"]});
            await peer.call("EditField", [id, newval]);
            treeView._onDidChangeTreeData.fire([resp.value["ParentID"]]);
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.rename', async (id: string) => {
            const resp = await peer.call("GetTreeItem", [id]);
            if (!resp.value) {
              return;
            }
            const newval = await vscode.window.showInputBox({title: "Rename", value: resp.value["label"]});
            await peer.call("Rename", [id, newval]);
            treeView._onDidChangeTreeData.fire([id]);
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.call', async (id: string) => {
            await peer.call("Call", [id]);
            treeView._onDidChangeTreeData.fire(undefined); 
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.delete', async (id: string) => {
            await peer.call("Delete", [id]);
            treeView._onDidChangeTreeData.fire(undefined); 
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.implementFor', async (iface: string, fieldID?: string) => {
          const newComLabel = "New component...";
          const items = [{label: newComLabel}, {kind: -1, label: "Main"}];
          let name = undefined;

          const coms = await peer.call("GetMainComponents", []);
          if (coms.value && coms.value.length > 0) {
            (coms.value||[]).forEach((sympath: any) => {
              items.push({label: sympath});
            });
            const sel = await vscode.window.showQuickPick(items);
            if (!sel) {
              return;
            }
            name = sel.label;

            // was going to use this to determine the filepath of the existing
            // component and then in ImplementFor it would append to the filepath
            // instead of making a new file... but for now we can live without:
            // console.log(await filepathForSymbol(sel.label));
          }
          
          if (!name || name === newComLabel) {
            const input = await vscode.window.showInputBox({title: "Component Name"});
            if (!input) {
              return;
            }
            name = input;
          }
          
          const resp = await peer.call("ImplementFor", [iface, name, fieldID]);
          if (resp.value) {
            await vscode.workspace.openTextDocument(resp.value.FilePath);
            setTimeout(async () => {
              if (!(await openComponentSymbol(resp.value.PkgPath))) {
                vscode.window.showTextDocument(vscode.Uri.file(resp.value.FilePath));
              }
            }, 500);
          }
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.deleteFromInspector', (ctx) => {
        vscode.commands.executeCommand('manifold.delete', ctx.id)
      }))

    context.subscriptions.push(vscode.commands.registerCommand(
      'manifold.callMethodFromInspector', (ctx) => {
        vscode.commands.executeCommand('manifold.callMethod', ctx.id)
      }))
    
    context.subscriptions.push(vscode.commands.registerCommand(
      'manifold.viewSourceFromInspector', (ctx) => {
        vscode.commands.executeCommand('manifold.viewSource', ctx.id)
      }))
    
    context.subscriptions.push(vscode.commands.registerCommand(
      'manifold.implementFromInspector', (ctx) => {
        vscode.commands.executeCommand('manifold.implementFor', ctx.iface, ctx.fieldID)
      }))

    context.subscriptions.push(vscode.commands.registerCommand(
      'manifold.newComponent', async (id: string) => {
        const item = await peer.call("GetTreeItem", [id]);
        if (!item.value) {
          return;
        }
        const name = await vscode.window.showInputBox({title: "Component Name", value: item.value["label"]});
        const com = await peer.call("NewComponent", [id, name]); 
        if (com.value) {
          await vscode.workspace.openTextDocument(com.value.FilePath);
          setTimeout(() => {
            openComponentSymbol(com.value.PkgPath);
          }, 200);
        }
        treeView._onDidChangeTreeData.fire([id]); 
      }
    ));


    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.toggleComponents', async (id: string) => {
            await peer.call("ToggleComponents", [id]);
            treeView._onDidChangeTreeData.fire([id]); 
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.signal', async (id: string) => {
            const signal = await vscode.window.showInputBox({title: "Signal"});
            await peer.call("Signal", [id, signal]);
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.dump', async (id: string) => {
            await peer.call("Dump", [id]);
	    }
    ));

    context.subscriptions.push(vscode.commands.registerCommand(
	    'manifold.reload', async (id: string) => {
          treeView._onDidChangeTreeData.fire([id]);
	    }
    ));


    // buffer editing

    const tempFS = new TempFileSystemProvider();
    const scheme = 'tmpfs';
    const registeredListeners = new Set<string>();
    context.subscriptions.push(vscode.workspace.registerFileSystemProvider(scheme, tempFS, { isCaseSensitive: true }));
    context.subscriptions.push(vscode.commands.registerCommand('manifold.editBuffer', async (path: string, data: Uint8Array) => {
      const uri = vscode.Uri.parse(`${scheme}:/${path}`);
      tempFS.writeFile(uri, data, { create: true, overwrite: true });
      if (!registeredListeners.has(uri.toString())) {
        tempFS.onDidChangeFile((events) => {
          for (const event of events) {
            if (event.uri.toString() === uri.toString() && event.type === vscode.FileChangeType.Changed) {
              const data = tempFS.readFile(uri);
              peer.call("bridge.BufferChanged", [path, data]);
            }
          }
        });
        registeredListeners.add(uri.toString());
      }
      const document = await vscode.workspace.openTextDocument(uri);
      await vscode.window.showTextDocument(document);
  }));
}



const treeViewTimeout = 1000;
export class TreeView implements vscode.TreeDataProvider<string> {
    peer: any;

    public _onDidChangeTreeData: vscode.EventEmitter<(string | undefined)[] | undefined> = new vscode.EventEmitter<string[] | undefined>();
	public onDidChangeTreeData: vscode.Event<any> = this._onDidChangeTreeData.event;

    constructor(peer: any) {
        this.peer = peer;
    }

    public async getChildren(id?: string|null): Promise<string[]> {
      let retries = 0;
      const attemptCall = () => new Promise<string[]>(async (resolve,reject) => {
        (new Promise<string[]>((res, rej) => {
          setTimeout(async () => {
            if (retries >= 2) {
              rej(new Error("call timed out"));
              return;
            }
            retries++;
            res(attemptCall());
          }, treeViewTimeout);
        })).then(resolve).catch(reject);
        
        try {
          resolve(await this._getChildren(id));
        } catch (e) {} // ignore exception, we will just retry

      });
      return await attemptCall();
    }

    private async _getChildren(id?: string|null): Promise<string[]> {
      try {
        const resp = await this.peer.call("GetChildren", [(id)?id:""]);
        return (resp.value||[]).map((n: any) => n.id);
      } catch (e) {
        return [];
      }
    }

    public async getTreeItem(id: string): Promise<vscode.TreeItem> {
      let retries = 0;
      const attemptCall = () => new Promise<vscode.TreeItem>(async (resolve,reject) => {
        (new Promise<vscode.TreeItem>((res, rej) => {
          setTimeout(async () => {
            if (retries >= 2) {
              rej(new Error("call timed out"));
              return;
            }
            retries++;
            res(attemptCall());
          }, treeViewTimeout);
        })).then(resolve).catch(reject);
        
        try {
          resolve(await this._getTreeItem(id));
        } catch (e) {} // ignore exception, we will just retry

      });
      return await attemptCall();
    }

    private async _getTreeItem(id: string): Promise<vscode.TreeItem> {
        const resp = await this.peer.call("GetTreeItem", [id]);
        const element = resp.value;
        let icon = "";
        let color = undefined;
        let attrNode = element as {Attrs: any};
        if (attrNode.Attrs) {
            switch (attrNode.Attrs["view"]) {
            case "components":
                //icon = "gear";
                // if (!attrNode.Attrs["desc"]) {
                //   element.description = "(components)";
                // }
                break;
            // case "fields":
                // icon = "symbol-object";
                // break;
            default:
                icon = "";
                if (attrNode.Attrs["view"] !== "objects" && !element.description) {
                  element.description = attrNode.Attrs["view"];
                }
            }
            icon = icon || attrNode.Attrs["icon"];
            if (attrNode.Attrs["color"]) {
                color = new vscode.ThemeColor(attrNode.Attrs["color"]);
            }
            if (attrNode.Attrs["busy"]) {
              icon = "sync~spin";
              color = new vscode.ThemeColor("activityBar.inactiveForeground");
            }
        }
        return Object.assign({}, element, {
            tooltip: (element.tooltip) ? new vscode.MarkdownString(element.tooltip as string) : undefined,
            iconPath: (icon) ? new vscode.ThemeIcon(icon, color) : undefined
        });
    }
}

class File {
    fsys: any;
    fd: number;
    isClosed: boolean;
  
    constructor(fsys: any, fd: number) {
      this.isClosed = false;
      this.fsys = fsys;
      this.fd = fd;
    }
  
    async read(p: Uint8Array): Promise<number | null> {
      if (this.isClosed) {
        throw new Error("file is closed");
      }
      const data = await this.fsys.read(this.fd, p.byteLength);
      if (data === null) {
        return null;
      }
      p.set(new Uint8Array(data), 0);
      return data.byteLength;
    }
  
    write(p: Uint8Array): Promise<number> {
      if (this.isClosed) {
        throw new Error("file is closed");
      }
      return this.fsys.write(this.fd, p);
    }
  
    async close(): Promise<void> {
      if (this.isClosed) {
        return Promise.resolve();
      }
      await this.fsys.close(this.fd);
      this.isClosed = true;
    }
  }


function getWebviewContent(initialUrl: string = 'https://www.example.com') {
  return `<!DOCTYPE html>
  <html lang="en">
  <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>Embedded Browser</title>
  </head>
  <body style="margin:0;">
      <div style="display: flex; flex-direction: row;">
        <input id="urlInput" type="text" value="${initialUrl}" style="flex-grow: 1; background-color: transparent; border: 0; color: #ccc;" />
        <button onclick="updateIframe()">Go</button>
      </div>
      <iframe id="browserFrame" src="${initialUrl}" style="width: 100%; height: 90vh; background-color: white; border: 0;"></iframe>
      <script>
          const vscode = acquireVsCodeApi();

          function updateIframe() {
              const url = document.getElementById('urlInput').value;
              if (document.getElementById('browserFrame').src === url) {
                document.getElementById('browserFrame').src = "about:blank";
              }
              document.getElementById('browserFrame').src = url;
              // vscode.postMessage({
              //     command: 'updateIframe',
              //     url: url
              // });
          }
      </script>
  </body>
  </html>`;
}

async function startOutput(logs: vscode.OutputChannel, peer: duplex.Peer) {
  // send subscribe to logs and send to output
  const dec = new TextDecoder();
  console.log("starting logs...");
  const resp = await peer.call("logs", []);
  console.log("receiving logs...");
  duplex.copy({write: async (p: Uint8Array): Promise<number> => {
    logs.append(dec.decode(p));
    return p.byteLength;
  }}, resp.channel);
}

const maxRetries = 10;
async function connectWithRetry(fsys: any, socketPath: string, onReconnect?: (s: File) => void): Promise<File> {
    let retryCount = 0;
    let connected = false;

    return new Promise((resolve, reject) => {
        async function attemptConnection() {
            const id = (await fsys.readText("net/tcp/new")).trim();
            await fsys.writeFile(`net/tcp/${id}/ctl`, "dial /var/run/manifold.sock\n");
            const fd = await fsys.open(`net/tcp/${id}/data`);
            const file = new File(fsys, fd);
            console.log('Connected to server');  
            retryCount = 0;
            connected = true;
            resolve(file);

            // socket.on('error', (err) => {
            //     // console.log('Connection error:', err.message);
            //     socket.destroy(); // Close the socket on error
            //     if (retryCount < maxRetries) {
            //         retryCount++;
            //         console.log(`Retrying in 2 second... (${retryCount}/${maxRetries})`);
            //         setTimeout(attemptConnection, 2000); // Retry after 1 second
            //     } else {
            //         reject(new Error('Max retries reached. Unable to connect.'));
            //     }
            // });

            // socket.on('close', () => {
            //     if (!connected) {
            //       return;
            //     }
            //     console.log(`Connection closed. Reconnecting in 1 second... (${retryCount}/${maxRetries})`);
            //     setTimeout(async () => {
            //       const socket = await connectWithRetry(socketPath, onReconnect);
            //       if (onReconnect) onReconnect(socket);
            //     }, 1000); 
            // });
        }

        attemptConnection(); // Start the initial connection attempt
    });
}

async function openComponentSymbol(symbolPath: string): Promise<Boolean> {
  const symbols = await vscode.commands.executeCommand<vscode.SymbolInformation[]>('vscode.executeWorkspaceSymbolProvider', symbolPath);
  
  // Assuming the first symbol is the desired one. 22 is kind "Struct"
  let symbol = symbols.filter(s => s.kind === 22 && symbolPath.endsWith(s.name))[0];
  if (!symbol) {
    return false;
  }
    
  // Create a new range to highlight the symbol
  let range = new vscode.Range(symbol.location.range.start, symbol.location.range.start);

  // Reveal the symbol in the corresponding document
  vscode.window.showTextDocument(symbol.location.uri, { selection: range });

  return true;
}

async function filepathForSymbol(symbolPath: string): Promise<string> {
  let symbols = await vscode.commands.executeCommand<vscode.SymbolInformation[]>('vscode.executeWorkspaceSymbolProvider', symbolPath);
  if (symbols.length === 0) {
    return "";
  }
  
  // Assuming the first symbol is the desired one. 22 is kind "Struct"
  let symbol = symbols.filter(s => s.kind === 22)[0];

  return symbol.location.uri.fsPath;
}