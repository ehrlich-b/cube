# 🌐 Web Interface Guide

Access the full power of the cube solver through your web browser with a terminal-style interface and REST API.

## 🚀 Quick Start

```bash
# Start the web server
./dist/cube serve

# Custom port and host
./dist/cube serve --port 3000 --host 0.0.0.0
```

Then open your browser to:
- **http://localhost:8080** (default)
- **http://localhost:3000** (custom port)
- **http://0.0.0.0:8080** (all interfaces)

## 🖥️ Terminal Web Interface

### Features

The web terminal interface provides the **exact same functionality** as the CLI:
- Full command support (`solve`, `verify`, `show`, `lookup`, `optimize`, `find`)
- Colored output with emoji blocks 🟦🟨🟩🟧🟥⬜
- Advanced notation support (M/E/S slices, wide moves, layer moves)
- Real-time algorithm discovery and optimization
- Algorithm database lookup with previews

### Basic Usage

Navigate to `http://localhost:8080/terminal` and try these commands:

```bash
# Basic solving
solve "R U R' U'" --color

# Different algorithms
solve "R U R' U'" --algorithm beginner --color
solve "R U R' U'" --algorithm cfop --color
solve "R U R' U'" --algorithm kociemba --color

# Advanced notation
solve "M E S" --dimension 3 --color
solve "Rw Fw Uw" --dimension 4 --color
solve "2R 3L 2F" --dimension 5 --color
```

### Advanced Web Commands

```bash
# Solution verification
verify "R U R' U'" "U R U' R'" --verbose --color

# Pattern highlighting
show "R U R' U'" --highlight-oll --color
show "R U R' F' R U R' U' R' F R2 U' R'" --highlight-pll --color

# Algorithm lookup
lookup sune --preview
lookup --category OLL
lookup --pattern "R U R' U'"

# Move optimization
optimize "R R R"
optimize "R U R' U' R U R' U'"

# Algorithm discovery
find pattern solved --max-moves 4
find sequence "R U R'" --max-moves 6
```

### Web Terminal Features

- **Command History**: Use ↑/↓ arrows to navigate previous commands
- **Tab Completion**: Auto-complete commands and flags (if supported by browser)
- **Copy/Paste**: Full clipboard support for sharing results
- **Responsive Design**: Works on desktop, tablet, and mobile
- **Real-time Output**: See results instantly as they're computed

## 🔌 REST API

### Starting the API Server

```bash
# Start server with API access
./dist/cube serve --port 8080

# API is available at:
# http://localhost:8080/api/exec
```

### API Endpoint

**POST** `/api/exec`

Execute any cube solver command via JSON API.

#### Request Format

```json
{
  "command": "solve \"R U R' U'\" --color"
}
```

#### Response Format

```json
{
  "output": "Solving 3x3x3 cube...\nUsing algorithm: beginner\nSolution: U R U' R'\nSolved in 4 steps (took 1.2ms)\n",
  "success": true
}
```

### API Examples

#### Basic Solving

```bash
curl -X POST http://localhost:8080/api/exec \
  -H "Content-Type: application/json" \
  -d '{"command": "solve \"R U R'\'\" U'\'\" --color"}'
```

#### Algorithm Comparison

```bash
# Test different algorithms via API
curl -X POST http://localhost:8080/api/exec \
  -H "Content-Type: application/json" \
  -d '{"command": "solve \"R U R'\'\" U'\'\" --algorithm beginner"}'

curl -X POST http://localhost:8080/api/exec \
  -H "Content-Type: application/json" \
  -d '{"command": "solve \"R U R'\'\" U'\'\" --algorithm cfop"}'
```

#### Advanced Features via API

```bash
# Algorithm lookup
curl -X POST http://localhost:8080/api/exec \
  -H "Content-Type: application/json" \
  -d '{"command": "lookup sune --preview"}'

# Move optimization
curl -X POST http://localhost:8080/api/exec \
  -H "Content-Type: application/json" \
  -d '{"command": "optimize \"R R R\""}'

# Algorithm discovery
curl -X POST http://localhost:8080/api/exec \
  -H "Content-Type: application/json" \
  -d '{"command": "find pattern solved --max-moves 4"}'
```

#### Solution Verification

```bash
curl -X POST http://localhost:8080/api/exec \
  -H "Content-Type: application/json" \
  -d '{"command": "verify \"R U R'\'\" U'\'\" \"U R U'\'\" R'\'\" --verbose"}'
```

## 🛠️ Integration Examples

### JavaScript/Node.js Integration

```javascript
async function solveCube(scramble, algorithm = 'beginner') {
    const response = await fetch('http://localhost:8080/api/exec', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            command: `solve "${scramble}" --algorithm ${algorithm} --color`
        })
    });
    
    const result = await response.json();
    return result.output;
}

// Usage
solveCube("R U R' U'", "cfop").then(console.log);
solveCube("M E S", "beginner").then(console.log);
```

### Python Integration

```python
import requests
import json

def solve_cube(scramble, algorithm="beginner", dimension=3):
    url = "http://localhost:8080/api/exec"
    command = f'solve "{scramble}" --algorithm {algorithm} --dimension {dimension} --color'
    
    response = requests.post(url, json={"command": command})
    return response.json()["output"]

# Usage
print(solve_cube("R U R' U'", "cfop"))
print(solve_cube("Rw Fw Uw", "beginner", 4))
```

### Bash/Shell Integration

```bash
#!/bin/bash

# Function to solve via API
solve_api() {
    local scramble="$1"
    local algorithm="${2:-beginner}"
    
    curl -s -X POST http://localhost:8080/api/exec \
        -H "Content-Type: application/json" \
        -d "{\"command\": \"solve \\\"$scramble\\\" --algorithm $algorithm --color\"}" \
        | jq -r '.output'
}

# Usage
solve_api "R U R' U'" "cfop"
solve_api "M E S" "beginner"
```

## 🎯 Web-Specific Workflows

### Interactive Algorithm Learning

1. **Start the web server**:
   ```bash
   ./dist/cube serve --port 8080
   ```

2. **Open terminal interface**: Navigate to `http://localhost:8080/terminal`

3. **Learn algorithms interactively**:
   ```bash
   # Look up an algorithm
   lookup sune --preview
   
   # Apply it to see the result
   solve "R U R' U R U2 R'" --color
   
   # Try it on different cube sizes
   solve "R U R' U R U2 R'" --dimension 4 --color
   ```

### API-Based Cube Solving Service

Create your own cube solving service:

```javascript
// Express.js example
const express = require('express');
const axios = require('axios');
const app = express();

app.use(express.json());

app.post('/solve', async (req, res) => {
    const { scramble, algorithm = 'beginner', dimension = 3 } = req.body;
    
    try {
        const response = await axios.post('http://localhost:8080/api/exec', {
            command: `solve "${scramble}" --algorithm ${algorithm} --dimension ${dimension}`
        });
        
        res.json({
            scramble,
            solution: response.data.output,
            algorithm,
            dimension
        });
    } catch (error) {
        res.status(500).json({ error: 'Solving failed' });
    }
});

app.listen(3000, () => {
    console.log('Cube API service running on port 3000');
});
```

### Batch Processing via API

```bash
#!/bin/bash

# Process multiple scrambles
scrambles=(
    "R U R' U'"
    "R U R' F' R U R' U' R' F R2 U' R'"
    "M E S"
    "Rw Fw Uw"
)

for scramble in "${scrambles[@]}"; do
    echo "Solving: $scramble"
    curl -s -X POST http://localhost:8080/api/exec \
        -H "Content-Type: application/json" \
        -d "{\"command\": \"solve \\\"$scramble\\\" --color\"}" \
        | jq -r '.output'
    echo "---"
done
```

## 🔧 Server Configuration

### Command Line Options

```bash
# Basic server start
./dist/cube serve

# Custom port
./dist/cube serve --port 3000

# Custom host (bind to all interfaces)
./dist/cube serve --host 0.0.0.0

# Custom host and port
./dist/cube serve --host 127.0.0.1 --port 8080

# Development mode (more verbose logging)
./dist/cube serve --port 8080 --verbose
```

### Environment Variables

```bash
# Set default port via environment
export CUBE_PORT=3000
./dist/cube serve

# Set default host via environment  
export CUBE_HOST=0.0.0.0
./dist/cube serve
```

### Production Deployment

```bash
# Run in background
nohup ./dist/cube serve --port 8080 --host 0.0.0.0 > cube.log 2>&1 &

# Or use systemd, Docker, etc.
# Example systemd service:
[Unit]
Description=Cube Solver Web Service
After=network.target

[Service]
Type=simple
User=cube
WorkingDirectory=/opt/cube
ExecStart=/opt/cube/dist/cube serve --port 8080 --host 0.0.0.0
Restart=always

[Install]
WantedBy=multi-user.target
```

## 🎯 Advanced Web Features

### Real-time Algorithm Discovery

Use the web interface for interactive algorithm discovery:

```bash
# In the web terminal
find pattern solved --max-moves 4 --steps

# Watch as the solver explores different move sequences
# Results appear in real-time as they're discovered
```

### Collaborative Cube Solving

Share your web terminal session:

1. Start server on public interface:
   ```bash
   ./dist/cube serve --host 0.0.0.0 --port 8080
   ```

2. Share your IP and port with others
3. Multiple people can access the same terminal interface
4. Great for teaching, collaboration, and competitions

### Mobile-Friendly Interface

The web terminal works great on mobile devices:
- Responsive design adapts to screen size
- Touch-friendly interface
- Virtual keyboard support
- Swipe gestures for command history

## 🚨 Troubleshooting

### Common Issues

#### Port Already in Use
```bash
# Error: bind: address already in use
# Solution: Use a different port
./dist/cube serve --port 8081
```

#### Connection Refused
```bash
# Can't connect to http://localhost:8080
# Check if server is running:
curl http://localhost:8080/api/exec
```

#### API Errors
```bash
# Test API connectivity
curl -X POST http://localhost:8080/api/exec \
  -H "Content-Type: application/json" \
  -d '{"command": "solve \"R\""}'
```

### Debug Mode

```bash
# Start server with verbose logging
./dist/cube serve --port 8080 --verbose

# Check server logs
tail -f cube.log
```

## 🎯 Next Steps

Ready to master web-based cube solving?

1. **Try the examples**: Test all the commands shown in this guide
2. **Build integrations**: Create your own applications using the API
3. **Explore mobile usage**: Use the web terminal on your phone/tablet
4. **Share with others**: Collaborate on cube solving challenges

**The web interface opens up endless possibilities for cube solving applications!** 🌐✨