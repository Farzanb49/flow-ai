#!/bin/bash

echo "🚀 Deploying Full-Stack Task Manager Application"
echo "================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if flow CLI exists
if [ ! -f "/Users/farzanbhuiyan/flow-ai/flow" ]; then
    print_error "Flow CLI not found. Please build it first with: go build -o flow cmd/deployer/*.go"
    exit 1
fi

# Deploy backend first
print_status "Deploying Task Manager Backend..."
cd /Users/farzanbhuiyan/flow-ai/test-apps/task-manager-backend

if /Users/farzanbhuiyan/flow-ai/flow deploy; then
    print_success "Backend deployed successfully!"
    
    # Get backend URL
    BACKEND_URL=$(kubectl get ksvc task-manager-backend -o jsonpath='{.status.url}' 2>/dev/null)
    if [ -z "$BACKEND_URL" ]; then
        print_warning "Could not get backend URL. You may need to wait for the service to be ready."
        BACKEND_URL="http://task-manager-backend.default.${CLUSTER_IP}.sslip.io"
    fi
    
    print_status "Backend URL: $BACKEND_URL"
    
    # Wait a moment for backend to be ready
    print_status "Waiting for backend to be ready..."
    sleep 10
    
    # Test backend health
    print_status "Testing backend health..."
    if curl -s "$BACKEND_URL/health" > /dev/null; then
        print_success "Backend is healthy!"
    else
        print_warning "Backend health check failed, but continuing with frontend deployment..."
    fi
    
else
    print_error "Backend deployment failed!"
    exit 1
fi

# Deploy frontend
print_status "Deploying Task Manager Frontend..."
cd /Users/farzanbhuiyan/flow-ai/test-apps/task-manager-frontend

# Set backend URL for frontend
export BACKEND_URL="$BACKEND_URL"

if /Users/farzanbhuiyan/flow-ai/flow deploy; then
    print_success "Frontend deployed successfully!"
    
    # Get frontend URL
    FRONTEND_URL=$(kubectl get ksvc task-manager-frontend -o jsonpath='{.status.url}' 2>/dev/null)
    if [ -z "$FRONTEND_URL" ]; then
        print_warning "Could not get frontend URL. You may need to wait for the service to be ready."
        FRONTEND_URL="http://task-manager-frontend.default.${CLUSTER_IP}.sslip.io"
    fi
    
    print_success "Frontend URL: $FRONTEND_URL"
    
    # Wait for frontend to be ready
    print_status "Waiting for frontend to be ready..."
    sleep 15
    
    # Test frontend
    print_status "Testing frontend..."
    if curl -s "$FRONTEND_URL/health" > /dev/null; then
        print_success "Frontend is healthy!"
    else
        print_warning "Frontend health check failed, but deployment completed."
    fi
    
    echo ""
    echo "🎉 Full-Stack Application Deployed Successfully!"
    echo "================================================"
    echo "📱 Frontend: $FRONTEND_URL"
    echo "🔧 Backend:  $BACKEND_URL"
    echo "📊 Backend Health: $BACKEND_URL/health"
    echo "📊 Frontend Health: $FRONTEND_URL/health"
    echo ""
    echo "💡 You can now open the frontend URL in your browser to use the Task Manager!"
    
else
    print_error "Frontend deployment failed!"
    exit 1
fi
