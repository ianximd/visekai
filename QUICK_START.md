# Quick Start Guide

## 🎯 Get Running in 5 Minutes

### Step 1: Environment Setup (1 min)

```bash
# Copy environment template
cp .env.example .env

# Edit the .env file - REQUIRED CHANGES:
# 1. Set POSTGRES_PASSWORD to a secure password
# 2. Set JWT_SECRET to a random 32+ character string
nano .env
```

**Minimum Required Changes in `.env`:**
```bash
POSTGRES_PASSWORD=your_secure_password_here_123456
JWT_SECRET=your_jwt_secret_must_be_at_least_32_chars_long_xyz
```

### Step 2: Start Services (2 min)

```bash
# Start all services
make up

# Or without make:
docker-compose up -d
```

**Expected output:**
```
Creating network "visekai_ocr_network" ...
Creating ocr_postgres ... done
Creating ocr_redis    ... done
Creating ocr_service  ... done
Creating ocr_backend  ... done
Creating ocr_frontend ... done
```

### Step 3: Verify Services (1 min)

```bash
# Check all services are running
make ps

# Or:
docker-compose ps
```

**All services should show "Up" status:**
- ocr_postgres (port 5432)
- ocr_redis (port 6379)
- ocr_backend (port 8080)
- ocr_service (port 8000)
- ocr_frontend (port 3000)

### Step 4: Access the Application (1 min)

Open your browser:

- **Frontend**: http://localhost:3000
- **Backend API Health**: http://localhost:8080/api/v1/health
- **OCR Service Health**: http://localhost:8000/health

### Step 5: View Logs

```bash
# All services
make logs

# Specific service
make logs-backend
make logs-frontend
make logs-ocr
```

---

## 🚨 Troubleshooting

### Services Won't Start

**Check Docker:**
```bash
docker --version
docker-compose --version
```

**Check Ports:**
```bash
# Make sure these ports are free:
netstat -tuln | grep -E '3000|8080|8000|5432|6379'
```

**Reset Everything:**
```bash
make down
make clean
make up
```

### Database Connection Issues

```bash
# Check PostgreSQL is running
docker-compose logs postgres

# Connect to database manually
docker-compose exec postgres psql -U ocr_user -d ocr_db
```

### GPU Not Available (OCR Service)

For development without GPU, comment out the GPU section in `docker-compose.yml`:

```yaml
# Comment these lines:
# deploy:
#   resources:
#     reservations:
#       devices:
#         - driver: nvidia
#           count: 1
#           capabilities: [gpu]
```

---

## 🔧 Development Mode

### Backend Development

```bash
cd backend
go mod download
go run cmd/server/main.go
```

**Test the API:**
```bash
curl http://localhost:8080/api/v1/health
```

### Frontend Development

```bash
cd frontend
npm install
npm run dev
```

**Access**: http://localhost:3000

### OCR Service Development

```bash
cd ocr-service
pip install -r requirements.txt
python main.py
```

**Test the API:**
```bash
curl http://localhost:8000/health
```

---

## 📚 Next Steps

1. **Read Documentation**:
   - `PROJECT_PLAN.md` - Complete architecture
   - `SETUP_COMPLETE.md` - Implementation roadmap
   - `README.md` - Full documentation

2. **Implement Features** (in order):
   - Authentication (login/register)
   - File upload
   - OCR processing
   - Results display

3. **Run Tests**:
   ```bash
   make test
   ```

---

## 🎓 Learning the Codebase

### Key Files to Start With:

**Backend:**
- `backend/cmd/server/main.go` - Entry point
- `backend/internal/handlers/handlers.go` - API endpoints
- `backend/internal/config/config.go` - Configuration

**Frontend:**
- `frontend/src/main.js` - Entry point
- `frontend/src/App.vue` - Root component
- `frontend/src/router/index.js` - Routes
- `frontend/src/views/Home.vue` - Home page

**OCR Service:**
- `ocr-service/main.py` - FastAPI app
- `ocr-service/api/routes.py` - API endpoints
- `ocr-service/deepseek_ocr/model.py` - OCR model wrapper

---

## 💡 Useful Commands

```bash
# Help
make help

# Start/Stop
make up
make down
make restart

# Logs
make logs
make logs-backend

# Clean up
make clean

# Database
make shell-db

# Backend shell
make shell-backend
```

---

## ✅ Verification Checklist

- [ ] Docker and Docker Compose installed
- [ ] `.env` file created and configured
- [ ] All services started successfully
- [ ] Can access frontend at localhost:3000
- [ ] Backend health check returns 200
- [ ] OCR service health check returns 200
- [ ] Database is accessible
- [ ] Redis is running

Run the verification script:
```bash
./verify-setup.sh
```

---

## 🆘 Need Help?

1. Check logs: `make logs`
2. Review `SETUP_COMPLETE.md`
3. Check GitHub issues
4. Consult `PROJECT_PLAN.md` for architecture details

---

**You're all set! Happy coding! 🚀**
