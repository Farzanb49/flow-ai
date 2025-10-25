const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const morgan = require('morgan');
const compression = require('compression');

const app = express();
const PORT = process.env.PORT || 3000;
const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080';

// Middleware
app.use(helmet({
  contentSecurityPolicy: false // Allow inline styles for simplicity
}));
app.use(compression());
app.use(morgan('combined'));
app.use(cors());
app.use(express.json());

// Serve static files
app.use(express.static('public'));

// Health check
app.get('/health', (req, res) => {
  res.json({ 
    status: 'healthy', 
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
    version: '1.0.0'
  });
});

// API proxy to backend
app.use('/api', (req, res) => {
  const fetch = require('node-fetch');
  const url = `${BACKEND_URL}${req.originalUrl}`;
  
  fetch(url, {
    method: req.method,
    headers: {
      'Content-Type': 'application/json',
      ...req.headers
    },
    body: req.method !== 'GET' ? JSON.stringify(req.body) : undefined
  })
  .then(response => response.json())
  .then(data => res.json(data))
  .catch(error => {
    console.error('Backend proxy error:', error);
    res.status(500).json({ 
      success: false, 
      message: 'Backend service unavailable',
      error: error.message 
    });
  });
});

// Serve the main app
app.get('*', (req, res) => {
  res.send(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Task Manager</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            color: #333;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        
        .header {
            text-align: center;
            margin-bottom: 40px;
            color: white;
        }
        
        .header h1 {
            font-size: 3rem;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
        }
        
        .header p {
            font-size: 1.2rem;
            opacity: 0.9;
        }
        
        .main-content {
            display: grid;
            grid-template-columns: 1fr 2fr;
            gap: 30px;
            margin-bottom: 30px;
        }
        
        .sidebar {
            background: white;
            border-radius: 15px;
            padding: 30px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.1);
            height: fit-content;
        }
        
        .stats {
            margin-bottom: 30px;
        }
        
        .stats h3 {
            margin-bottom: 20px;
            color: #667eea;
            font-size: 1.5rem;
        }
        
        .stat-item {
            display: flex;
            justify-content: space-between;
            padding: 10px 0;
            border-bottom: 1px solid #eee;
        }
        
        .stat-item:last-child {
            border-bottom: none;
        }
        
        .stat-value {
            font-weight: bold;
            color: #667eea;
        }
        
        .task-form {
            background: white;
            border-radius: 15px;
            padding: 30px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.1);
        }
        
        .form-group {
            margin-bottom: 20px;
        }
        
        .form-group label {
            display: block;
            margin-bottom: 8px;
            font-weight: 600;
            color: #333;
        }
        
        .form-group input,
        .form-group select,
        .form-group textarea {
            width: 100%;
            padding: 12px;
            border: 2px solid #e1e5e9;
            border-radius: 8px;
            font-size: 16px;
            transition: border-color 0.3s;
        }
        
        .form-group input:focus,
        .form-group select:focus,
        .form-group textarea:focus {
            outline: none;
            border-color: #667eea;
        }
        
        .btn {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
            width: 100%;
        }
        
        .btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
        }
        
        .tasks-container {
            background: white;
            border-radius: 15px;
            padding: 30px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.1);
        }
        
        .tasks-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 30px;
        }
        
        .tasks-header h2 {
            color: #333;
            font-size: 2rem;
        }
        
        .filter-buttons {
            display: flex;
            gap: 10px;
        }
        
        .filter-btn {
            padding: 8px 16px;
            border: 2px solid #e1e5e9;
            background: white;
            border-radius: 20px;
            cursor: pointer;
            transition: all 0.3s;
        }
        
        .filter-btn.active {
            background: #667eea;
            color: white;
            border-color: #667eea;
        }
        
        .search-box {
            width: 100%;
            padding: 12px;
            border: 2px solid #e1e5e9;
            border-radius: 8px;
            margin-bottom: 20px;
            font-size: 16px;
        }
        
        .task-list {
            display: flex;
            flex-direction: column;
            gap: 15px;
        }
        
        .task-item {
            background: #f8f9fa;
            border-radius: 10px;
            padding: 20px;
            border-left: 4px solid #e1e5e9;
            transition: all 0.3s;
        }
        
        .task-item.todo {
            border-left-color: #ffc107;
        }
        
        .task-item.in-progress {
            border-left-color: #17a2b8;
        }
        
        .task-item.done {
            border-left-color: #28a745;
        }
        
        .task-item.high {
            background: #fff5f5;
        }
        
        .task-item.medium {
            background: #fffbf0;
        }
        
        .task-item.low {
            background: #f0fff4;
        }
        
        .task-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 10px;
        }
        
        .task-title {
            font-size: 1.2rem;
            font-weight: 600;
            color: #333;
            margin-bottom: 5px;
        }
        
        .task-meta {
            display: flex;
            gap: 10px;
            align-items: center;
        }
        
        .task-status {
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: 600;
            text-transform: uppercase;
        }
        
        .task-status.todo {
            background: #fff3cd;
            color: #856404;
        }
        
        .task-status.in-progress {
            background: #d1ecf1;
            color: #0c5460;
        }
        
        .task-status.done {
            background: #d4edda;
            color: #155724;
        }
        
        .task-priority {
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: 600;
            text-transform: uppercase;
        }
        
        .task-priority.high {
            background: #f8d7da;
            color: #721c24;
        }
        
        .task-priority.medium {
            background: #fff3cd;
            color: #856404;
        }
        
        .task-priority.low {
            background: #d4edda;
            color: #155724;
        }
        
        .task-description {
            color: #666;
            margin-bottom: 15px;
            line-height: 1.5;
        }
        
        .task-actions {
            display: flex;
            gap: 10px;
        }
        
        .btn-small {
            padding: 6px 12px;
            font-size: 12px;
            border-radius: 6px;
            border: none;
            cursor: pointer;
            transition: all 0.3s;
        }
        
        .btn-edit {
            background: #17a2b8;
            color: white;
        }
        
        .btn-delete {
            background: #dc3545;
            color: white;
        }
        
        .btn-complete {
            background: #28a745;
            color: white;
        }
        
        .btn-small:hover {
            transform: translateY(-1px);
            box-shadow: 0 2px 8px rgba(0,0,0,0.2);
        }
        
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #666;
        }
        
        .empty-state h3 {
            margin-bottom: 10px;
            color: #333;
        }
        
        .loading {
            text-align: center;
            padding: 40px;
            color: #666;
        }
        
        .error {
            background: #f8d7da;
            color: #721c24;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        
        .success {
            background: #d4edda;
            color: #155724;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        
        @media (max-width: 768px) {
            .main-content {
                grid-template-columns: 1fr;
            }
            
            .header h1 {
                font-size: 2rem;
            }
            
            .tasks-header {
                flex-direction: column;
                gap: 20px;
                align-items: stretch;
            }
            
            .filter-buttons {
                justify-content: center;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📋 Task Manager</h1>
            <p>Organize your tasks with style and efficiency</p>
        </div>
        
        <div class="main-content">
            <div class="sidebar">
                <div class="stats">
                    <h3>📊 Statistics</h3>
                    <div id="stats-container">
                        <div class="loading">Loading statistics...</div>
                    </div>
                </div>
            </div>
            
            <div class="task-form">
                <h2>➕ Add New Task</h2>
                <form id="task-form">
                    <div class="form-group">
                        <label for="title">Title *</label>
                        <input type="text" id="title" name="title" required>
                    </div>
                    <div class="form-group">
                        <label for="description">Description</label>
                        <textarea id="description" name="description" rows="3"></textarea>
                    </div>
                    <div class="form-group">
                        <label for="priority">Priority</label>
                        <select id="priority" name="priority">
                            <option value="low">Low</option>
                            <option value="medium" selected>Medium</option>
                            <option value="high">High</option>
                        </select>
                    </div>
                    <button type="submit" class="btn">Create Task</button>
                </form>
            </div>
        </div>
        
        <div class="tasks-container">
            <div class="tasks-header">
                <h2>📝 Your Tasks</h2>
                <div class="filter-buttons">
                    <button class="filter-btn active" data-status="all">All</button>
                    <button class="filter-btn" data-status="todo">Todo</button>
                    <button class="filter-btn" data-status="in-progress">In Progress</button>
                    <button class="filter-btn" data-status="done">Done</button>
                </div>
            </div>
            <input type="text" class="search-box" id="search" placeholder="🔍 Search tasks...">
            <div id="tasks-container">
                <div class="loading">Loading tasks...</div>
            </div>
        </div>
    </div>

    <script>
        class TaskManager {
            constructor() {
                this.tasks = [];
                this.currentFilter = 'all';
                this.searchTerm = '';
                this.init();
            }
            
            async init() {
                await this.loadStats();
                await this.loadTasks();
                this.setupEventListeners();
            }
            
            setupEventListeners() {
                // Task form
                document.getElementById('task-form').addEventListener('submit', (e) => {
                    e.preventDefault();
                    this.createTask();
                });
                
                // Filter buttons
                document.querySelectorAll('.filter-btn').forEach(btn => {
                    btn.addEventListener('click', (e) => {
                        document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
                        e.target.classList.add('active');
                        this.currentFilter = e.target.dataset.status;
                        this.renderTasks();
                    });
                });
                
                // Search
                document.getElementById('search').addEventListener('input', (e) => {
                    this.searchTerm = e.target.value.toLowerCase();
                    this.renderTasks();
                });
            }
            
            async loadStats() {
                try {
                    const response = await fetch('/api/stats');
                    const data = await response.json();
                    
                    if (data.success) {
                        const stats = data.data;
                        document.getElementById('stats-container').innerHTML = \`
                            <div class="stat-item">
                                <span>Total Tasks</span>
                                <span class="stat-value">\${stats.total}</span>
                            </div>
                            <div class="stat-item">
                                <span>To Do</span>
                                <span class="stat-value">\${stats.todo}</span>
                            </div>
                            <div class="stat-item">
                                <span>In Progress</span>
                                <span class="stat-value">\${stats.inProgress}</span>
                            </div>
                            <div class="stat-item">
                                <span>Done</span>
                                <span class="stat-value">\${stats.done}</span>
                            </div>
                            <div class="stat-item">
                                <span>High Priority</span>
                                <span class="stat-value">\${stats.high}</span>
                            </div>
                            <div class="stat-item">
                                <span>Medium Priority</span>
                                <span class="stat-value">\${stats.medium}</span>
                            </div>
                            <div class="stat-item">
                                <span>Low Priority</span>
                                <span class="stat-value">\${stats.low}</span>
                            </div>
                        \`;
                    }
                } catch (error) {
                    console.error('Error loading stats:', error);
                    document.getElementById('stats-container').innerHTML = '<div class="error">Failed to load statistics</div>';
                }
            }
            
            async loadTasks() {
                try {
                    const response = await fetch('/api/tasks');
                    const data = await response.json();
                    
                    if (data.success) {
                        this.tasks = data.data;
                        this.renderTasks();
                    } else {
                        this.showError('Failed to load tasks');
                    }
                } catch (error) {
                    console.error('Error loading tasks:', error);
                    this.showError('Failed to load tasks');
                }
            }
            
            async createTask() {
                const formData = new FormData(document.getElementById('task-form'));
                const taskData = {
                    title: formData.get('title'),
                    description: formData.get('description'),
                    priority: formData.get('priority')
                };
                
                try {
                    const response = await fetch('/api/tasks', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        body: JSON.stringify(taskData)
                    });
                    
                    const data = await response.json();
                    
                    if (data.success) {
                        this.showSuccess('Task created successfully!');
                        document.getElementById('task-form').reset();
                        await this.loadTasks();
                        await this.loadStats();
                    } else {
                        this.showError(data.message || 'Failed to create task');
                    }
                } catch (error) {
                    console.error('Error creating task:', error);
                    this.showError('Failed to create task');
                }
            }
            
            async updateTask(id, updates) {
                try {
                    const response = await fetch(\`/api/tasks/\${id}\`, {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        body: JSON.stringify(updates)
                    });
                    
                    const data = await response.json();
                    
                    if (data.success) {
                        this.showSuccess('Task updated successfully!');
                        await this.loadTasks();
                        await this.loadStats();
                    } else {
                        this.showError(data.message || 'Failed to update task');
                    }
                } catch (error) {
                    console.error('Error updating task:', error);
                    this.showError('Failed to update task');
                }
            }
            
            async deleteTask(id) {
                if (!confirm('Are you sure you want to delete this task?')) {
                    return;
                }
                
                try {
                    const response = await fetch(\`/api/tasks/\${id}\`, {
                        method: 'DELETE'
                    });
                    
                    const data = await response.json();
                    
                    if (data.success) {
                        this.showSuccess('Task deleted successfully!');
                        await this.loadTasks();
                        await this.loadStats();
                    } else {
                        this.showError(data.message || 'Failed to delete task');
                    }
                } catch (error) {
                    console.error('Error deleting task:', error);
                    this.showError('Failed to delete task');
                }
            }
            
            renderTasks() {
                const container = document.getElementById('tasks-container');
                
                let filteredTasks = this.tasks;
                
                // Apply filter
                if (this.currentFilter !== 'all') {
                    filteredTasks = filteredTasks.filter(task => task.status === this.currentFilter);
                }
                
                // Apply search
                if (this.searchTerm) {
                    filteredTasks = filteredTasks.filter(task => 
                        task.title.toLowerCase().includes(this.searchTerm) ||
                        task.description.toLowerCase().includes(this.searchTerm)
                    );
                }
                
                if (filteredTasks.length === 0) {
                    container.innerHTML = \`
                        <div class="empty-state">
                            <h3>No tasks found</h3>
                            <p>\${this.searchTerm ? 'Try adjusting your search terms' : 'Create your first task to get started!'}</p>
                        </div>
                    \`;
                    return;
                }
                
                container.innerHTML = filteredTasks.map(task => \`
                    <div class="task-item \${task.status} \${task.priority}">
                        <div class="task-header">
                            <div>
                                <div class="task-title">\${task.title}</div>
                                <div class="task-meta">
                                    <span class="task-status \${task.status}">\${task.status.replace('-', ' ')}</span>
                                    <span class="task-priority \${task.priority}">\${task.priority}</span>
                                </div>
                            </div>
                        </div>
                        \${task.description ? \`<div class="task-description">\${task.description}</div>\` : ''}
                        <div class="task-actions">
                            \${task.status !== 'done' ? \`
                                <button class="btn-small btn-complete" onclick="taskManager.updateTask('\${task.id}', {status: 'done'})">
                                    ✓ Complete
                                </button>
                            \` : ''}
                            \${task.status === 'todo' ? \`
                                <button class="btn-small btn-edit" onclick="taskManager.updateTask('\${task.id}', {status: 'in-progress'})">
                                    ▶ Start
                                </button>
                            \` : ''}
                            <button class="btn-small btn-delete" onclick="taskManager.deleteTask('\${task.id}')">
                                🗑 Delete
                            </button>
                        </div>
                    </div>
                \`).join('');
            }
            
            showSuccess(message) {
                this.showMessage(message, 'success');
            }
            
            showError(message) {
                this.showMessage(message, 'error');
            }
            
            showMessage(message, type) {
                const container = document.querySelector('.container');
                const messageDiv = document.createElement('div');
                messageDiv.className = type;
                messageDiv.textContent = message;
                messageDiv.style.position = 'fixed';
                messageDiv.style.top = '20px';
                messageDiv.style.right = '20px';
                messageDiv.style.zIndex = '1000';
                messageDiv.style.minWidth = '300px';
                
                container.appendChild(messageDiv);
                
                setTimeout(() => {
                    messageDiv.remove();
                }, 3000);
            }
        }
        
        // Initialize the app
        const taskManager = new TaskManager();
    </script>
</body>
</html>
  `);
});

// Start server
app.listen(PORT, '0.0.0.0', () => {
  console.log(`🎨 Task Manager Frontend running on port ${PORT}`);
  console.log(`📊 Health check: http://localhost:${PORT}/health`);
  console.log(`🌐 App: http://localhost:${PORT}`);
});

module.exports = app;
