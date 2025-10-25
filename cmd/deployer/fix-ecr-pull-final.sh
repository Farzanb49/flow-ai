#!/bin/bash

# Final ECR Pull Fix for EKS
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Fixing ECR pull for EKS cluster...${NC}"

# Get cluster info
CLUSTER_NAME="zanny-playground"
ACCOUNT_ID="${AWS_ACCOUNT_ID:-YOUR_AWS_ACCOUNT_ID}"
REGION="us-east-1"
OIDC_URL="oidc.eks.us-east-1.amazonaws.com/id/CDFA0163865769963393AAC95373631A"

echo -e "${YELLOW}Cluster: $CLUSTER_NAME${NC}"
echo -e "${YELLOW}Account ID: $ACCOUNT_ID${NC}"

# Delete existing resources
echo -e "${YELLOW}Cleaning up existing resources...${NC}"
kubectl delete serviceaccount ecr-pull-sa 2>/dev/null || true
kubectl delete secret ecr-secret 2>/dev/null || true

# Create a new ECR policy that's more permissive
echo -e "${YELLOW}Creating comprehensive ECR policy...${NC}"

cat > /tmp/ecr-policy.json << EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ecr:GetAuthorizationToken"
            ],
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "ecr:BatchCheckLayerAvailability",
                "ecr:GetDownloadUrlForLayer",
                "ecr:BatchGetImage",
                "ecr:DescribeRepositories",
                "ecr:DescribeImages"
            ],
            "Resource": "arn:aws:ecr:${REGION}:${ACCOUNT_ID}:repository/*"
        }
    ]
}
EOF

# Create or update the policy
aws iam create-policy --policy-name ECRPullPolicy --policy-document file:///tmp/ecr-policy.json 2>/dev/null || \
aws iam create-policy-version --policy-arn "arn:aws:iam::${ACCOUNT_ID}:policy/ECRPullPolicy" --policy-document file:///tmp/ecr-policy.json --set-as-default

# Create a simpler trust policy
cat > /tmp/trust-policy.json << EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Federated": "arn:aws:iam::${ACCOUNT_ID}:oidc-provider/${OIDC_URL}"
            },
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringEquals": {
                    "${OIDC_URL}:aud": "sts.amazonaws.com"
                },
                "StringLike": {
                    "${OIDC_URL}:sub": "system:serviceaccount:*:*"
                }
            }
        }
    ]
}
EOF

# Update the IAM role
echo -e "${YELLOW}Updating IAM role...${NC}"
aws iam update-assume-role-policy --role-name EKSECRPullRole --policy-document file:///tmp/trust-policy.json

# Attach the policy to the role
aws iam attach-role-policy --role-name EKSECRPullRole --policy-arn "arn:aws:iam::${ACCOUNT_ID}:policy/ECRPullPolicy"

# Create service account
echo -e "${YELLOW}Creating service account...${NC}"
kubectl create serviceaccount ecr-pull-sa
kubectl annotate serviceaccount ecr-pull-sa eks.amazonaws.com/role-arn=arn:aws:iam::${ACCOUNT_ID}:role/EKSECRPullRole

# Create ECR secret as backup
echo -e "${YELLOW}Creating ECR secret...${NC}"
ECR_TOKEN=$(aws ecr get-login-password --region $REGION)
kubectl create secret docker-registry ecr-secret \
  --docker-server=${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com \
  --docker-username=AWS \
  --docker-password=$ECR_TOKEN

# Add secret to service account
kubectl patch serviceaccount ecr-pull-sa -p '{"imagePullSecrets": [{"name": "ecr-secret"}]}'

# Test ECR access
echo -e "${YELLOW}Testing ECR access...${NC}"
kubectl run ecr-test --image=amazon/aws-cli:latest --rm -i --restart=Never --overrides='{"spec":{"serviceAccountName":"ecr-pull-sa","containers":[{"name":"ecr-test","image":"amazon/aws-cli:latest","command":["aws","ecr","describe-images","--repository-name","nodejs-app","--region","us-east-1"]}]}}' -- aws ecr describe-images --repository-name nodejs-app --region us-east-1

# Update the Knative service
echo -e "${YELLOW}Updating Knative service...${NC}"
kubectl patch ksvc nodejs-app --type='merge' -p='{"spec":{"template":{"spec":{"serviceAccountName":"ecr-pull-sa"}}}}'

# Clean up temp files
rm /tmp/ecr-policy.json /tmp/trust-policy.json

echo -e "${GREEN}ECR pull fix complete!${NC}"
echo -e "${YELLOW}Check status with: kubectl get ksvc nodejs-app${NC}"
