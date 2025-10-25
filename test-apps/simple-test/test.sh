#!/bin/bash

# Simple test script to test pack build without ECR push
set -e

echo "Testing pack build with simple Node.js app..."

# Create a simple package.json
cat > package.json << EOF
{
  "name": "simple-test",
  "version": "1.0.0",
  "main": "server.js",
  "scripts": {
    "start": "node server.js"
  },
  "engines": {
    "node": ">=18.0.0"
  }
}
EOF

# Create a simple server.js
cat > server.js << EOF
const http = require('http');
const port = process.env.PORT || 8080;

const server = http.createServer((req, res) => {
  res.writeHead(200, {'Content-Type': 'application/json'});
  res.end(JSON.stringify({
    message: 'Hello from simple test!',
    timestamp: new Date().toISOString()
  }));
});

server.listen(port, '0.0.0.0', () => {
  console.log(\`Server running on port \${port}\`);
});
EOF

echo "Created simple Node.js app"
echo "Testing pack build..."

# Test pack build with local Docker image using a different builder
# Use --publish to avoid containerd storage issues
/Users/farzanbhuiyan/flow-ai/cmd/deployer/pack build simple-test:latest --builder paketobuildpacks/builder:tiny --pull-policy always --verbose --publish

echo "Build successful! Testing the image..."
docker run --rm -p 8080:8080 simple-test:latest &
CONTAINER_PID=$!

# Wait a moment for the container to start
sleep 3

# Test the application
echo "Testing application..."
curl -s http://localhost:8080 | jq .

# Cleanup
kill $CONTAINER_PID 2>/dev/null || true
docker rmi simple-test:latest

echo "Test completed successfully!"
