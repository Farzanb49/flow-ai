flow-ai scaffold (placeholders for EKS/Knative/ECR)

Prereqs:
- AWS CLI v2, kubectl, pack, docker, Node 18+, Go 1.22+
- Knative Serving installed on target cluster (install later if not ready)

Quickstart:
1) API: (cd server && go run .)
2) Web (setup below) shows deployments from API
3) CLI build: (cd cmd/deployer && go mod tidy && go build -o flow .)
4) Build with Paketo:
   ./flow build --app /path/to/app \
     --image 000000000000.dkr.ecr.us-east-1.amazonaws.com/apps/myapp:latest
5) Push to ECR (placeholder region/account ok to replace later):
   ./flow push --image 000000000000.dkr.ecr.us-east-1.amazonaws.com/apps/myapp:latest
6) Deploy Knative Service (no EKS details required if kubeconfig current):
   ./flow deploy --name myapp --image 000000000000.dkr.ecr.us-east-1.amazonaws.com/apps/myapp:latest \
     --namespace default --server http://localhost:8080 --kubecontext ""
7) Attach DB/Redis:
   ./flow attach-db --name myapp --host HOST --db DB --user app --password secret
   ./flow attach-redis --name myapp --host HOST --password secret

Web UI:
- Create Vite app: (cd web && npm create vite@latest . -- --template react-ts && npm i && npm i axios)
- Replace src/App.tsx and add src/api.ts from instructions later.
