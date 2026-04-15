# Codegen written in Go lang

## Lexer and Parser

This implementation contains a handwritten lexer and parser for the FCL language
(see more here [pdf](https://www.researchgate.net/profile/Peter-Sestoft/publication/213882771_Partial_Evaluation_and_Automatic_Program_Generation/links/5704f27108ae13eb88b92a5f/Partial-Evaluation-and-Automatic-Program-Generation.pdf))

## Semantics

The semantics follows those provided here: [link](https://link.springer.com/chapter/10.1007/978-3-642-29709-0_13).

## Requirements

- Go 1.24.2 or higher

## Building

To build all CLI tools, run:

```bash
make all
```

This will create the following binaries in the `bin/` directory:
- `bin/parser` - Parse FCL programs and display the AST
- `bin/cogen` - Code generator for partial evaluation
- `bin/evaluator` - Evaluate FCL programs

To build individual tools:

```bash
make bin/parser    # Build only the parser
make bin/cogen     # Build only the code generator
make bin/evaluator # Build only the evaluator
```

To clean build artifacts:

```bash
make clean
```

## CLI Usage

### Parser

Parse an FCL file and display its abstract syntax tree:

```bash
./bin/parser <inputfile>
```

Example:
```bash
./bin/parser example.fcl
```

### Code Generator (Cogen)

Generate specialized code by providing static parameter indices (delta):

```bash
./bin/cogen <inputfile> [delta...]
```

Example:
```bash
# Generate code treating parameter at index 0 as static
./bin/cogen pow.fcl 0

# Generate code treating parameters at indices 0 and 1 as static
./bin/cogen ackermann.fcl 0 1
```

### Evaluator

Run an FCL program with the given arguments:

```bash
./bin/evaluator <inputfile> [args...]
```

Example:
```bash
# Evaluate pow.fcl with m=2 and n=3
./bin/evaluator pow.fcl 2 3

# Evaluate ackermann.fcl with m=2 and n=3
./bin/evaluator ackermann.fcl 2 3
```

### REPL

Start an interactive REPL session:

```bash
go run ./cmd/repl
```

Or build and run:
```bash
go build -o bin/repl ./cmd/repl && ./bin/repl
```

## Running the Web Interface

The project includes a web interface that provides both code generation and evaluation capabilities through a browser.

### Option 1: Run Locally (Development)

1. Navigate to the web directory:
   ```bash
   cd web
   ```

2. Run the server:
   ```bash
   go run main.go
   ```

3. Open your browser and navigate to:
   ```
   http://localhost:8080
   ```

### Option 2: Build and Run Binary

1. Build the web server binary:
   ```bash
   cd web
   go build -o server .
   ```

2. Run the server:
   ```bash
   ./server
   ```

3. Access the web interface at `http://localhost:8080`

### Option 3: Run with Docker

1. Build the Docker image from the project root:
   ```bash
   docker build -f web/Dockerfile -t cogen-web .
   ```

2. Run the container:
   ```bash
   docker run -p 8080:8080 cogen-web
   ```

3. Access the web interface at `http://localhost:8080`

### Web Interface Features

The web interface provides two modes:

1. **Generator Mode**: Generate specialized FCL code by specifying which parameters should be treated as static (compile-time constants)
   - Enter your FCL program in the text area
   - Specify delta indices (comma-separated) for static parameters
   - Click "Run" to generate specialized code

2. **Evaluator Mode**: Execute FCL programs with specific argument values
   - Enter your FCL program
   - Specify the number of arguments
   - Enter values for each argument
   - Click "Run" to evaluate and see the result

Example programs are available in the dropdown menu (pow, ackermann, turing machine).

## Example FCL Files

The repository includes several example FCL programs:
- `pow.fcl` - Power function (m^n)
- `ackermann.fcl` - Ackermann function
- `turing_machine.fcl` - Turing machine simulation
- `example.fcl` - Basic example program

## Project Structure

```
.
├── ast/          # Abstract Syntax Tree definitions
├── cmd/          # CLI tools (parser, cogen, evaluator, repl)
├── evaluator/    # FCL interpreter/evaluator
├── generator/    # Code generator for partial evaluation
├── lexer/        # Lexical analyzer
├── object/       # Runtime object types
├── parser/       # Parser implementation
├── token/        # Token definitions
├── web/          # Web interface
│   ├── main.go   # Web server
│   └── static/   # Frontend files (HTML, CSS, JS)
└── *.fcl         # Example FCL programs
```

## API Endpoints

When running the web server, the following API endpoints are available:

- `POST /api/generate` - Generate specialized code
  - Request body: `{"program": "...", "delta": [0, 1]}`
  - Response: `{"result": "..."}` or `{"error": "..."}`

- `POST /api/evaluate` - Evaluate a program
  - Request body: `{"program": "...", "args": ["2", "3"]}`
  - Response: `{"result": "..."}` or `{"error": "..."}`

- `GET /` - Web interface
