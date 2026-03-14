import * as vscode from 'vscode';
import * as cp from 'child_process';
import * as path from 'path';

let diagnosticCollection: vscode.DiagnosticCollection;

export function activate(context: vscode.ExtensionContext) {
	diagnosticCollection = vscode.languages.createDiagnosticCollection('nutshell');
	context.subscriptions.push(diagnosticCollection);

	// Register commands
	context.subscriptions.push(
		vscode.commands.registerCommand('nutshell.init', cmdInit),
		vscode.commands.registerCommand('nutshell.pack', cmdPack),
		vscode.commands.registerCommand('nutshell.unpack', cmdUnpack),
		vscode.commands.registerCommand('nutshell.inspect', cmdInspect),
		vscode.commands.registerCommand('nutshell.validate', cmdValidate),
		vscode.commands.registerCommand('nutshell.check', cmdCheck),
		vscode.commands.registerCommand('nutshell.serve', cmdServe),
	);

	// Auto-validate on save
	context.subscriptions.push(
		vscode.workspace.onDidSaveTextDocument((doc) => {
			if (path.basename(doc.uri.fsPath) === 'nutshell.json') {
				const config = vscode.workspace.getConfiguration('nutshell');
				if (config.get<boolean>('autoValidate', true)) {
					validateManifest(doc.uri);
				}
			}
		})
	);

	// Validate all open nutshell.json files on activation
	vscode.workspace.textDocuments.forEach((doc) => {
		if (path.basename(doc.uri.fsPath) === 'nutshell.json') {
			validateManifest(doc.uri);
		}
	});
}

export function deactivate() {
	diagnosticCollection?.dispose();
}

// ── Helpers ──

function getCliPath(): string {
	return vscode.workspace.getConfiguration('nutshell').get<string>('cliPath', 'nutshell');
}

function runCli(args: string[], cwd?: string): Promise<{ stdout: string; stderr: string }> {
	return new Promise((resolve, reject) => {
		const cli = getCliPath();
		const opts: cp.ExecOptions = { maxBuffer: 4 * 1024 * 1024 };
		if (cwd) {
			opts.cwd = cwd;
		}
		cp.execFile(cli, args, opts, (err, stdout, stderr) => {
			if (err) {
				reject(new Error(stderr || err.message));
			} else {
				resolve({ stdout, stderr });
			}
		});
	});
}

async function pickWorkspaceDir(): Promise<string | undefined> {
	const folders = vscode.workspace.workspaceFolders;
	if (!folders) {
		vscode.window.showErrorMessage('No workspace folder open');
		return undefined;
	}
	if (folders.length === 1) {
		return folders[0].uri.fsPath;
	}
	const picked = await vscode.window.showWorkspaceFolderPick({ placeHolder: 'Select workspace folder' });
	return picked?.uri.fsPath;
}

async function pickNutFile(): Promise<string | undefined> {
	const uris = await vscode.window.showOpenDialog({
		canSelectFiles: true,
		canSelectFolders: false,
		canSelectMany: false,
		filters: { 'Nutshell Bundles': ['nut'] },
		title: 'Select .nut bundle',
	});
	return uris?.[0]?.fsPath;
}

// ── Commands ──

async function cmdInit() {
	const dir = await pickWorkspaceDir();
	if (!dir) { return; }

	try {
		await runCli(['init', '--dir', dir]);
		vscode.window.showInformationMessage(`Nutshell bundle initialized in ${dir}`);
		// Open the created manifest
		const manifestUri = vscode.Uri.file(path.join(dir, 'nutshell.json'));
		const doc = await vscode.workspace.openTextDocument(manifestUri);
		await vscode.window.showTextDocument(doc);
	} catch (e: unknown) {
		vscode.window.showErrorMessage(`Init failed: ${(e as Error).message}`);
	}
}

async function cmdPack() {
	const dir = await pickWorkspaceDir();
	if (!dir) { return; }

	const outName = await vscode.window.showInputBox({
		prompt: 'Output .nut filename',
		value: 'bundle.nut',
	});
	if (!outName) { return; }

	try {
		const { stdout } = await runCli(['pack', '--dir', dir, '-o', path.join(dir, outName)]);
		vscode.window.showInformationMessage(stdout.trim() || `Packed ${outName}`);
	} catch (e: unknown) {
		vscode.window.showErrorMessage(`Pack failed: ${(e as Error).message}`);
	}
}

async function cmdUnpack() {
	const file = await pickNutFile();
	if (!file) { return; }

	const uris = await vscode.window.showOpenDialog({
		canSelectFiles: false,
		canSelectFolders: true,
		canSelectMany: false,
		title: 'Select output directory',
	});
	if (!uris?.[0]) { return; }

	try {
		const { stdout } = await runCli(['unpack', file, '-o', uris[0].fsPath]);
		vscode.window.showInformationMessage(stdout.trim() || 'Unpacked successfully');
	} catch (e: unknown) {
		vscode.window.showErrorMessage(`Unpack failed: ${(e as Error).message}`);
	}
}

async function cmdInspect() {
	const file = await pickNutFile();
	if (!file) { return; }

	try {
		const { stdout } = await runCli(['inspect', file, '--json']);
		const doc = await vscode.workspace.openTextDocument({
			content: stdout,
			language: 'json',
		});
		await vscode.window.showTextDocument(doc, { preview: true });
	} catch (e: unknown) {
		vscode.window.showErrorMessage(`Inspect failed: ${(e as Error).message}`);
	}
}

async function cmdValidate(uri?: vscode.Uri) {
	let target: string;
	if (uri) {
		target = uri.fsPath;
	} else {
		const dir = await pickWorkspaceDir();
		if (!dir) { return; }
		target = dir;
	}

	try {
		const { stdout } = await runCli(['validate', target, '--json']);
		const result = JSON.parse(stdout);
		if (result.valid) {
			vscode.window.showInformationMessage('✓ Bundle is valid');
		} else {
			const errors: string[] = result.errors || [];
			vscode.window.showWarningMessage(`Bundle has ${errors.length} issue(s): ${errors[0] || ''}`);
		}
	} catch (e: unknown) {
		vscode.window.showErrorMessage(`Validate failed: ${(e as Error).message}`);
	}
}

async function cmdCheck() {
	const dir = await pickWorkspaceDir();
	if (!dir) { return; }

	try {
		const { stdout } = await runCli(['check', '--dir', dir, '--json']);
		const doc = await vscode.workspace.openTextDocument({
			content: stdout,
			language: 'json',
		});
		await vscode.window.showTextDocument(doc, { preview: true });
	} catch (e: unknown) {
		vscode.window.showErrorMessage(`Check failed: ${(e as Error).message}`);
	}
}

async function cmdServe() {
	const dir = await pickWorkspaceDir();
	if (!dir) { return; }

	const port = await vscode.window.showInputBox({
		prompt: 'Port for web viewer',
		value: '8080',
	});
	if (!port) { return; }

	const terminal = vscode.window.createTerminal({ name: 'Nutshell Viewer', cwd: dir });
	terminal.show();
	terminal.sendText(`${getCliPath()} serve "${dir}" --port ${port}`);
	vscode.window.showInformationMessage(`Nutshell web viewer starting on port ${port}`);
}

// ── Diagnostics ──

async function validateManifest(uri: vscode.Uri) {
	const dir = path.dirname(uri.fsPath);
	try {
		const { stdout } = await runCli(['validate', dir, '--json']);
		const result = JSON.parse(stdout);
		const diagnostics: vscode.Diagnostic[] = [];

		if (result.errors) {
			for (const err of result.errors as string[]) {
				const diag = new vscode.Diagnostic(
					new vscode.Range(0, 0, 0, 0),
					err,
					vscode.DiagnosticSeverity.Error,
				);
				diag.source = 'nutshell';
				diagnostics.push(diag);
			}
		}

		if (result.warnings) {
			for (const warn of result.warnings as string[]) {
				const diag = new vscode.Diagnostic(
					new vscode.Range(0, 0, 0, 0),
					warn,
					vscode.DiagnosticSeverity.Warning,
				);
				diag.source = 'nutshell';
				diagnostics.push(diag);
			}
		}

		diagnosticCollection.set(uri, diagnostics);
	} catch {
		// CLI not available or invalid JSON — clear diagnostics silently
		diagnosticCollection.delete(uri);
	}
}
