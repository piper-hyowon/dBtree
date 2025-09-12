
## Prerequisites
- 환경: 맥북 M1 (local), EC2 Ubuntu Linux (production)
- EC2에 PostgreSQL, Redis 설치 완료
- kubectl 설정 완료

## Build & Push Images
### Backend
```bash
docker buildx build --platform linux/amd64 \
-t[ID]/dbtree-backend:v[버전] \
--push .
```


### Operator
```bash
make docker-build docker-push IMG=[ID]/dbtree-operator:v[버전]
```

해당 버전으로 deployment yaml 각 backend, operator 반영


1. 로컬에서 CRD 파일을 EC2로 복사 및 적용
```bash
# Local에서
scp -i [pem key] \
operator/config/crd/bases/dbtree.cloud_dbinstances.yaml \
ubuntu@[EC2-IP]:~/

# EC2에서
kubectl apply -f ~/dbtree.cloud_dbinstances.yaml
```

2. Secrets 설정
```bash
cp backend-secrets.example.yaml backend-secrets.yaml
# 실제 값으로 수정
kubectl apply -f backend-secrets.yaml
```

3. 배포
```bash
# BE
kubectl apply -f backend-rbac.yaml 
kubectl apply -f backend-deployment.yaml
kubectl apply -f backend-service.yaml 

# Operator
kubectl apply -f  operator-rbac.yaml
kubectl apply -f operator-deployment.yaml
```

#### Update & Rollout
이미지 업데이트 후:
```bash
kubectl rollout restart deployment/backend
kubectl rollout restart deployment/dbtree-operator
```

#### Monitoring 
```bash
kubectl get pods
kubectl logs -f deployment/backend
kubectl logs -f deployment/dbtree-operator
```
