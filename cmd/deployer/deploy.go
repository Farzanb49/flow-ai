package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type deployReport struct {
	ID          string    `json:"id"`
	Project     string    `json:"project"`
	Namespace   string    `json:"namespace"`
	Image       string    `json:"image"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

func newBuildCmd() *cobra.Command {
	var (
		appPath   string
		imageRef  string
		builder   string
		envs      []string
	)
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build OCI image with Paketo (pack build)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if appPath == "" || imageRef == "" {
				return fmt.Errorf("app and image are required")
			}
			if builder == "" {
				builder = "paketobuildpacks/builder:base"
			}
			packArgs := []string{"build", imageRef, "--path", appPath, "--builder", builder, "--pull-policy", "if-not-present"}
			for _, e := range envs { packArgs = append(packArgs, "--env", e) }
			c := exec.Command("pack", packArgs...)
			c.Stdout, c.Stderr = os.Stdout, os.Stderr
			return c.Run()
		},
	}
	cmd.Flags().StringVar(&appPath, "app", ".", "Path to application source")
	cmd.Flags().StringVar(&imageRef, "image", "", "Target image reference (e.g. 000000000000.dkr.ecr.us-east-1.amazonaws.com/apps/myapp:tag)")
	cmd.Flags().StringVar(&builder, "builder", "", "Paketo builder image")
	cmd.Flags().StringSliceVar(&envs, "env", []string{}, "Build env key=value")
	return cmd
}

func newPushCmd() *cobra.Command {
	var imageRef string
	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push built image to ECR (docker push)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if imageRef == "" {
				return fmt.Errorf("image is required")
			}
			return dockerPushWithECRLogin(imageRef)
		},
	}
	cmd.Flags().StringVar(&imageRef, "image", "", "Image to push")
	return cmd
}

func newDeployCmd() *cobra.Command {
	var (
		namespace     string
		port          int
		cpu           string
		mem           string
		envs          []string
		serverURL     string
		kubecontext   string
		dbHost        string
		dbName        string
		dbUser        string
		dbPassword    string
		dbPort        int
		redisHost     string
		redisPassword string
		redisPort     int
		secrets       []string
	)
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Build, push, and deploy application with optional database, cache, and secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Auto-detect project name from current directory
			projectName, err := getProjectName()
			if err != nil {
				return fmt.Errorf("failed to detect project name: %v", err)
			}
			
			// Auto-generate image reference
			imageRef, err := generateImageRef(projectName)
			if err != nil {
				return fmt.Errorf("failed to generate image reference: %v", err)
			}
			
			if namespace == "" { namespace = "default" }
			
			// Step 1: Build the application
			fmt.Printf("Building application %s...\n", projectName)
			if err := buildApplication(".", imageRef, envs); err != nil {
				return fmt.Errorf("build failed: %v", err)
			}
			
		// Step 2: Push to ECR
		fmt.Printf("Pushing image %s...\n", imageRef)
		if err := dockerPushWithECRLogin(imageRef); err != nil {
			return fmt.Errorf("push failed: %v", err)
		}
			
		// Step 3: Deploy Knative Service
		fmt.Printf("Deploying service %s...\n", projectName)
		if err := knServiceApply(projectName, imageRef, namespace, cpu, mem, toEnvMap(envs), kubecontext, port); err != nil {
			return fmt.Errorf("deploy failed: %v", err)
		}
		
		// Step 3.5: Configure ECR pull permissions
		fmt.Printf("Configuring ECR pull permissions...\n")
		if err := configureECRPullPermissions(projectName, namespace, kubecontext); err != nil {
			fmt.Printf("Warning: Failed to configure ECR pull permissions: %v\n", err)
			fmt.Printf("The service may not be able to pull images from ECR.\n")
		}
			
			// Step 4: Attach database if specified
			if dbHost != "" {
				fmt.Printf("Attaching database...\n")
				if err := attachDatabase(projectName, namespace, dbUser, dbPassword, dbHost, dbPort, dbName, kubecontext); err != nil {
					return fmt.Errorf("database attach failed: %v", err)
				}
			}
			
			// Step 5: Attach Redis if specified
			if redisHost != "" {
				fmt.Printf("Attaching Redis...\n")
				if err := attachRedis(projectName, namespace, redisHost, redisPort, redisPassword, kubecontext); err != nil {
					return fmt.Errorf("redis attach failed: %v", err)
				}
			}
			
			// Step 6: Create secrets if specified
			if len(secrets) > 0 {
				fmt.Printf("Creating secrets...\n")
				if err := createSecrets(projectName, namespace, secrets, kubecontext); err != nil {
					return fmt.Errorf("secrets creation failed: %v", err)
				}
			}
			
			// Report deployment
			if serverURL != "" {
				_ = report(serverURL, deployReport{
					ID:          fmt.Sprintf("%s:%d", projectName, time.Now().UnixNano()),
					Project:     projectName,
					Namespace:   namespace,
					Image:       imageRef,
					Status:      "deployed",
					Description: "Application deployed with auto-build and push",
					CreatedAt:   time.Now(),
				})
			}
			
			
			fmt.Printf("Successfully deployed %s to namespace %s\n", projectName, namespace)
			return nil
		},
	}
	
	// Core deployment flags
	cmd.Flags().StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	cmd.Flags().IntVar(&port, "port", 8080, "Service port")
	cmd.Flags().StringVar(&cpu, "cpu", "250m", "CPU request/limit")
	cmd.Flags().StringVar(&mem, "mem", "256Mi", "Memory request/limit")
	cmd.Flags().StringSliceVar(&envs, "env", []string{}, "Runtime environment variables (key=value)")
	cmd.Flags().StringVar(&serverURL, "server", "", "API server to report deployments")
	cmd.Flags().StringVar(&kubecontext, "kubecontext", "", "kubectl context to use")
	
	// Database flags
	cmd.Flags().StringVar(&dbHost, "db-host", "", "Database host")
	cmd.Flags().StringVar(&dbName, "db-name", "", "Database name")
	cmd.Flags().StringVar(&dbUser, "db-user", "app", "Database user")
	cmd.Flags().StringVar(&dbPassword, "db-password", "changeme", "Database password")
	cmd.Flags().IntVar(&dbPort, "db-port", 5432, "Database port")
	
	// Redis flags
	cmd.Flags().StringVar(&redisHost, "redis-host", "", "Redis host")
	cmd.Flags().StringVar(&redisPassword, "redis-password", "", "Redis password")
	cmd.Flags().IntVar(&redisPort, "redis-port", 6379, "Redis port")
	
	// Secrets flags
	cmd.Flags().StringSliceVar(&secrets, "secret", []string{}, "Secret key=value pairs")
	
	
	return cmd
}

func dockerPushWithECRLogin(imageRef string) error {
	// Get AWS account ID and region
	region := os.Getenv("AWS_REGION")
	if region == "" { region = os.Getenv("AWS_DEFAULT_REGION") }
	if region == "" { region = "us-east-1" }
	
	// Get AWS account ID
	cmd := exec.Command("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get AWS account ID for ECR push: %v\n\nTroubleshooting:\n1. Run: aws configure\n2. Ensure your AWS credentials are valid\n3. Set AWS_ACCOUNT_ID environment variable", err)
	}
	accountID := strings.TrimSpace(string(out))
	if accountID == "" {
		return fmt.Errorf("AWS account ID is empty")
	}
	
	// Handle local image names (e.g., "nodejs-app:latest")
	var ecrImageRef string
	if !strings.Contains(imageRef, ".dkr.ecr.") {
		// This is a local image, need to tag it for ECR
		projectName := strings.Split(imageRef, ":")[0]
		ecrImageRef = fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:latest", accountID, region, projectName)
		
		fmt.Printf("Tagging local image %s as %s...\n", imageRef, ecrImageRef)
		tagCmd := exec.Command("docker", "tag", imageRef, ecrImageRef)
		tagCmd.Stdout, tagCmd.Stderr = os.Stdout, os.Stderr
		if err := tagCmd.Run(); err != nil {
			return fmt.Errorf("failed to tag image for ECR: %v", err)
		}
		imageRef = ecrImageRef
	}
	
	parts := strings.Split(imageRef, "/")
	if len(parts) < 1 { return fmt.Errorf("invalid image: %s", imageRef) }
	reg := parts[0]

	// Extract repository name from image reference
	repoName := strings.Split(imageRef, "/")[1]
	repoName = strings.Split(repoName, ":")[0]

	fmt.Printf("Authenticating with ECR...\n")
	authCmd := exec.Command("aws", "ecr", "get-login-password", "--region", region)
	authOut, err := authCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get ECR login password: %v\n\nTroubleshooting:\n1. Run: aws configure\n2. Ensure your AWS credentials are valid\n3. Check if you have ECR permissions", err)
	}

	auth := exec.Command("docker", "login", "--username", "AWS", "--password-stdin", reg)
	auth.Stdin = bytes.NewReader(authOut)
	auth.Stdout, auth.Stderr = os.Stdout, os.Stderr
	if err := auth.Run(); err != nil {
		return fmt.Errorf("failed to login to ECR: %v", err)
	}

	// Check if repository exists, create if not
	fmt.Printf("Checking ECR repository: %s\n", repoName)
	checkRepo := exec.Command("aws", "ecr", "describe-repositories", "--repository-names", repoName, "--region", region)
	if err := checkRepo.Run(); err != nil {
		fmt.Printf("Repository %s not found, creating...\n", repoName)
		createRepo := exec.Command("aws", "ecr", "create-repository", "--repository-name", repoName, "--region", region)
		createRepo.Stdout, createRepo.Stderr = os.Stdout, os.Stderr
		if err := createRepo.Run(); err != nil {
			return fmt.Errorf("failed to create ECR repository %s: %v\n\nTroubleshooting:\n1. Ensure you have ecr:CreateRepository permission\n2. Run: ./setup-ecr.sh", repoName, err)
		}
		fmt.Printf("Repository %s created successfully\n", repoName)
	}

	fmt.Printf("Pushing image %s...\n", imageRef)
	push := exec.Command("docker", "push", imageRef)
	push.Stdout, push.Stderr = os.Stdout, os.Stderr
	if err := push.Run(); err != nil {
		return fmt.Errorf("failed to push image to ECR: %v\n\nTroubleshooting:\n1. Check ECR permissions\n2. Ensure repository exists\n3. Run: ./setup-ecr.sh", err)
	}

	return nil
}

func toEnvMap(pairs []string) map[string]string {
	m := map[string]string{}
	for _, p := range pairs {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) == 2 { m[kv[0]] = kv[1] }
	}
	return m
}

func knServiceApply(name, image, ns, cpu, mem string, env map[string]string, kubecontext string, port int) error {
	y := newKnServiceYAML(name, image, ns, cpu, mem, env, port)
	args := []string{"apply", "-f", "-"}
	if kubecontext != "" { args = append([]string{"--context", kubecontext}, args...) }
	c := exec.Command("kubectl", args...)
	c.Stdin = bytes.NewBufferString(y)
	c.Stdout, c.Stderr = os.Stdout, os.Stderr
	return c.Run()
}

func newKnServiceYAML(name, image, ns, cpu, mem string, env map[string]string, port int) string {
	var envLines []string
	for k, v := range env {
		envLines = append(envLines, fmt.Sprintf("          - name: %s\n            value: \"%s\"", k, v))
	}
	envBlock := ""
	if len(envLines) > 0 { envBlock = "        env:\n" + strings.Join(envLines, "\n") }
	return fmt.Sprintf(`apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: %s
  namespace: %s
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"
    spec:
      containers:
        - image: %s
          ports:
            - containerPort: %d
              name: http1
%s
`, name, ns, image, port, envBlock)
}

func getClient(namespace string) (*kubernetes.Clientset, string, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" { kubeconfig = clientcmd.RecommendedHomeFile }
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil { return nil, "", err }
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil { return nil, "", err }
	if namespace == "" { namespace = "default" }
	return clientset, namespace, nil
}

func report(server string, payload deployReport) error {
	b, _ := json.Marshal(payload)
	c := exec.Command("curl", "-sS", "-X", "POST", "-H", "Content-Type: application/json", "-d", string(b), server+"/deployments")
	c.Stdout, c.Stderr = os.Stdout, os.Stderr
	return c.Run()
}

// Helper functions for the new deploy command

func getProjectName() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	parts := strings.Split(wd, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("unable to determine project name from current directory")
	}
	return parts[len(parts)-1], nil
}

func generateImageRef(projectName string) (string, error) {
	// Get AWS account ID and region
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if region == "" {
		region = "us-east-1"
	}
	
	// Try to get AWS account ID from AWS CLI
	cmd := exec.Command("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text")
	out, err := cmd.Output()
	if err != nil {
		// If AWS CLI fails, try to get account ID from environment or use a fallback
		accountID := os.Getenv("AWS_ACCOUNT_ID")
		if accountID == "" {
			// Try to parse from existing ECR URLs in environment
			ecrURL := os.Getenv("ECR_URL")
			if ecrURL != "" {
				// Extract account ID from ECR URL like "123456789012.dkr.ecr.us-east-1.amazonaws.com"
				parts := strings.Split(ecrURL, ".")
				if len(parts) > 0 && strings.Contains(parts[0], "dkr") {
					accountID = strings.Split(parts[0], ".")[0]
				}
			}
		}
		
		if accountID == "" {
			// Last resort: use a local image name for building, will be tagged for ECR during push
			fmt.Printf("⚠️  Could not determine AWS account ID. Using local image name for building.\n")
			imageRef := fmt.Sprintf("%s:latest", projectName)
			return imageRef, nil
		}
		
		imageRef := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:latest", accountID, region, projectName)
		return imageRef, nil
	}
	
	accountID := strings.TrimSpace(string(out))
	if accountID == "" {
		return "", fmt.Errorf("AWS account ID is empty")
	}
	
	// Generate image reference
	imageRef := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:latest", accountID, region, projectName)
	return imageRef, nil
}


func buildApplication(appPath, imageRef string, envs []string) error {
	// Try to use bundled pack CLI first, fallback to system pack
	packPath := findPackCLI()
	if packPath == "" {
		fmt.Printf("Pack CLI not found, falling back to Docker build...\n")
		return buildWithDocker(appPath, imageRef, envs)
	}
	
	// Use a more stable builder image
	builder := "paketobuildpacks/builder:tiny"
	args := []string{"build", imageRef, "--path", appPath, "--builder", builder, "--pull-policy", "always", "--verbose"}
	for _, e := range envs {
		args = append(args, "--env", e)
	}
	
	fmt.Printf("Running pack command: %s %v\n", packPath, args)
	c := exec.Command(packPath, args...)
	c.Stdout, c.Stderr = os.Stdout, os.Stderr
	
	// Try pack build first
	if err := c.Run(); err != nil {
		fmt.Printf("Pack build failed: %v\n", err)
		fmt.Println("Falling back to Docker build...")
		return buildWithDocker(appPath, imageRef, envs)
	}
	
	return nil
}

func buildWithDocker(appPath, imageRef string, envs []string) error {
	// Create a simple Dockerfile for the application
	dockerfile := createDockerfile(appPath)
	
	// Build with Docker for linux/amd64 platform (EKS compatibility)
	args := []string{"build", "--platform", "linux/amd64", "-t", imageRef, "-f", "-", appPath}
	c := exec.Command("docker", args...)
	c.Stdin = strings.NewReader(dockerfile)
	c.Stdout, c.Stderr = os.Stdout, os.Stderr
	
	return c.Run()
}

func createDockerfile(appPath string) string {
	// Detect the application type and create appropriate Dockerfile
	if _, err := os.Stat(filepath.Join(appPath, "package.json")); err == nil {
		// Node.js application
		return `FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install --omit=dev
COPY . .
EXPOSE 8080
CMD ["node", "server.js"]`
	} else if _, err := os.Stat(filepath.Join(appPath, "requirements.txt")); err == nil {
		// Python application
		return `FROM python:3.11-slim
RUN apt-get update && apt-get install -y gcc && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8080
CMD ["gunicorn", "--bind", "0.0.0.0:8080", "app:app"]`
	} else if _, err := os.Stat(filepath.Join(appPath, "go.mod")); err == nil {
		// Go application
		return `FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]`
	}
	
	// Default Dockerfile
	return `FROM alpine:latest
WORKDIR /app
COPY . .
EXPOSE 8080
CMD ["echo", "No specific buildpack detected, using default"]`
}

func findPackCLI() string {
	// First try to find bundled pack CLI
	if bundledPath := findBundledPack(); bundledPath != "" {
		return bundledPath
	}
	
	// Fallback to system pack CLI
	if systemPath, err := exec.LookPath("pack"); err == nil {
		return systemPath
	}
	
	return ""
}

func findBundledPack() string {
	// Look for pack binary in the cmd/deployer directory
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}
	
	execDir := filepath.Dir(execPath)
	// Look in cmd/deployer subdirectory
	deployerDir := filepath.Join(execDir, "cmd", "deployer")
	packPath := filepath.Join(deployerDir, "pack")
	
	// Check if pack binary exists and is executable
	if info, err := os.Stat(packPath); err == nil && !info.IsDir() {
		// Make sure it's executable
		os.Chmod(packPath, 0755)
		return packPath
	}
	
	// Also try in the same directory as executable
	packPath = filepath.Join(execDir, "pack")
	if info, err := os.Stat(packPath); err == nil && !info.IsDir() {
		os.Chmod(packPath, 0755)
		return packPath
	}
	
	return ""
}

func attachDatabase(name, namespace, user, password, host string, port int, db string, kubecontext string) error {
	client, ns, err := getClient(namespace)
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, db)
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-db",
			Namespace: ns,
		},
		StringData: map[string]string{"DATABASE_URL": url},
	}
	
	_, err = client.CoreV1().Secrets(ns).Get(context.Background(), sec.Name, metav1.GetOptions{})
	if err == nil {
		_, err = client.CoreV1().Secrets(ns).Update(context.Background(), sec, metav1.UpdateOptions{})
		return err
	}
	_, err = client.CoreV1().Secrets(ns).Create(context.Background(), sec, metav1.CreateOptions{})
	return err
}

func attachRedis(name, namespace, host string, port int, password string, kubecontext string) error {
	client, ns, err := getClient(namespace)
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("redis://:%s@%s:%d", password, host, port)
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-redis",
			Namespace: ns,
		},
		StringData: map[string]string{"REDIS_URL": url},
	}
	
	_, err = client.CoreV1().Secrets(ns).Get(context.Background(), sec.Name, metav1.GetOptions{})
	if err == nil {
		_, err = client.CoreV1().Secrets(ns).Update(context.Background(), sec, metav1.UpdateOptions{})
		return err
	}
	_, err = client.CoreV1().Secrets(ns).Create(context.Background(), sec, metav1.CreateOptions{})
	return err
}

func createSecrets(name, namespace string, pairs []string, kubecontext string) error {
	client, ns, err := getClient(namespace)
	if err != nil {
		return err
	}
	
	data := map[string]string{}
	for _, p := range pairs {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid secret pair: %s", p)
		}
		data[kv[0]] = kv[1]
	}
	
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-secrets",
			Namespace: ns,
		},
		StringData: data,
	}
	
	_, err = client.CoreV1().Secrets(ns).Get(context.Background(), sec.Name, metav1.GetOptions{})
	if err == nil {
		_, err = client.CoreV1().Secrets(ns).Update(context.Background(), sec, metav1.UpdateOptions{})
		return err
	}
	_, err = client.CoreV1().Secrets(ns).Create(context.Background(), sec, metav1.CreateOptions{})
	return err
}
