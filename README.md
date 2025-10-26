# VisEkai - OCR Web Application

A full-stack web application for reading scanned papers (printed or handwritten) using DeepSeek-OCR engine.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go](https://img.shields.io/badge/Go-1.21-blue.svg)
![Vue](https://img.shields.io/badge/Vue.js-3.4-green.svg)
![Python](https://img.shields.io/badge/Python-3.11-yellow.svg)

## Features

- 🚀 **High-Performance OCR** - Powered by DeepSeek-OCR with vLLM
- 📄 **Multiple Formats** - Support for images (JPG, PNG, TIFF) and PDFs
- ✍️ **Handwriting Recognition** - Advanced handwritten text recognition
- 📊 **Structured Output** - Extract tables, figures, and formatted text
- 🔒 **Secure** - JWT authentication, file encryption, and access control
- 📈 **Scalable** - Microservices architecture with Docker
- 🎨 **Modern UI** - Responsive Vue.js interface with Element Plus

## Tech Stack

### Frontend
- Vue.js 3 + Vite
- Element Plus UI
- Pinia (State Management)
- Axios (HTTP Client)

### Backend
- Go (Golang) with Gin framework
- PostgreSQL with pgx driver
- Redis (Caching & Job Queue)
- JWT Authentication

### OCR Engine
- DeepSeek-OCR
- vLLM for inference
- FastAPI service wrapper

### Infrastructure
- Docker + Docker Compose
- Nginx (Reverse Proxy)
- Optional: Prometheus + Grafana

## Project Structure

```
visekai/
├── backend/          # Go backend service
├── frontend/         # Vue.js frontend
├── ocr-service/      # Python OCR service
├── database/         # Database migrations
├── storage/          # File storage
├── nginx/            # Nginx configuration
├── docker-compose.yml
├── Makefile
└── PROJECT_PLAN.md   # Detailed project plan
```

## Quick Start

### Prerequisites

- Docker & Docker Compose (v20.10+)
- NVIDIA GPU with CUDA support
- NVIDIA Docker Runtime (nvidia-docker2)
- 16GB+ RAM (32GB recommended)
- 50GB+ disk space

### Setup

1. **Clone the repository**:
```bash
git clone https://github.com/ianximd/visekai.git
cd visekai
```

2. **Configure environment**:
```bash
cp .env.example .env
# Edit .env with your settings
nano .env
```

3. **Start services**:
```bash
make up
# or
docker-compose up -d
```

4. **Check service status**:
```bash
make ps
```

5. **View logs**:
```bash
make logs
```

### Access the Application

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080/api/v1
- **OCR Service**: http://localhost:8000
- **API Health Check**: http://localhost:8080/api/v1/health

## Development

### Backend Development

```bash
cd backend
go mod download
go run cmd/server/main.go
```

### Frontend Development

```bash
cd frontend
npm install
npm run dev
```

### OCR Service Development

```bash
cd ocr-service
pip install -r requirements.txt
python main.py
```

## Available Commands

```bash
make help          # Show all available commands
make up            # Start all services
make down          # Stop all services
make build         # Build all services
make restart       # Restart all services
make logs          # View all logs
make clean         # Clean up everything
make test          # Run all tests
```

## Testing

```bash
# Run all tests
make test

# Backend tests
cd backend && go test -v ./...

# Frontend tests
cd frontend && npm run test
```

## API Documentation

See [PROJECT_PLAN.md](PROJECT_PLAN.md) for complete API documentation.

### Example API Calls

**Upload Document**:
```bash
curl -X POST http://localhost:8080/api/v1/documents/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@document.pdf"
```

**Submit OCR Job**:
```bash
curl -X POST http://localhost:8080/api/v1/ocr/submit \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "document_id": "doc-uuid",
    "mode": "document",
    "resolution": "base"
  }'
```

## Configuration

### Environment Variables

Key environment variables (see `.env.example` for full list):

```bash
# Database
POSTGRES_PASSWORD=your_secure_password

# JWT
JWT_SECRET=your_jwt_secret_key_min_32_chars

# OCR Settings
MODEL_PATH=deepseek-ai/DeepSeek-OCR
BASE_SIZE=1024
IMAGE_SIZE=640
CROP_MODE=true
```

## Deployment

### Production Setup

1. Update environment variables for production
2. Configure SSL certificates in `nginx/ssl/`
3. Build and start with production profile:

```bash
make prod-build
make prod-up
```

### Using Nginx Reverse Proxy

```bash
docker-compose --profile production up -d
```

## Monitoring

### Health Checks

```bash
# Backend health
curl http://localhost:8080/api/v1/health

# OCR service health
curl http://localhost:8000/health

# Database connection
docker-compose exec postgres pg_isready
```

### View Metrics

```bash
# Service status
docker-compose ps

# Resource usage
docker stats

# Logs
docker-compose logs -f [service-name]
```

## Troubleshooting

### GPU Not Detected

```bash
# Verify NVIDIA runtime
docker run --rm --gpus all nvidia/cuda:12.0-base nvidia-smi
```

### Database Connection Issues

```bash
# Check PostgreSQL logs
docker-compose logs postgres

# Connect to database
docker-compose exec postgres psql -U ocr_user -d ocr_db
```

### Port Conflicts

Check if ports 3000, 8080, 8000, 5432, 6379 are available:
```bash
netstat -tuln | grep -E '3000|8080|8000|5432|6379'
```

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [DeepSeek-OCR](https://github.com/deepseek-ai/DeepSeek-OCR) - OCR engine
- [vLLM](https://github.com/vllm-project/vllm) - LLM inference engine
- [Gin](https://github.com/gin-gonic/gin) - Go web framework
- [Vue.js](https://vuejs.org/) - Frontend framework
- [Element Plus](https://element-plus.org/) - UI library

## Contact

- Project: [https://github.com/ianximd/visekai](https://github.com/ianximd/visekai)
- Issues: [https://github.com/ianximd/visekai/issues](https://github.com/ianximd/visekai/issues)

## Roadmap

See [PROJECT_PLAN.md](PROJECT_PLAN.md) for detailed development roadmap and future enhancements.

---

Made with ❤️ by the VisEkai Team
