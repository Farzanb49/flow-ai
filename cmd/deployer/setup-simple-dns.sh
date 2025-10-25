#!/bin/bash

# Simple DNS Setup for Flow Deploy Apps (Immediate Access)
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}Setting up Simple DNS for Flow Deploy Apps...${NC}"

# Configuration
CLUSTER_NAME="zanny-playground"
ACCOUNT_ID="${AWS_ACCOUNT_ID:-YOUR_AWS_ACCOUNT_ID}"
REGION="us-east-1"

echo -e "${YELLOW}Cluster: $CLUSTER_NAME${NC}"
echo -e "${YELLOW}Account ID: $ACCOUNT_ID${NC}"

# Step 1: Get the current LoadBalancer IP
echo -e "${BLUE}Step 1: Getting LoadBalancer information...${NC}"

ALB_DNS=$(kubectl get svc kourier -n kourier-system -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
echo -e "${GREEN}ALB DNS: $ALB_DNS${NC}"

# Get the actual IP address
ALB_IP=$(dig +short $ALB_DNS | head -1)
echo -e "${GREEN}ALB IP: $ALB_IP${NC}"

# Step 2: Create a simple ingress for immediate access
echo -e "${BLUE}Step 2: Creating simple ingress...${NC}"

cat > /tmp/simple-ingress.yaml << EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: flow-apps-simple-ingress
  namespace: default
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
spec:
  rules:
  - host: $ALB_IP
    http:
      paths:
      - path: /nodejs-app
        pathType: Prefix
        backend:
          service:
            name: nodejs-app
            port:
              number: 80
      - path: /python-app
        pathType: Prefix
        backend:
          service:
            name: python-app
            port:
              number: 80
      - path: /go-app
        pathType: Prefix
        backend:
          service:
            name: go-app
            port:
              number: 80
      - path: /
        pathType: Prefix
        backend:
          service:
            name: nodejs-app
            port:
              number: 80
EOF

# Apply the ingress
kubectl apply -f /tmp/simple-ingress.yaml

echo -e "${GREEN}Simple ingress applied!${NC}"

# Step 3: Test the setup
echo -e "${BLUE}Step 3: Testing the setup...${NC}"

echo -e "${YELLOW}Testing Node.js app...${NC}"
curl -s "http://$ALB_IP/nodejs-app/health" | head -1

echo -e "${YELLOW}Testing root path...${NC}"
curl -s "http://$ALB_IP/" | head -1

# Step 4: Create a simple HTML page for easy navigation
echo -e "${BLUE}Step 4: Creating navigation page...${NC}"

cat > /tmp/index.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Flow Deploy Apps</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; }
        .app-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin: 30px 0; }
        .app-card { background: #f8f9fa; padding: 20px; border-radius: 8px; border-left: 4px solid #007bff; }
        .app-card h3 { margin: 0 0 10px 0; color: #007bff; }
        .app-card p { margin: 0 0 15px 0; color: #666; }
        .app-link { display: inline-block; background: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; }
        .app-link:hover { background: #0056b3; }
        .status { text-align: center; margin: 20px 0; }
        .status-item { display: inline-block; margin: 0 20px; }
        .status-value { font-weight: bold; color: #28a745; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Flow Deploy Apps</h1>
        
        <div class="status">
            <div class="status-item">
                <div>Status</div>
                <div class="status-value">✅ Online</div>
            </div>
            <div class="status-item">
                <div>Load Balancer</div>
                <div class="status-value">AWS ALB</div>
            </div>
        </div>
        
        <div class="app-grid">
            <div class="app-card">
                <h3>Node.js App</h3>
                <p>Express.js application with health monitoring</p>
                <a href="/nodejs-app" class="app-link">Open App</a>
                <br><br>
                <a href="/nodejs-app/health" class="app-link" style="background: #28a745;">Health Check</a>
            </div>
            
            <div class="app-card">
                <h3>Python App</h3>
                <p>Flask application with Gunicorn</p>
                <a href="/python-app" class="app-link">Open App</a>
                <br><br>
                <a href="/python-app/health" class="app-link" style="background: #28a745;">Health Check</a>
            </div>
            
            <div class="app-card">
                <h3>Go App</h3>
                <p>Go HTTP server application</p>
                <a href="/go-app" class="app-link">Open App</a>
                <br><br>
                <a href="/go-app/health" class="app-link" style="background: #28a745;">Health Check</a>
            </div>
        </div>
        
        <div style="text-align: center; margin-top: 30px; color: #666;">
            <p>Powered by Flow Deploy CLI</p>
        </div>
    </div>
</body>
</html>
EOF

# Create a simple nginx pod to serve the index page
kubectl create configmap index-html --from-file=index.html=/tmp/index.html 2>/dev/null || kubectl create configmap index-html --from-file=index.html=/tmp/index.html --dry-run=client -o yaml | kubectl apply -f -

cat > /tmp/nginx-pod.yaml << EOF
apiVersion: v1
kind: Pod
metadata:
  name: nginx-index
  labels:
    app: nginx-index
spec:
  containers:
  - name: nginx
    image: nginx:alpine
    ports:
    - containerPort: 80
    volumeMounts:
    - name: index-html
      mountPath: /usr/share/nginx/html
  volumes:
  - name: index-html
    configMap:
      name: index-html
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-index
spec:
  selector:
    app: nginx-index
  ports:
  - port: 80
    targetPort: 80
EOF

kubectl apply -f /tmp/nginx-pod.yaml

# Update the ingress to include the index page
kubectl patch ingress flow-apps-simple-ingress --type='merge' -p='
{
  "spec": {
    "rules": [
      {
        "host": "'$ALB_IP'",
        "http": {
          "paths": [
            {
              "path": "/",
              "pathType": "Prefix",
              "backend": {
                "service": {
                  "name": "nginx-index",
                  "port": {
                    "number": 80
                  }
                }
              }
            },
            {
              "path": "/nodejs-app",
              "pathType": "Prefix",
              "backend": {
                "service": {
                  "name": "nodejs-app",
                  "port": {
                    "number": 80
                  }
                }
              }
            }
          ]
        }
      }
    ]
  }
}'

# Clean up temp files
rm /tmp/simple-ingress.yaml /tmp/nginx-pod.yaml /tmp/index.html

echo -e "${GREEN}🎉 Simple DNS setup complete!${NC}"
echo -e "${YELLOW}Your apps are now accessible at:${NC}"
echo -e "  🌐 Main page: http://$ALB_IP/"
echo -e "  🟢 Node.js app: http://$ALB_IP/nodejs-app"
echo -e "  🐍 Python app: http://$ALB_IP/python-app"
echo -e "  🐹 Go app: http://$ALB_IP/go-app"
echo ""
echo -e "${YELLOW}Health checks:${NC}"
echo -e "  http://$ALB_IP/nodejs-app/health"
echo -e "  http://$ALB_IP/python-app/health"
echo -e "  http://$ALB_IP/go-app/health"
echo ""
echo -e "${BLUE}You can now open http://$ALB_IP/ in your browser!${NC}"
