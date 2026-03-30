import * as vscode from 'vscode';

export class TempFileSystemProvider implements vscode.FileSystemProvider {

    private readonly buffers = new Map<string, Uint8Array>();

    // Event Emitters
    private readonly onDidChangeFileEmitter = new vscode.EventEmitter<vscode.FileChangeEvent[]>();
    readonly onDidChangeFile: vscode.Event<vscode.FileChangeEvent[]> = this.onDidChangeFileEmitter.event;

    private readonly watchers = new Map<string, vscode.FileSystemWatcher>();

    // Stat implementation
    stat(uri: vscode.Uri): vscode.FileStat {
        return {
            type: vscode.FileType.File,
            ctime: 0,
            mtime: 0,
            size: this.buffers.get(uri.path)?.length ?? 0
        };
    }

    // Read file implementation
    readFile(uri: vscode.Uri): Uint8Array {
        return this.buffers.get(uri.path) ?? new Uint8Array();
    }

    // Write file implementation
    writeFile(uri: vscode.Uri, content: Uint8Array, options: { create: boolean, overwrite: boolean }): void {
        this.buffers.set(uri.path, content);
        this.onDidChangeFileEmitter.fire([{ type: vscode.FileChangeType.Changed, uri }]);
    }

    // Other required methods
    readDirectory(uri: vscode.Uri): [string, vscode.FileType][] { return []; }
    createDirectory(uri: vscode.Uri): void {}
    delete(uri: vscode.Uri): void {
        this.buffers.delete(uri.path);
        this.onDidChangeFileEmitter.fire([{ type: vscode.FileChangeType.Deleted, uri }]);
        this.notifyWatchers([{ type: vscode.FileChangeType.Deleted, uri }]);
    }
    rename(oldUri: vscode.Uri, newUri: vscode.Uri, options: { overwrite: boolean }): void {
        const content = this.readFile(oldUri);
        this.writeFile(newUri, content, { create: true, overwrite: options.overwrite });
        this.delete(oldUri);
        this.onDidChangeFileEmitter.fire([
            { type: vscode.FileChangeType.Deleted, uri: oldUri },
            { type: vscode.FileChangeType.Created, uri: newUri }
        ]);
        this.notifyWatchers([
            { type: vscode.FileChangeType.Deleted, uri: oldUri },
            { type: vscode.FileChangeType.Created, uri: newUri }
        ]);
    }

    // Watch implementation
    watch(uri: vscode.Uri, options: { recursive: boolean, excludes: string[] }): vscode.Disposable {
        const watcher = vscode.workspace.createFileSystemWatcher(uri.fsPath);
        this.watchers.set(uri.path, watcher);

        watcher.onDidChange((uri) => this.onDidChangeFileEmitter.fire([{ type: vscode.FileChangeType.Changed, uri }]));
        watcher.onDidCreate((uri) => this.onDidChangeFileEmitter.fire([{ type: vscode.FileChangeType.Created, uri }]));
        watcher.onDidDelete((uri) => this.onDidChangeFileEmitter.fire([{ type: vscode.FileChangeType.Deleted, uri }]));

        return {
            dispose: () => {
                this.watchers.delete(uri.path);
                watcher.dispose();
            }
        };
    }

    private notifyWatchers(events: vscode.FileChangeEvent[]) {
        for (const [path, watcher] of this.watchers.entries()) {
            for (const event of events) {
                if (event.uri.path === path) {
                    this.onDidChangeFileEmitter.fire(events);
                }
            }
        }
    }
}