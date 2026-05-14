
import * as vscode from 'vscode';
import * as duplex from "@progrium/duplex";
import { manifold } from '../webview/webview.js';

export async function activate(ctx: vscode.ExtensionContext, fsys: any, peer: duplex.Peer, realm: manifold.Realm) {
    ctx.subscriptions.push(
        
        vscode.commands.registerCommand('manifold.activate', async (id: string) => {
            await peer.call("Toggle", [id, true]);
            // treeView._onDidChangeTreeData.fire([id]);
        }),
        vscode.commands.registerCommand('manifold.activateAll', async (id: string) => {
            await peer.call("Toggle", ["", true]);
            // treeView._onDidChangeTreeData.fire(undefined);  
        }),
        vscode.commands.registerCommand('manifold.viewSource', async (id: string) => {
            const com = realm.resolve(id)?.raw.Component;
            if (!com) return;
            const [path, type] = com.split(".");

            const workspaceFolder = vscode.workspace.workspaceFolders?.[0].uri;
            if (!workspaceFolder) return;
            const fileUri = vscode.Uri.joinPath(workspaceFolder, "root/go/src", path);
            const document = await vscode.workspace.openTextDocument(fileUri);
            await vscode.window.showTextDocument(document);
            // console.log("viewSource", id);
            // if (!id) return;
            // const resp = await peer.call("GetTreeItem", [id]);
            // if (!resp.value.description) {
            //   return;
            // }
            // filepathForSymbol("main.Main");
            // openComponentSymbol("main.Main");
        }),
        vscode.commands.registerCommand('manifold.fullReload', async () => {
            await peer.call("Toggle", ["", false]);
            vscode.commands.executeCommand('workbench.action.reloadWindow');
        }),
        vscode.commands.registerCommand('manifold.deactivate', async (id: string) => {
            await peer.call("Toggle", [id, false]);
            // treeView._onDidChangeTreeData.fire([id]);
        }),
        vscode.commands.registerCommand('manifold.deactivateAll', async (id: string) => {
            await peer.call("Toggle", ["", false]);
            // treeView._onDidChangeTreeData.fire(undefined);  
        }),
        vscode.commands.registerCommand('manifold.callMethod', async (id: string) => {
            const resp = await peer.call("Methods", [id]);
            if (!resp.value) {
              return;
            }
            const sel = await vscode.window.showQuickPick(resp.value as vscode.QuickPickItem[]);
            if (!sel) {
              return;
            }
            await peer.call("Call", [(sel as any)["id"]]);
            // treeView._onDidChangeTreeData.fire(undefined);
        }),
        vscode.commands.registerCommand('manifold.switchView', async (id: string) => {
            const resp = await peer.call("Views", [id]);
            const sel = await vscode.window.showQuickPick(resp.value);
            await peer.call("SwitchView", [id, sel]);
            // treeView._onDidChangeTreeData.fire([id])
        }),
        vscode.commands.registerCommand('manifold.add', async (id: string) => {
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
            // treeView._onDidChangeTreeData.fire([id]);
        }),
        vscode.commands.registerCommand('manifold.addComponent', async (id: any) => {
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
            // treeView._onDidChangeTreeData.fire([id]);
        }),
        vscode.commands.registerCommand('manifold.editField', async (id: string) => {
            const resp = await peer.call("PreEditField", [id]);
            if (!resp.value["OK"]) {
                return;
            }
            const newval = await vscode.window.showInputBox({title: resp.value["Name"] as string, value: resp.value["Value"]});
            await peer.call("EditField", [id, newval]);
            // treeView._onDidChangeTreeData.fire([resp.value["ParentID"]]);
        }),
        vscode.commands.registerCommand('manifold.rename', async (id: string) => {
            const resp = await peer.call("GetTreeItem", [id]);
            if (!resp.value) {
                return;
            }
            const newval = await vscode.window.showInputBox({title: "Rename", value: resp.value["label"]});
            await peer.call("Rename", [id, newval]);
            // treeView._onDidChangeTreeData.fire([id]);
        }),
        vscode.commands.registerCommand('manifold.call', async (id: string) => {
            await peer.call("Call", [id]);
            // treeView._onDidChangeTreeData.fire(undefined); 
        }),
        vscode.commands.registerCommand('manifold.delete', async (id: string) => {
            await peer.call("Delete", [id]);
            // treeView._onDidChangeTreeData.fire(undefined); 
        }),
        vscode.commands.registerCommand('manifold.implementFor', async (iface: string, fieldID?: string) => {
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
        }),
        vscode.commands.registerCommand(
            'manifold.toggleComponents', async (id: string) => {
                await peer.call("ToggleComponents", [id]);
                // treeView._onDidChangeTreeData.fire([id]); 
            }
        ),
        vscode.commands.registerCommand(
            'manifold.signal', async (id: string) => {
                const signal = await vscode.window.showInputBox({title: "Signal"});
                await peer.call("Signal", [id, signal]);
            }
        ),
        vscode.commands.registerCommand(
            'manifold.dump', async (id: string) => {
                await peer.call("Dump", [id]);
            }
        ),
        vscode.commands.registerCommand(
            'manifold.reload', async (id: string) => {
            //   treeView._onDidChangeTreeData.fire([id]);
            }
        ),
        vscode.commands.registerCommand(
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
            //   treeView._onDidChangeTreeData.fire([id]); 
            }
        ),
        vscode.commands.registerCommand(
            'manifold.deleteFromInspector', (ctx) => {
            vscode.commands.executeCommand('manifold.delete', ctx.id)
        }),
        vscode.commands.registerCommand(
          'manifold.callMethodFromInspector', (ctx) => {
            vscode.commands.executeCommand('manifold.callMethod', ctx.id)
        }),
        vscode.commands.registerCommand(
          'manifold.viewSourceFromInspector', (ctx) => {
            vscode.commands.executeCommand('manifold.viewSource', ctx.id)
        }),
        vscode.commands.registerCommand(
          'manifold.implementFromInspector', (ctx) => {
            vscode.commands.executeCommand('manifold.implementFor', ctx.iface, ctx.fieldID)
        }),
    );
}


async function openComponentSymbol(symbolPath: string): Promise<Boolean> {
    const symbols = await vscode.commands.executeCommand<vscode.SymbolInformation[]>('vscode.executeWorkspaceSymbolProvider', symbolPath);
    console.log("open: symbols", symbols);
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
    console.log("filepath: symbols", symbols);
    if (symbols.length === 0) {
      return "";
    }
    
    // Assuming the first symbol is the desired one. 22 is kind "Struct"
    let symbol = symbols.filter(s => s.kind === 22)[0];
  
    return symbol.location.uri.fsPath;
}