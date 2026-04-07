
import * as vscode from 'vscode';
import * as duplex from "@progrium/duplex";
import { manifold } from '../webview/webview.js';

export async function activate(ctx: vscode.ExtensionContext, fsys: any, peer: duplex.Peer, realm: manifold.Realm) {
  const treeView = new TreeView(peer, realm);
  const view = vscode.window.createTreeView('manifold.hierarchy', {
    treeDataProvider: treeView, 
    showCollapseAll: true, 
    canSelectMany: true
  });
  ctx.subscriptions.push(view);

  realm.addEventListener("change", (event) => {
    const ids = (event as CustomEvent<string[]>).detail;
    console.log("change:", ids);
    treeView._onDidChangeTreeData.fire(undefined);
  });

  view.onDidChangeCheckboxState(async (event) => {
    event.items.forEach(async (item) => {
      await peer.call("Toggle", [item[0], item[1] === 1]);
      treeView._onDidChangeTreeData.fire([item[0]]);
    });
  });
  view.onDidChangeSelection(async (event) => {
    // const selected = event.selection[0];
    // if (selected) {
    //   vscode.commands.executeCommand("manifold.inspect", selected);
    // }
  });
}

export class TreeView implements vscode.TreeDataProvider<string> {
    peer: any;
    realm: manifold.Realm;

    public _onDidChangeTreeData: vscode.EventEmitter<(string | undefined)[] | undefined> = new vscode.EventEmitter<string[] | undefined>();
	  public onDidChangeTreeData: vscode.Event<any> = this._onDidChangeTreeData.event;

    constructor(peer: any, realm: manifold.Realm) {
        this.peer = peer;
        this.realm = realm;
    }

    public async getChildren(id?: string|null): Promise<string[]> {
      await this.realm.ready;
      if (!id) {
        return ["@main"];
      }
      const node = this.realm.resolve(id);
      if (!node || !node.children) {
        console.log("no children for:", id);
        return [];
      }
      return node.children.map(n => n.id);
    }

    public async getTreeItem(id: string): Promise<vscode.TreeItem> {
      await this.realm.ready;
      const node = this.realm.resolve(id);
      if (!node) {
        throw new Error("Node not found");
      }

      const item: vscode.TreeItem = {
        label: node.name,
        id: node.id,
        tooltip: node.id,
      };
      const tags = [];
      if (node.isComponent) {
        item.description = node.componentType;
      } else {
        tags.push("obj");
      }
      item.contextValue = tags.join(",");
      if (node.children.length > 0) {
        item.collapsibleState = vscode.TreeItemCollapsibleState.Collapsed;
      }
      item.command = {
        command: 'manifold.inspect',
        arguments: [id],
        title: 'Inspect',
      };
      return item;

        // const resp = await this.peer.call("GetTreeItem", [id]);
        // const element = resp.value;
        // let icon = "";
        // let color = undefined;
        // let attrNode = element as {Attrs: any};
        // if (attrNode.Attrs) {
        //     switch (attrNode.Attrs["view"]) {
        //     case "components":
        //         //icon = "gear";
        //         // if (!attrNode.Attrs["desc"]) {
        //         //   element.description = "(components)";
        //         // }
        //         break;
        //     // case "fields":
        //         // icon = "symbol-object";
        //         // break;
        //     default:
        //         icon = "";
        //         if (attrNode.Attrs["view"] !== "objects" && !element.description) {
        //           element.description = attrNode.Attrs["view"];
        //         }
        //     }
        //     icon = icon || attrNode.Attrs["icon"];
        //     if (attrNode.Attrs["color"]) {
        //         color = new vscode.ThemeColor(attrNode.Attrs["color"]);
        //     }
        //     if (attrNode.Attrs["busy"]) {
        //       icon = "sync~spin";
        //       color = new vscode.ThemeColor("activityBar.inactiveForeground");
        //     }
        // }
        // return Object.assign({}, element, {
        //     tooltip: (element.tooltip) ? new vscode.MarkdownString(element.tooltip as string) : undefined,
        //     iconPath: (icon) ? new vscode.ThemeIcon(icon, color) : undefined
        // });
    }
}

