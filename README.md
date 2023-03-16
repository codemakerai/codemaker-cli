# codemaker-cli

CodeMaker CLI

## Overview

CodeMaker AI offers tools and automation for software developers for writing, testing, and documenting source code.

## Features

Fallowing operations are supported:

* Generating source code documentation.

## Supported languages

Following programming languages are supported:

* Java 17

## Installation

### On MacOS

MacOS users can simply install the latest release of the CLI using [Homebrew Tap](https://github.com/codemakerai/homebrew-tap) by running:

```bash
$ brew install codemakerai/tap/codemaker-cli
```

### On Linux

Download the [latest CLI release](https://github.com/codemakerai/codemaker-cli/releases) and unzip it.

### On Windows

Download the [latest CLI release](https://github.com/codemakerai/codemaker-cli/releases) and unzip it.

## Getting started

1. Sign up for the Early Access Program at https://codemaker.ai.
2. Receive the Early Access Program invitation email. 
3. Install CLI.
4. Add the CLI to your PATH.

```bash
export PATH=$PATH:/bin
```

5. Configure the CLI and provide the API Key.

```bash
$ codemaker configure
```

5. Run it.

```bash
$ codemaker generate docs **/*.java
```

# License

MIT License