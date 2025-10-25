# Environment Variables

This document describes the environment variables used by the Flow AI deployment system.

## Required Environment Variables

### AWS Configuration
- `AWS_ACCOUNT_ID` - Your AWS account ID (12-digit number)
  - Example: `123456789012`
  - Used for ECR repository URLs and IAM role ARNs

### AI Configuration
- `OPENAI_API_KEY` - OpenAI API key for AI-powered error analysis
  - Example: `sk-...`
  - Required for enhanced AI agent features

## Optional Environment Variables

### Cluster Configuration
- `CLUSTER_IP` - IP address of your Kubernetes cluster
  - Example: `98.90.165.126`
  - Used for sslip.io URLs in development

### AWS Region
- `AWS_REGION` - AWS region (defaults to `us-east-1`)
  - Example: `us-west-2`

## Setting Environment Variables

### For Development
```bash
export AWS_ACCOUNT_ID="123456789012"
export OPENAI_API_KEY="sk-your-key-here"
export CLUSTER_IP="98.90.165.126"
```

### For Production
Create a `.env` file (not committed to git):
```bash
AWS_ACCOUNT_ID=123456789012
OPENAI_API_KEY=sk-your-key-here
CLUSTER_IP=98.90.165.126
```

## Security Notes

- Never commit actual API keys or account IDs to version control
- Use environment variables or secure secret management systems
- The `.gitignore` file is configured to exclude `.env` files
- Replace placeholder values with your actual values before deployment

## Script Usage

Most shell scripts will use these environment variables with fallbacks:
```bash
ACCOUNT_ID="${AWS_ACCOUNT_ID:-YOUR_AWS_ACCOUNT_ID}"
```

This means:
- Use `AWS_ACCOUNT_ID` if set
- Otherwise, use `YOUR_AWS_ACCOUNT_ID` as a placeholder
- You should replace `YOUR_AWS_ACCOUNT_ID` with your actual account ID
