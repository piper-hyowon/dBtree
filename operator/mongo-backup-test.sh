#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

NAMESPACE="user-demo-user-456"
INSTANCE_NAME="demo-mongodb"

echo -e "${YELLOW}=== Backup Debug ===${NC}"

# 1. 현재 리소스 상태 확인
echo -e "\n${YELLOW}1. Current resources:${NC}"
echo "CronJobs:"
kubectl get cronjob -n $NAMESPACE
echo -e "\nJobs:"
kubectl get jobs -n $NAMESPACE
echo -e "\nPVCs:"
kubectl get pvc -n $NAMESPACE | grep backup

# 2. CronJob 상세 정보
echo -e "\n${YELLOW}2. CronJob details:${NC}"
kubectl get cronjob $INSTANCE_NAME-backup -n $NAMESPACE -o yaml | grep -A20 "volumeMounts:" || echo "volumeMounts not found"

# 3. 새로운 백업 Job 생성 및 모니터링
echo -e "\n${YELLOW}3. Creating new backup job...${NC}"
JOB_NAME="debug-backup-$(date +%s)"
kubectl create job --from=cronjob/$INSTANCE_NAME-backup $JOB_NAME -n $NAMESPACE

# 실시간 모니터링
echo -e "\n${YELLOW}4. Monitoring job...${NC}"
for i in {1..30}; do
    STATUS=$(kubectl get job $JOB_NAME -n $NAMESPACE -o jsonpath='{.status.conditions[0].type}' 2>/dev/null)
    ACTIVE=$(kubectl get job $JOB_NAME -n $NAMESPACE -o jsonpath='{.status.active}' 2>/dev/null)
    SUCCEEDED=$(kubectl get job $JOB_NAME -n $NAMESPACE -o jsonpath='{.status.succeeded}' 2>/dev/null)
    FAILED=$(kubectl get job $JOB_NAME -n $NAMESPACE -o jsonpath='{.status.failed}' 2>/dev/null)

    echo -ne "\rStatus: $STATUS | Active: ${ACTIVE:-0} | Succeeded: ${SUCCEEDED:-0} | Failed: ${FAILED:-0}    "

    if [[ "$SUCCEEDED" == "1" ]]; then
        echo -e "\n${GREEN}✅ Job completed successfully!${NC}"
        break
    elif [[ "$FAILED" == "1" ]]; then
        echo -e "\n${RED}❌ Job failed!${NC}"
        break
    fi

    sleep 2
done

# 5. Job 로그 확인
echo -e "\n\n${YELLOW}5. Job logs:${NC}"
kubectl logs job/$JOB_NAME -n $NAMESPACE 2>/dev/null || echo "No logs available"

# 6. Pod 상태 확인
echo -e "\n${YELLOW}6. Pod details:${NC}"
POD_NAME=$(kubectl get pods -n $NAMESPACE -l job-name=$JOB_NAME -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
if [ -n "$POD_NAME" ]; then
    echo "Pod: $POD_NAME"
    kubectl get pod $POD_NAME -n $NAMESPACE -o wide

    # Pod events
    echo -e "\nPod Events:"
    kubectl describe pod $POD_NAME -n $NAMESPACE | tail -20
else
    echo "No pod found for job $JOB_NAME"
fi

# 7. 이전 성공한 백업 찾기
echo -e "\n${YELLOW}7. Previous successful backups:${NC}"
kubectl get jobs -n $NAMESPACE -o json | jq -r '.items[] | select(.status.succeeded==1) | .metadata.name' | tail -5

# Cleanup
echo -e "\n${YELLOW}Cleanup (press Enter to delete test job, Ctrl+C to keep):${NC}"
read
kubectl delete job $JOB_NAME -n $NAMESPACE --ignore-not-found=true