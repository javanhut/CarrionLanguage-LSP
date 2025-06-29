import * as path from 'path';
import * as vscode from 'vscode';
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
    TransportKind
} from 'vscode-languageclient/node';

let client: LanguageClient;

export function activate(context: vscode.ExtensionContext) {
    const config = vscode.workspace.getConfiguration('carrion');
    const serverPath = config.get<string>('server.path', 'carrion-lsp');
    
    // Server options
    const serverOptions: ServerOptions = {
        run: {
            command: serverPath,
            args: ['--stdio'],
            transport: TransportKind.stdio
        },
        debug: {
            command: serverPath,
            args: ['--stdio', '--log=/tmp/carrion-lsp-debug.log'],
            transport: TransportKind.stdio
        }
    };

    // Client options
    const clientOptions: LanguageClientOptions = {
        documentSelector: [{ scheme: 'file', language: 'carrion' }],
        synchronize: {
            fileEvents: vscode.workspace.createFileSystemWatcher('**/*.crl')
        },
        outputChannelName: 'Carrion Language Server',
        traceOutputChannelName: 'Carrion Language Server Trace'
    };

    // Create and start the language client
    client = new LanguageClient(
        'carrion',
        'Carrion Language Server',
        serverOptions,
        clientOptions
    );

    client.start();

    // Register additional commands if needed
    context.subscriptions.push(
        vscode.commands.registerCommand('carrion.restartServer', () => {
            client.stop().then(() => client.start());
        })
    );
}

export function deactivate(): Thenable<void> | undefined {
    if (!client) {
        return undefined;
    }
    return client.stop();
}