#!/bin/bash

# Variables
AWS_REGION="us-east-1"
REPOSITORY_NAME="cherubgyre-dev"
IMAGE_TAG="latest"
CLUSTER_NAME="cherubgyre-dev"
TASK_DEFINITION_FAMILY="cherubgyre-dev"
SUBNETS='["subnet-0ac341ecb24bee027", "subnet-0c0d9667c504aa776", "subnet-002bf7f9fbe43da9a", "subnet-06f2c8c9315a207f8", "subnet-0101c89782bed53be", "subnet-0f52ad3011c093e0c"]'
SECURITY_GROUP="sg-0001d3ac435b86000"

echo "Checking if repository exists..."
EXISTING_REPO=$(aws ecr describe-repositories --repository-names $REPOSITORY_NAME --region $AWS_REGION --query 'repositories[0].repositoryUri' --output text)

if [ "$EXISTING_REPO" == "None" ]; then
    echo "Repository does not exist. Creating repository..."
    REPOSITORY_URI=$(aws ecr create-repository --repository-name $REPOSITORY_NAME --region $AWS_REGION --query 'repository.repositoryUri' --output text)
else
    echo "Repository already exists: $EXISTING_REPO"
    REPOSITORY_URI=$EXISTING_REPO
fi


echo "Repository URI: $REPOSITORY_URI"

# Authenticate Docker with ECR
echo "Authenticating Docker with ECR..."
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $REPOSITORY_URI

# Clone the repository
echo "Cloning the repository..."
git clone https://github.com/davidemerson/cherubgyre
cd cherubgyre || exit

# Build Docker Image
echo "Building Docker image..."
docker build -t $REPOSITORY_NAME .

# Tag Docker Image
echo "Tagging Docker image..."
docker tag $REPOSITORY_NAME:$IMAGE_TAG $REPOSITORY_URI:$IMAGE_TAG

# Push Docker Image to ECR
echo "Pushing Docker image to ECR..."
docker push $REPOSITORY_URI:$IMAGE_TAG

# Go back to initial directory
cd ..

# Create ECS Cluster
echo "Creating ECS cluster..."
aws ecs create-cluster --cluster-name $CLUSTER_NAME --region $AWS_REGION

# Register Task Definition
echo "Registering ECS Task Definition..."
cat <<EOF > task-definition.json
{
    "containerDefinitions": [
        {
            "name": "$REPOSITORY_NAME",
            "image": "$REPOSITORY_URI:$IMAGE_TAG",
            "cpu": 0,
            "portMappings": [
                {
                    "name": "rust-api-container-8080-tcp",
                    "containerPort": 8080,
                    "hostPort": 8080,
                    "protocol": "tcp",
                    "appProtocol": "http"
                }
            ],
            "essential": true,
            "logConfiguration": {
                "logDriver": "awslogs",
                "options": {
                    "awslogs-group": "/ecs/$CLUSTER_NAME",
                    "awslogs-region": "$AWS_REGION",
                    "awslogs-create-group": "true",
                    "awslogs-stream-prefix": "ecs"
                }
            }
        }
    ],
    "family": "$TASK_DEFINITION_FAMILY",
    "executionRoleArn": "arn:aws:iam::[youraccount]:role/ecsTaskExecutionRole",
    "networkMode": "awsvpc",
    "requiresCompatibilities": [
        "FARGATE"
    ],
    "cpu": "512",
    "memory": "1024",
    "runtimePlatform": {
        "cpuArchitecture": "X86_64",
        "operatingSystemFamily": "LINUX"
    }
}
EOF

aws ecs register-task-definition --cli-input-json file://task-definition.json

# Run Task
echo "Running ECS Task..."
TASK_ARN=$(aws ecs run-task \
    --cluster $CLUSTER_NAME \
    --launch-type FARGATE \
    --task-definition $TASK_DEFINITION_FAMILY \
    --network-configuration "awsvpcConfiguration={subnets=$SUBNETS,securityGroups=[$SECURITY_GROUP],assignPublicIp='ENABLED'}" \
    --query 'tasks[0].taskArn' --output text)

echo "Task ARN: $TASK_ARN"

# Wait for Task to Start
echo "Waiting for task to start..."
aws ecs wait tasks-running --cluster $CLUSTER_NAME --tasks $TASK_ARN

# Get Task Details
TASK_DETAILS=$(aws ecs describe-tasks --cluster $CLUSTER_NAME --tasks $TASK_ARN --query 'tasks[0].attachments[0].details')
PUBLIC_IP=$(echo $TASK_DETAILS | jq -r '.[] | select(.name=="networkInterfaceId") | .value' | xargs -I{} aws ec2 describe-network-interfaces --network-interface-ids {} --query 'NetworkInterfaces[0].Association.PublicIp' --output text)

echo "API is now running at http://$PUBLIC_IP:8080"