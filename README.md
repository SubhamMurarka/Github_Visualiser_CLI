# GitHub Visualizer CLI

GitHub Visualizer CLI is a powerful command-line tool for tracking and visualizing your Git repositories directly from the terminal. Add directories containing Git repositories and view detailed commit statistics in a clean, visual format.

## Demo

Check out the [demo video](https://github.com/SubhamMurarka/Github_Visualiser_CLI/assets/108292932/341e5cf1-3184-4245-828f-26b26a480526) to see GitHub Visualizer CLI in action!

## Features

- **Add Directories:** Scan and track multiple Git repositories within specified directories.
- **Email Filtering:** View commit statistics filtered by a specific email address.
- **Visual Statistics:** Clean, color-coded display of commit activity over the last six months.

## Tech Stack

![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white&style=for-the-badge)

## Installation

To install GitHub Visualizer CLI, clone the repository and navigate to the project directory:

```bash
git clone https://github.com/SubhamMurarka/Github_Visualiser_CLI.git
cd Github_Visualiser_CLI
```

## Add a directory

This command adds the specified directory, allowing the CLI to scan for Git repositories and track them:

```bash
go run . -add "/path/to/your/folder"
```

## Filter commits by email

This command filters and displays commit statistics for the specified email address:

```bash
go run . -email "your-email@example.com"
```
