package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/ehrlich-b/cube/internal/cube"
)

type SolveRequest struct {
	Scramble  string `json:"scramble"`
	Algorithm string `json:"algorithm"`
	Dimension int    `json:"dimension"`
}

type SolveResponse struct {
	Solution string `json:"solution"`
	Steps    int    `json:"steps"`
	Time     string `json:"time"`
}

type ExecRequest struct {
	Command string `json:"command"`
}

type ExecResponse struct {
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
	ExitCode int    `json:"exit_code"`
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Cube Solver</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        .container { background: #f5f5f5; padding: 20px; border-radius: 8px; }
        input, select, button { padding: 10px; margin: 5px; }
        button { background: #007cba; color: white; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #005a8b; }
        .result { background: white; padding: 15px; margin-top: 20px; border-radius: 4px; }
    </style>
</head>
<body>
    <h1>🧩 Cube Solver</h1>
    <div class="container">
        <h2>Solve Your Cube</h2>
        <form id="solveForm">
            <div>
                <label>Scramble:</label><br>
                <input type="text" id="scramble" placeholder="R U R' U' F R F'" style="width: 300px;">
            </div>
            <div>
                <label>Algorithm:</label>
                <select id="algorithm">
                    <option value="beginner">Beginner</option>
                    <option value="cfop">CFOP</option>
                    <option value="kociemba">Kociemba</option>
                </select>
            </div>
            <div>
                <label>Dimension:</label>
                <select id="dimension">
                    <option value="2">2x2x2</option>
                    <option value="3" selected>3x3x3</option>
                    <option value="4">4x4x4</option>
                </select>
            </div>
            <button type="submit">Solve</button>
        </form>
        <div id="result" class="result" style="display: none;"></div>
    </div>

    <script>
        document.getElementById('solveForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const scramble = document.getElementById('scramble').value;
            const algorithm = document.getElementById('algorithm').value;
            const dimension = parseInt(document.getElementById('dimension').value);

            try {
                const response = await fetch('/api/solve', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ scramble, algorithm, dimension })
                });

                const result = await response.json();
                document.getElementById('result').innerHTML =
                    '<h3>Solution:</h3><p>' + result.solution + '</p>' +
                    '<p><strong>Steps:</strong> ' + result.steps + '</p>' +
                    '<p><strong>Time:</strong> ' + result.time + '</p>';
                document.getElementById('result').style.display = 'block';
            } catch (error) {
                document.getElementById('result').innerHTML = '<p style="color: red;">Error: ' + error.message + '</p>';
                document.getElementById('result').style.display = 'block';
            }
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func (s *Server) handleSolve(w http.ResponseWriter, r *http.Request) {
	var req SolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Import cube package at top of file
	c := cube.NewCube(req.Dimension)
	moves, err := cube.ParseScramble(req.Scramble)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing scramble: %v", err), http.StatusBadRequest)
		return
	}

	c.ApplyMoves(moves)

	solver, err := cube.GetSolver(req.Algorithm)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting solver: %v", err), http.StatusBadRequest)
		return
	}

	result, err := solver.Solve(c)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error solving cube: %v", err), http.StatusInternalServerError)
		return
	}

	// Format solution
	var solutionParts []string
	for _, move := range result.Solution {
		solutionParts = append(solutionParts, move.String())
	}

	response := SolveResponse{
		Solution: strings.Join(solutionParts, " "),
		Steps:    result.Steps,
		Time:     result.Duration.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleTerminal(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Cube Solver Terminal</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body {
            margin: 0;
            padding: 0;
            background: #1e1e1e;
            color: #d4d4d4;
            font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
            font-size: 14px;
            line-height: 1.4;
            overflow: hidden;
        }

        .terminal {
            height: 100vh;
            display: flex;
            flex-direction: column;
            padding: 20px;
            box-sizing: border-box;
        }

        .header {
            color: #569cd6;
            margin-bottom: 20px;
            font-weight: bold;
        }

        .output {
            flex: 1;
            overflow-y: auto;
            white-space: pre-wrap;
            font-family: monospace;
            padding-bottom: 10px;
        }

        .input-line {
            display: flex;
            align-items: center;
            margin-top: 10px;
        }

        .prompt {
            color: #4ec9b0;
            margin-right: 8px;
            user-select: none;
        }

        .input {
            flex: 1;
            background: transparent;
            border: none;
            color: #d4d4d4;
            font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
            font-size: 14px;
            outline: none;
        }

        .cursor {
            background: #d4d4d4;
            width: 8px;
            height: 16px;
            display: inline-block;
            animation: blink 1s infinite;
        }

        @keyframes blink {
            0%, 50% { opacity: 1; }
            51%, 100% { opacity: 0; }
        }

        .command-output {
            margin: 5px 0;
        }

        .error {
            color: #f44747;
        }

        .success {
            color: #4ec9b0;
        }

        /* ANSI color support */
        .ansi-white { color: #d4d4d4; }
        .ansi-yellow { color: #dcdcaa; }
        .ansi-red { color: #f44747; }
        .ansi-magenta { color: #c586c0; }
        .ansi-blue { color: #569cd6; }
        .ansi-green { color: #4ec9b0; }
    </style>
</head>
<body>
    <div class="terminal">
        <div class="header">🧩 Cube Solver Terminal - Type 'help' for available commands</div>
        <div class="output" id="output"></div>
        <div class="input-line">
            <span class="prompt">cube$</span>
            <input type="text" class="input" id="commandInput" autocomplete="off" spellcheck="false">
            <span class="cursor" id="cursor"></span>
        </div>
    </div>

    <script>
        const output = document.getElementById('output');
        const input = document.getElementById('commandInput');
        const cursor = document.getElementById('cursor');
        let history = [];
        let historyIndex = -1;

        // Focus input on page load
        input.focus();

        // Keep input focused
        document.addEventListener('click', () => input.focus());

        // Command history
        input.addEventListener('keydown', (e) => {
            if (e.key === 'ArrowUp') {
                e.preventDefault();
                if (historyIndex < history.length - 1) {
                    historyIndex++;
                    input.value = history[history.length - 1 - historyIndex];
                }
            } else if (e.key === 'ArrowDown') {
                e.preventDefault();
                if (historyIndex > 0) {
                    historyIndex--;
                    input.value = history[history.length - 1 - historyIndex];
                } else if (historyIndex === 0) {
                    historyIndex = -1;
                    input.value = '';
                }
            } else if (e.key === 'Enter') {
                e.preventDefault();
                executeCommand();
            }
        });

        function addToOutput(text, className = '') {
            const div = document.createElement('div');
            div.className = 'command-output ' + className;

            // Convert ANSI escape codes to HTML classes
            text = text.replace(/\u001b\[37m(.)\u001b\[0m/g, '<span class="ansi-white">$1</span>');
            text = text.replace(/\u001b\[33m(.)\u001b\[0m/g, '<span class="ansi-yellow">$1</span>');
            text = text.replace(/\u001b\[31m(.)\u001b\[0m/g, '<span class="ansi-red">$1</span>');
            text = text.replace(/\u001b\[35m(.)\u001b\[0m/g, '<span class="ansi-magenta">$1</span>');
            text = text.replace(/\u001b\[34m(.)\u001b\[0m/g, '<span class="ansi-blue">$1</span>');
            text = text.replace(/\u001b\[32m(.)\u001b\[0m/g, '<span class="ansi-green">$1</span>');

            div.innerHTML = text;
            output.appendChild(div);
            output.scrollTop = output.scrollHeight;
        }

        async function executeCommand() {
            const command = input.value.trim();
            if (!command) return;

            // Add to history
            history.push(command);
            historyIndex = -1;

            // Show command in output
            addToOutput('cube$ ' + command, 'success');

            // Clear input
            input.value = '';

            try {
                const response = await fetch('/api/exec', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ command: command })
                });

                const result = await response.json();

                if (result.output) {
                    addToOutput(result.output);
                }

                if (result.error) {
                    addToOutput(result.error, 'error');
                }

            } catch (error) {
                addToOutput('Error: ' + error.message, 'error');
            }
        }

        // Welcome message
        addToOutput('Welcome to Cube Solver Terminal!');
        addToOutput('Try commands like:');
        addToOutput('  twist "R U R\' U\'" --color');
        addToOutput('  solve "R U R\' U\'" --color');
        addToOutput('  verify "R U R\' U\'" "U R U\' R\'"');
        addToOutput('  show "R U R\' U\'" --highlight-oll --color');
        addToOutput('  lookup sune --preview');
        addToOutput('  help');
        addToOutput('');
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func (s *Server) handleExec(w http.ResponseWriter, r *http.Request) {
	var req ExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	command := strings.TrimSpace(req.Command)
	if command == "" {
		response := ExecResponse{
			Output:   "",
			ExitCode: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse command into parts, handling quoted strings properly
	parts, parseErr := parseCommand(command)
	if parseErr != nil {
		response := ExecResponse{
			Error:    fmt.Sprintf("Error parsing command: %v", parseErr),
			ExitCode: 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	if len(parts) == 0 {
		response := ExecResponse{
			Output:   "",
			ExitCode: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get the cube binary path
	cubePath := "./dist/cube"
	if _, err := os.Stat(cubePath); os.IsNotExist(err) {
		// Try alternative paths
		if _, err := os.Stat("cube"); err == nil {
			cubePath = "cube"
		} else {
			response := ExecResponse{
				Error:    "Cube binary not found. Please run 'make build' first.",
				ExitCode: 1,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Execute cube command
	cmd := exec.Command(cubePath, parts...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	response := ExecResponse{
		Output:   stdout.String(),
		ExitCode: exitCode,
	}

	if stderr.Len() > 0 {
		response.Error = stderr.String()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// parseCommand parses a command string into arguments, properly handling quoted strings
func parseCommand(command string) ([]string, error) {
	// Use regex to split on whitespace but preserve quoted strings
	re := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)'`)
	matches := re.FindAllStringSubmatch(command, -1)

	var args []string
	for _, match := range matches {
		if match[1] != "" {
			// Double quoted string
			args = append(args, match[1])
		} else if match[2] != "" {
			// Single quoted string
			args = append(args, match[2])
		} else {
			// Unquoted string
			args = append(args, match[0])
		}
	}

	return args, nil
}
