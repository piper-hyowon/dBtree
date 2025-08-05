#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m'

NAMESPACE="user-demo-user-456"
INSTANCE_NAME="demo-mongodb"
DB_NAME="testdb"
USERNAME="admin"
PASSWORD="demopass789"
NODE_PORT=30100

echo "리소스 정리..."
kubectl delete dbinstance -n $NAMESPACE $INSTANCE_NAME 2>/dev/null
kubectl delete all --all -n $NAMESPACE 2>/dev/null
kubectl delete pvc --all -n $NAMESPACE 2>/dev/null
kubectl delete configmap --all -n $NAMESPACE 2>/dev/null
kubectl delete secret --all -n $NAMESPACE 2>/dev/null

sleep 5
kubectl create namespace $NAMESPACE 2>/dev/null || true

echo "Secret 생성..."
kubectl apply -f new-secret.yaml

echo "NodePort Service 생성..."
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: ${INSTANCE_NAME}-external
  namespace: ${NAMESPACE}
spec:
  type: NodePort
  selector:
    app: ${INSTANCE_NAME}
    dbtree.cloud/instance-id: demo-external-id-123
  ports:
  - port: 27017
    targetPort: 27017
    nodePort: ${NODE_PORT}
EOF

echo "DBInstance 생성..."
kubectl apply -f demo-dbinstance.yaml

echo "StatefulSet 생성 대기..."
sleep 10

# StatefulSet이 생성될 때까지 대기
for i in {1..30}; do
    if kubectl get statefulset ${INSTANCE_NAME}-sts -n $NAMESPACE 2>/dev/null; then
        echo -e "${GREEN}StatefulSet 생성됨${NC}"
        break
    fi
    echo -n "."
    sleep 5
done

echo "Pod 생성 대기..."
# Pod이 생성될 때까지 대기
for i in {1..30}; do
    POD_COUNT=$(kubectl get pods -n $NAMESPACE -l app=$INSTANCE_NAME --no-headers 2>/dev/null | wc -l)
    if [ "$POD_COUNT" -gt 0 ]; then
        echo -e "${GREEN}Pod 생성됨${NC}"
        break
    fi
    echo -n "."
    sleep 5
done

# Pod 이름 가져오기
POD_NAME="${INSTANCE_NAME}-sts-0"

echo "Pod Ready 대기..."
kubectl wait --for=condition=ready pod/$POD_NAME -n $NAMESPACE --timeout=300s

echo "내부 접속 테스트..."
kubectl exec -n $NAMESPACE $POD_NAME -- mongosh -u $USERNAME -p $PASSWORD --authenticationDatabase admin --eval "db.adminCommand({ping: 1})"

echo "데이터 삽입..."
kubectl exec -n $NAMESPACE $POD_NAME -- mongosh -u $USERNAME -p $PASSWORD --authenticationDatabase admin --eval "use $DB_NAME; db.users.insertOne({name: 'Test User'}); db.users.find();"

echo "외부 접속 테스트..."
mongosh "mongodb://$USERNAME:$PASSWORD@localhost:$NODE_PORT/$DB_NAME?authSource=admin" --eval "db.users.find()"

echo -e "\n리소스 상태:"
kubectl get all -n $NAMESPACE

echo -e "${GREEN}테스트 완료!${NC}"