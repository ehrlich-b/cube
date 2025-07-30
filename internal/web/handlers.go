package web

import (
	"encoding/json"
	"fmt"
	"net/http"
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