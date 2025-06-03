## Zed CLI For Windows (Unofficial)

This project is an unofficial Windows CLI for Zed (built from source). It provides all basic features:

- Launching Zed (`zed`),
- Opening directories or projects with `zed <path>`
- Adding an 'Open with Zed' context menu integration.

## Table of Content

- [Usage](#usage)
- [Features & Behavior](#features--behavior)
  - [Auto-Directory Creation](#auto-directory-creation)
  - [Single Instance Limitation](#single-instance-limitation)
- [Installation](#installation)
  - [Native Installation Scripts](#native-installation-scripts)
  - [Scoop](#scoop)
  - [Chocolatey (Coming Soon)](#chocolatey-coming-soon)
  - [Manual (GitHub Release)](#manual-github-release)
- [Disclaimer & Affiliation](#disclaimer--affiliation)
- [License](#license)

## Usage

| Command                 | Description                          | Example                           |
| ----------------------- | ------------------------------------ | --------------------------------- |
| `zed`                   | Open Zed with last project           | `zed`                             |
| `zed <path>`            | Open specific file or directory      | `zed C:\projects\my-app`          |
| `zed .`                 | Open current directory               | `zed .`                           |
| `zed config get`        | Get current Zed executable path      | `zed config get`                  |
| `zed config set <path>` | Set Zed executable path              | `zed config set "C:\Zed\zed.exe"` |
| `zed context install`   | Install "Open with Zed" context menu | `zed context install`             |
| `zed context uninstall` | Remove "Open with Zed" context menu  | `zed context uninstall`           |

> [!NOTE]
> Use `zed context install` to add "Open with Zed" to your Windows context menu for easy right-click access. By default, it's not installed.

## Features & Behavior

### Auto-Directory Creation

When opening a non-existent path, the CLI automatically creates the required directories before launching Zed:

```bash
zed D:\projects\monkeypress
```

In this example, if `monkeypress` doesn't exist but `D:\projects\` does, the CLI will:

1. Create the `monkeypress` directory under `D:\projects\`
2. Open the newly created directory in Zed

![A terminal-like window with a dark background shows a command and its output. The command entered is `zed D:\projects\monkeypress`. Below it are three lines of output](./public/auto-directory.png)

### Single Instance Limitation

**Important:** This CLI cannot launch multiple Zed instances when one is already running. This limitation exists because:

- The CLI is not first-party (unofficial)
- No access to Zed's internal IPC (Inter-Process Communication) system
- IPC typically handles multi-instance management in editors

When attempting to open a project while Zed is already running, the CLI will notify you of this limitation.

![
A pop-up window with a dark background displays two lines of text. The first line has a red 'X' icon followed by "Zed is already running in another instance!!". The second line has a yellow triangle icon followed by "This CLI cannot launch a second instance due to Zed's limitation". The words "in" and "limitation" are highlighted in green. The overall design features a subtle geometric grid pattern in the background.](./public/mutliple-instance-running.png)

## Installation

Recommended installation methods in order of preference:

1. [Native Installation Scripts](#native-installation-scripts)
2. [Scoop](#scoop)
3. [Chocolatey (Coming Soon)](#chocolatey-coming-soon)
4. [Manual (GitHub Release)](#manual-github-release)

### Native Installation Scripts

Running the native PowerShell scripts will handle everything related to setting up environment variables on your system.

#### Install CLI

Download and run the installation script for the Unofficial Zed CLI:

```powershell
irm https://raw.githubusercontent.com/SameerJS6/zed-cli-win-unofficial/blob/main/scripts/release/install-wrapper.ps1 | iex
```

#### Install Zed + CLI (All-in-One)

Install both Zed (Unofficial Build) and the Unofficial CLI with zero setup. This script handles everything automatically:

```powershell
irm https://raw.githubusercontent.com/SameerJS6/zed-cli-win-unofficial/blob/main/scripts/release/install-with-zed-wrapper.ps1 | iex
```

> [!NOTE]
> Installing via this method will place Zed and the Unofficial Zed CLI in their default directories.
>
> #### Zed
>
> ```powershell
> %LOCALAPPDATA%\Programs\Zed
> ```
>
> #### Unofficial Zed CLI
>
> ```powershell
> %LOCALAPPDATA%\zed-cli-win-unofficial
> ```

### Scoop

Install using [Scoop](https://scoop.sh/) for easy updates and management:

```powershell
scoop bucket add zed-cli-unofficial https://github.com/SameerJS6/zed-cli-win-unofficial
scoop install zed-cli-unofficial/zed-cli-win-unofficial
```

> [!TIP]
> If you don't have **Scoop** installed, run the following commands in **PowerShell** to install it:
>
> ```powershell
> Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
> Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression
> ```

### Chocolatey (Coming Soon)

> A Chocolatey package for zed-cli-win-unofficial is coming soon.

### Manual (GitHub Release)

Download and install manually from the GitHub releases page:

1. Visit the [Releases page](https://github.com/SameerJS6/zed-cli-win-unofficial/releases).
2. Download the `windows-x86_64.zip` asset.
3. Extract to a folder of your choosing (default: `%LOCALAPPDATA%\\zed-cli-win-unofficial`).
4. Update your user `PATH` to include that folder (choose one method below):

<details><summary>PowerShell (CLI)</summary>

```powershell
$path = "$env:LOCALAPPDATA\\zed-cli-win-unofficial"
[Environment]::SetEnvironmentVariable('PATH', $env:PATH + ';' + $path, 'User')
```

</details>

<details><summary>GUI</summary>

- Press Win, type "Environment Variables", and open "Edit user environment variables".
- Under "User variables", select "Path" → click "Edit" → click "New".
- Paste `%LOCALAPPDATA%\\zed-cli-win-unofficial` and click "OK" on all dialogs.

> [!TIP]
> If you have PowerToys installed, you can use the PowerToys _Environment Variables_ tool to manage your variables more easily.

</details>

<details><summary>Common Pitfalls</summary>

- Unblock the downloaded ZIP if prompted (Right-click → Properties → Unblock).
- Verify both `zed-cli-win-unofficial.exe` and `zed.bat` are present.
- Restart your terminal after updating the `PATH`.

</details>

## Disclaimer & Affiliation

This project is an unofficial Windows CLI launcher for [Zed](https://zed.dev). It is not affiliated with or endorsed by the Zed team.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
