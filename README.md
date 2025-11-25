# Gorviz - GORM Visualization Tool

## Project Title
Gorviz

## Description
Gorviz is a command-line interface (CLI) tool designed to help Go developers visualize their GORM (Go Object Relational Mapper) models and their relationships. It parses Go source files to extract GORM model definitions and then generates an interactive HTML ERD (Entity-Relationship Diagram) for better understanding of the database schema.

## Features
*   **GORM Model Parsing:** Scans Go project directories to identify and parse GORM model structs and their associated fields, types, and GORM tags.
*   **Relationship Extraction:** Automatically detects various GORM relationships (e.g., `HasOne`, `HasMany`, `BelongsTo`, `Many2Many`, and `Embedded` structs).
*   **YAML Schema Generation:** Generates an intermediate `schema.yaml` file containing a structured representation of the extracted GORM models and their relationships.
*   **Interactive HTML ERD:** Compiles the YAML schema into a single, static HTML file (`gorviz.html`) that provides:
    *   A browsable list of all detected models with their fields and tags.
    *   An interactive Entity-Relationship Diagram (ERD) powered by `vis-network.js` to visually represent model connections.
    *   Dynamic filtering and model detailing within the HTML view.

## Installation
To install the `gorviz` tool, make sure you have Go installed (version 1.25.3 or later, as per `go.mod`). Then run:
```bash
go install gorviz
```
This will install the executable in your `$GOPATH/bin` directory, which should be in your system's PATH.

## Usage
The tool operates in two main steps: `init` to parse your Go code and generate a schema, and `compile` to create the HTML visualization.

1.  **Generate `schema.yaml`:**
    Navigate to the root of your Go project (or specify the path) and run the `init` command, providing the path to your GORM models.
    ```bash
    gorviz init ./path/to/your/models
    ```
    Replace `./path/to/your/models` with the actual directory containing your GORM model definitions. This will create a `schema.yaml` file in the current working directory.

2.  **Compile to HTML:**
    Once `schema.yaml` is generated, run the `compile` command to produce the interactive HTML visualization.
    ```bash
    gorviz compile
    ```
    This will generate a `gorviz.html` file in the current working directory. Open this file in your web browser to view your GORM schema ERD.

## Contributing
We welcome contributions! Please feel free to fork the repository, create a feature branch, commit your changes, and open a pull request.

## License
(License information was not available in the provided files. Please add appropriate license information here if applicable.)