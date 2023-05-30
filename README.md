# codemaker-cli

CodeMaker CLI

## Overview

CodeMaker AI offers tools and automation for software developers for writing, testing, and documenting source code.

## Features

Fallowing operations are supported:

* Context-aware source code generation.
* Generating unit tests.
* Migrating the programming language syntax.
* Generating source code documentation.
* Refactoring and renaming local variables.

## Supported languages

Following programming languages are supported:

* JavaScript
* Java
* Go
* Kotlin
  
More language support is coming soon.

## Installation

### On MacOS

MacOS users can simply install the latest release of the CLI using [Homebrew Tap](https://github.com/codemakerai/homebrew-tap) by running:

```bash
brew install codemakerai/tap/codemaker-cli
```

### On Linux

1. Download the [latest CLI release](https://github.com/codemakerai/codemaker-cli/releases) and unzip it.
2. Add the CLI to your PATH.

```bash
export PATH=$PATH:/bin
```

### On Windows

1. Download the [latest CLI release](https://github.com/codemakerai/codemaker-cli/releases) and unzip it.
2. Add the CLI to your PATH.

```bash
export PATH=%PATH%;/bin
```

### Using Golang

1. Install the package by running:

```bash
go install github.com/codemakerai/codemaker-cli
```

## Getting started

1. Sign up for the Early Access Program at https://codemaker.ai.
2. Receive the Early Access Program invitation email. 
3. [Install CLI](#installation).
4. Configure the CLI and provide the API Key.

```bash
$ codemaker configure
```

5. Run it.

```bash
$ codemaker generate docs **/*.java
```

# License

MIT License