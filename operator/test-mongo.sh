#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

NAMESPACE="user-demo-user-456"
INSTANCE_NAME="demo-mongodb"

echo "리소스 정리..."
kubectl delete dbinstance -n $NAMESPACE $INSTANCE_NAME 2>/dev/null
kubectl delete all --all -n $NAMESPACE 2>/dev/null
kubectl delete pvc --all -n $NAMESPACE 2>/dev/null
kubectl delete configmap --all -n $NAMESPACE 2>/dev/null
kubectl delete secret --all -n $NAMESPACE 2>/dev/null

sleep 5

echo "Secret 생성..."
kubectl apply -f new-secret.yaml

echo "DBInstance 생성..."
kubectl apply -f demo-dbinstance.yaml

echo "Pod Ready 대기"
kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=$INSTANCE_NAME -n $NAMESPACE --timeout=60s

echo "초기화 로그 확인"
kubectl logs -n $NAMESPACE $INSTANCE_NAME-sts-0 | grep -E "(First run|User created|Initialization)"

echo "authentication..."
if kubectl exec -n $NAMESPACE $INSTANCE_NAME-sts-0 -- mongosh -u admin -p demopass789 --eval "db.adminCommand({ping: 1})" &>/dev/null; then
    echo -e "${GREEN}✅ Authentication successful!${NC}"

    echo "🧪 DB 테스트"
    kubectl exec -n $NAMESPACE $INSTANCE_NAME-sts-0 -- mongosh -u admin -p demopass789 --eval "
        use testdb;
        db.users.insertOne({name: 'Test User', created: new Date()});
        print('Insert successful');
        print('Users in DB:', db.users.countDocuments());
    "
else
    echo -e "${RED}❌ Authentication failed${NC}"

    echo "유저 목록 확인ㅇ"
    kubectl exec -n $NAMESPACE $INSTANCE_NAME-sts-0 -- mongosh --eval "use admin; db.getUsers()"
fi

echo "테스트 완료"