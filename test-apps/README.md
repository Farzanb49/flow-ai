# Full-Stack Task Manager Application

A modern, full-stack task management application built with Node.js and Express, featuring a beautiful React-like frontend and a robust REST API backend.

## 🏗️ Architecture

### Backend (`task-manager-backend`)
- **Framework**: Node.js with Express
- **Features**: 
  - RESTful API with CRUD operations
  - Task management (create, read, update, delete)
  - Task filtering and search
  - Statistics and analytics
  - Rate limiting and security middleware
  - Health check endpoints

### Frontend (`task-manager-frontend`)
- **Framework**: Vanilla JavaScript with modern ES6+ features
- **UI**: Responsive design with CSS Grid and Flexbox
- **Features**:
  - Beautiful, modern interface
  - Real-time task management
  - Task filtering (All, Todo, In Progress, Done)
  - Search functionality
  - Statistics dashboard
  - Mobile-responsive design

## 🚀 Quick Start

### Prerequisites
- Flow CLI built and ready
- Kubernetes cluster with Knative installed
- AWS ECR access configured

### Deploy Both Services

```bash
# Navigate to the test-apps directory
cd /Users/farzanbhuiyan/flow-ai/test-apps

# Run the deployment script
./deploy-fullstack.sh
```

### Deploy Services Individually

#### Backend Only
```bash
cd task-manager-backend
flow deploy
```

#### Frontend Only
```bash
cd task-manager-frontend
BACKEND_URL=http://your-backend-url flow deploy
```

## 📱 Features

### Task Management
- ✅ Create new tasks with title, description, and priority
- ✅ Update task status (Todo → In Progress → Done)
- ✅ Delete tasks
- ✅ Filter tasks by status
- ✅ Search tasks by title or description
- ✅ Priority levels (High, Medium, Low)

### Statistics Dashboard
- 📊 Total task count
- 📊 Tasks by status (Todo, In Progress, Done)
- 📊 Tasks by priority (High, Medium, Low)
- 📊 Real-time updates

### User Experience
- 🎨 Modern, gradient-based design
- 📱 Mobile-responsive layout
- ⚡ Fast, real-time updates
- 🔍 Instant search and filtering
- 💫 Smooth animations and transitions
- 🎯 Intuitive task management workflow

## 🔧 API Endpoints

### Backend API (`/api/`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/tasks` | Get all tasks (with optional filtering) |
| GET | `/api/tasks/:id` | Get single task |
| POST | `/api/tasks` | Create new task |
| PUT | `/api/tasks/:id` | Update task |
| DELETE | `/api/tasks/:id` | Delete task |
| GET | `/api/stats` | Get task statistics |

### Query Parameters
- `status`: Filter by task status (todo, in-progress, done)
- `priority`: Filter by priority (high, medium, low)
- `search`: Search in title and description

## 🎨 UI Components

### Main Layout
- **Header**: Application title and description
- **Sidebar**: Statistics dashboard
- **Task Form**: Create new tasks
- **Task List**: Display and manage tasks

### Task Item
- **Status Indicators**: Color-coded status badges
- **Priority Indicators**: Color-coded priority badges
- **Action Buttons**: Complete, Start, Delete actions
- **Responsive Design**: Adapts to different screen sizes

## 🔒 Security Features

- **Helmet.js**: Security headers
- **CORS**: Cross-origin resource sharing
- **Rate Limiting**: API request throttling
- **Input Validation**: Server-side validation
- **Error Handling**: Comprehensive error management

## 📊 Monitoring

Both services include health check endpoints:
- Backend: `GET /health`
- Frontend: `GET /health`

Health checks return:
- Service status
- Timestamp
- Uptime
- Version information

## 🛠️ Development

### Local Development

#### Backend
```bash
cd task-manager-backend
npm install
npm run dev
```

#### Frontend
```bash
cd task-manager-frontend
npm install
BACKEND_URL=http://localhost:8080 npm start
```

### Environment Variables

#### Backend
- `PORT`: Server port (default: 8080)
- `NODE_ENV`: Environment (development/production)

#### Frontend
- `PORT`: Server port (default: 3000)
- `BACKEND_URL`: Backend API URL

## 🎯 Use Cases

This application demonstrates:
- **Microservices Architecture**: Separate frontend and backend services
- **Modern Web Development**: REST APIs, responsive design
- **Cloud Deployment**: Kubernetes, Knative, ECR
- **DevOps Practices**: Automated deployment, health checks
- **User Experience**: Intuitive, modern interface

## 🔄 Deployment Workflow

1. **Build**: Paketo buildpacks or Docker fallback
2. **Push**: Images pushed to AWS ECR
3. **Deploy**: Services deployed to Knative
4. **Configure**: ECR pull permissions set up automatically
5. **Access**: Public URLs provided for both services

## 📈 Scalability

- **Horizontal Scaling**: Knative auto-scaling
- **Load Balancing**: Kubernetes service mesh
- **Resource Management**: CPU and memory limits
- **Monitoring**: Health checks and metrics

This full-stack application showcases the power of the Flow CLI for deploying modern, production-ready applications to the cloud with minimal configuration.