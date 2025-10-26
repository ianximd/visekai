# OCR Web Application - Project Plan

> **Last Updated**: October 26, 2025  
> **Version**: 2.0 (Enhanced)

## Project Overview

A full-stack web application for reading scanned papers (printed or handwritten) using DeepSeek-OCR engine.

### Tech Stack
- **Frontend**: Vue.js 3 + Vite
- **Backend**: Go (Golang) with Gin framework
- **OCR Engine**: DeepSeek-OCR (Python service using vLLM)
- **Database**: PostgreSQL with pgx driver
- **File Storage**: Local filesystem (with future cloud storage support)
- **Communication**: REST API
- **Containerization**: Docker + Docker Compose

---

## System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client (Browser)                         │
│                          Vue.js SPA                              │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTP/REST API
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│                      Go Backend Service                          │
│  - API Gateway                                                   │
│  - File Upload Handler                                           │
│  - Job Queue Manager                                             │
│  - Database Operations                                           │
└──────────┬───────────────────────────────────┬──────────────────┘
           │                                   │
           │ gRPC/HTTP                         │ SQL
           │                                   │
┌──────────▼──────────────┐       ┌───────────▼─────────────┐
│  DeepSeek-OCR Service   │       │   PostgreSQL Database   │
│  (Python + vLLM)        │       │   - Users               │
│  - Image Processing     │       │   - Documents           │
│  - OCR Inference        │       │   - OCR Jobs            │
│  - Result Generation    │       │   - Results             │
└─────────────────────────┘       └─────────────────────────┘
           │
           │
┌──────────▼──────────────┐
│   File Storage          │
│   - Uploaded Images     │
│   - Processed Results   │
│   - Thumbnails          │
└─────────────────────────┘
```

---

## Database Schema

### Tables

#### 1. users
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 2. documents
```sql
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100),
    file_hash VARCHAR(64), -- SHA-256 hash for deduplication
    num_pages INTEGER DEFAULT 1, -- for PDFs
    thumbnail_path VARCHAR(500), -- for preview
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP -- soft delete support
);

CREATE INDEX idx_documents_user_id ON documents(user_id);
CREATE INDEX idx_documents_file_hash ON documents(file_hash);
```

#### 3. ocr_jobs
```sql
CREATE TABLE ocr_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL, -- pending, processing, completed, failed, cancelled
    ocr_mode VARCHAR(50) DEFAULT 'document', -- document, handwritten, general, figure
    resolution_mode VARCHAR(50) DEFAULT 'base', -- tiny, small, base, large, gundam
    priority INTEGER DEFAULT 0,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    progress_percentage INTEGER DEFAULT 0, -- 0-100
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    metadata JSONB -- for additional job-specific data
);

CREATE INDEX idx_ocr_jobs_status ON ocr_jobs(status);
CREATE INDEX idx_ocr_jobs_user_id ON ocr_jobs(user_id);
CREATE INDEX idx_ocr_jobs_created_at ON ocr_jobs(created_at DESC);
```

#### 4. ocr_results
```sql
CREATE TABLE ocr_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID REFERENCES ocr_jobs(id) ON DELETE CASCADE,
    document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
    raw_text TEXT,
    markdown_text TEXT,
    json_data JSONB, -- for structured data (tables, figures, etc.)
    confidence_score FLOAT,
    processing_time_ms INTEGER,
    num_pages INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ocr_results_job_id ON ocr_results(job_id);
CREATE INDEX idx_ocr_results_document_id ON ocr_results(document_id);
```

#### 5. job_logs
```sql
CREATE TABLE job_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID REFERENCES ocr_jobs(id) ON DELETE CASCADE,
    level VARCHAR(20), -- info, warning, error
    message TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_job_logs_job_id ON job_logs(job_id);
CREATE INDEX idx_job_logs_created_at ON job_logs(created_at DESC);
```

#### 6. user_settings (NEW)
```sql
CREATE TABLE user_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    default_ocr_mode VARCHAR(50) DEFAULT 'document',
    default_resolution VARCHAR(50) DEFAULT 'base',
    email_notifications BOOLEAN DEFAULT true,
    language VARCHAR(10) DEFAULT 'en',
    theme VARCHAR(20) DEFAULT 'light',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 7. api_keys (NEW - for programmatic access)
```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(64) NOT NULL, -- hashed API key
    name VARCHAR(255) NOT NULL,
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
```

---

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout
- `POST /api/v1/auth/refresh` - Refresh access token
- `GET /api/v1/auth/me` - Get current user info
- `PUT /api/v1/auth/profile` - Update user profile
- `PUT /api/v1/auth/password` - Change password

### Documents
- `POST /api/v1/documents/upload` - Upload document(s) (supports multipart/form-data)
- `GET /api/v1/documents` - List user's documents (with pagination, filtering)
- `GET /api/v1/documents/:id` - Get document details
- `GET /api/v1/documents/:id/thumbnail` - Get document thumbnail
- `DELETE /api/v1/documents/:id` - Delete document (soft delete)

### OCR Jobs
- `POST /api/v1/ocr/submit` - Submit OCR job
- `POST /api/v1/ocr/batch` - Submit batch OCR jobs
- `GET /api/v1/ocr/jobs` - List user's jobs (with pagination, status filter)
- `GET /api/v1/ocr/jobs/:id` - Get job status and progress
- `GET /api/v1/ocr/jobs/:id/result` - Get OCR result
- `PUT /api/v1/ocr/jobs/:id/cancel` - Cancel pending job
- `DELETE /api/v1/ocr/jobs/:id` - Delete job and results
- `GET /api/v1/ocr/jobs/:id/logs` - Get job processing logs

### Results
- `GET /api/v1/results/:id` - Get result details
- `GET /api/v1/results/:id/download` - Download result (markdown/json/txt)
- `GET /api/v1/results/:id/preview` - Get result preview
- `POST /api/v1/results/:id/export` - Export to different format (PDF, DOCX)

### User Settings
- `GET /api/v1/settings` - Get user settings
- `PUT /api/v1/settings` - Update user settings

### API Keys (for developers)
- `GET /api/v1/api-keys` - List user's API keys
- `POST /api/v1/api-keys` - Create new API key
- `DELETE /api/v1/api-keys/:id` - Revoke API key

### Statistics & Analytics
- `GET /api/v1/stats/usage` - Get usage statistics
- `GET /api/v1/stats/jobs` - Get job statistics

### Health & Monitoring
- `GET /api/v1/health` - Health check
- `GET /api/v1/status` - System status
- `GET /api/v1/metrics` - System metrics (admin only)

---

## Project Structure

```
visekai/
├── docker-compose.yml
├── .env.example
├── README.md
│
├── frontend/                      # Vue.js Frontend
│   ├── Dockerfile
│   ├── package.json
│   ├── vite.config.js
│   ├── index.html
│   ├── public/
│   └── src/
│       ├── main.js
│       ├── App.vue
│       ├── router/
│       ├── stores/               # Pinia state management
│       ├── components/
│       │   ├── UploadArea.vue
│       │   ├── DocumentList.vue
│       │   ├── JobStatus.vue
│       │   └── ResultViewer.vue
│       ├── views/
│       │   ├── Home.vue
│       │   ├── Upload.vue
│       │   ├── Documents.vue
│       │   ├── Jobs.vue
│       │   └── Results.vue
│       ├── services/
│       │   └── api.js
│       └── assets/
│
├── backend/                       # Go Backend
│   ├── Dockerfile
│   ├── .dockerignore
│   ├── go.mod
│   ├── go.sum
│   ├── Makefile                   # Build and dev commands
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go
│   │   ├── database/
│   │   │   ├── postgres.go
│   │   │   ├── migrations/
│   │   │   └── seed.go
│   │   ├── models/
│   │   │   ├── user.go
│   │   │   ├── document.go
│   │   │   ├── job.go
│   │   │   ├── result.go
│   │   │   └── api_key.go
│   │   ├── handlers/
│   │   │   ├── auth.go
│   │   │   ├── document.go
│   │   │   ├── ocr.go
│   │   │   ├── result.go
│   │   │   ├── settings.go
│   │   │   └── health.go
│   │   ├── services/
│   │   │   ├── auth_service.go
│   │   │   ├── document_service.go
│   │   │   ├── ocr_service.go
│   │   │   ├── storage_service.go
│   │   │   ├── queue_service.go   # Job queue management
│   │   │   └── notification_service.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── logger.go
│   │   │   ├── rate_limit.go
│   │   │   └── error_handler.go
│   │   ├── client/
│   │   │   └── ocr_client.go      # Client for OCR service
│   │   └── repository/
│   │       ├── user_repo.go
│   │       ├── document_repo.go
│   │       ├── job_repo.go
│   │       └── result_repo.go
│   └── pkg/
│       ├── logger/
│       ├── validator/
│       ├── utils/
│       └── errors/
│
├── ocr-service/                   # DeepSeek-OCR Service
│   ├── Dockerfile
│   ├── .dockerignore
│   ├── requirements.txt
│   ├── config.py
│   ├── main.py                    # FastAPI server
│   ├── deepseek_ocr/              # DeepSeek-OCR integration
│   │   ├── __init__.py
│   │   ├── model.py
│   │   ├── processor.py
│   │   ├── inference.py
│   │   └── utils.py
│   ├── api/
│   │   ├── __init__.py
│   │   ├── routes.py
│   │   ├── schemas.py
│   │   └── dependencies.py
│   ├── core/
│   │   ├── config.py
│   │   ├── logging.py
│   │   └── queue.py              # Job queue management
│   ├── utils/
│   │   ├── image_processing.py
│   │   ├── pdf_processing.py
│   │   └── validation.py
│   └── tests/
│       ├── test_api.py
│       └── test_inference.py
│
├── database/                      # Database setup
│   ├── migrations/
│   │   ├── 001_init_schema.sql
│   │   ├── 002_add_indexes.sql
│   │   └── ...
│   └── seed/
│       └── test_data.sql
│
└── storage/                       # Local file storage (mounted volume)
    ├── uploads/
    ├── results/
    └── temp/
```

---

## API Response Standards

### Success Response Format
```json
{
  "success": true,
  "data": {
    // Response data
  },
  "message": "Operation completed successfully",
  "timestamp": "2025-10-26T10:30:00Z"
}
```

### Error Response Format
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input parameters",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ]
  },
  "timestamp": "2025-10-26T10:30:00Z"
}
```

### HTTP Status Codes
- `200 OK`: Successful GET, PUT requests
- `201 Created`: Successful POST request
- `204 No Content`: Successful DELETE request
- `400 Bad Request`: Invalid input/validation error
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict (duplicate)
- `413 Payload Too Large`: File size exceeds limit
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Service temporarily unavailable

### Error Codes
- `AUTH_001`: Invalid credentials
- `AUTH_002`: Token expired
- `AUTH_003`: Insufficient permissions
- `VAL_001`: Validation error
- `VAL_002`: Invalid file type
- `VAL_003`: File too large
- `RES_001`: Resource not found
- `RES_002`: Resource already exists
- `OCR_001`: OCR processing failed
- `OCR_002`: Invalid image format
- `SYS_001`: Database error
- `SYS_002`: Storage error
- `SYS_003`: External service error

### Pagination Response
```json
{
  "success": true,
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 100,
      "total_pages": 5,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

---

## DeepSeek-OCR Integration

### OCR Service Setup

The OCR service will use DeepSeek-OCR with vLLM for high-performance inference.

#### Configuration Options:
- **Resolution Modes**:
  - `tiny`: 512×512 (64 vision tokens) - fastest
  - `small`: 640×640 (100 vision tokens) - balanced
  - `base`: 1024×1024 (256 vision tokens) - good quality
  - `large`: 1280×1280 (400 vision tokens) - high quality
  - `gundam`: n×640×640 + 1×1024×1024 - dynamic resolution

- **Prompts**:
  - Document: `<image>\n<|grounding|>Convert the document to markdown.`
  - Handwritten: `<image>\n<|grounding|>OCR this image.`
  - Free OCR: `<image>\nFree OCR.`
  - Figures: `<image>\nParse the figure.`

#### API Interface (REST):
```python
# POST /ocr/process
{
    "image_path": "/path/to/image.jpg",
    "mode": "document",  # document, handwritten, general
    "resolution": "base",  # tiny, small, base, large, gundam
    "output_format": "markdown"  # markdown, json, text
}

# Response
{
    "job_id": "uuid",
    "status": "completed",
    "result": {
        "text": "...",
        "markdown": "...",
        "confidence": 0.95,
        "processing_time_ms": 1500
    }
}
```

---

## Development Roadmap

### Phase 1: Project Setup & Infrastructure (Week 1-2)
- [ ] Initialize project repository structure
- [ ] Set up Docker Compose configuration
- [ ] Configure PostgreSQL with initial schema
- [ ] Set up DeepSeek-OCR service container
- [ ] Create basic Go backend with Gin framework
- [ ] Initialize Vue.js frontend with Vite
- [ ] Set up development environment

### Phase 2: Core Backend Development (Week 3-4)
- [ ] Implement database models and migrations
- [ ] Create authentication system (JWT)
- [ ] Implement file upload handler
- [ ] Develop document management API
- [ ] Create OCR job queue system
- [ ] Implement OCR service client
- [ ] Add basic error handling and logging

### Phase 3: OCR Service Integration (Week 5)
- [ ] Set up DeepSeek-OCR with vLLM
- [ ] Create FastAPI wrapper for OCR service
- [ ] Implement image preprocessing pipeline
- [ ] Add support for multiple resolution modes
- [ ] Test OCR accuracy with sample documents
- [ ] Optimize inference performance

### Phase 4: Frontend Development (Week 6-7)
- [ ] Create landing page and layout
- [ ] Implement file upload interface
- [ ] Build document management view
- [ ] Create job status dashboard
- [ ] Implement result viewer with markdown preview
- [ ] Add authentication UI
- [ ] Implement state management with Pinia

### Phase 5: Integration & Testing (Week 8)
- [ ] End-to-end integration testing
- [ ] API testing with Postman/Thunder Client
- [ ] Frontend-backend integration
- [ ] OCR accuracy testing
- [ ] Performance testing and optimization
- [ ] Security testing

### Phase 6: Advanced Features (Week 9-10)
- [ ] Batch processing support
- [ ] PDF to images conversion
- [ ] Real-time job progress tracking (WebSockets)
- [ ] Export results (PDF, DOCX)
- [ ] Search functionality
- [ ] User settings and preferences

### Phase 7: Deployment & Documentation (Week 11-12)
- [ ] Production Docker configuration
- [ ] CI/CD pipeline setup
- [ ] API documentation (Swagger/OpenAPI)
- [ ] User documentation
- [ ] Deployment guides
- [ ] Monitoring and logging setup

---

## Technology Details

### Go Backend Dependencies
```go
// main dependencies
github.com/gin-gonic/gin          // Web framework
github.com/jackc/pgx/v5           // PostgreSQL driver
github.com/golang-jwt/jwt/v5      // JWT authentication
github.com/joho/godotenv          // Environment variables
github.com/google/uuid            // UUID generation
go.uber.org/zap                   // Logging
github.com/go-playground/validator/v10  // Input validation
github.com/swaggo/gin-swagger     // Swagger documentation
github.com/swaggo/swag            // Swagger generator
golang.org/x/crypto/bcrypt        // Password hashing
github.com/ulule/limiter/v3       // Rate limiting
github.com/robfig/cron/v3         // Scheduled tasks
```

### Vue.js Frontend Dependencies
```json
{
  "dependencies": {
    "vue": "^3.4.0",
    "vue-router": "^4.2.0",
    "pinia": "^2.1.0",
    "axios": "^1.6.0",
    "element-plus": "^2.5.0",
    "marked": "^11.0.0",
    "highlight.js": "^11.9.0",
    "@vueuse/core": "^10.7.0",
    "socket.io-client": "^4.6.0",
    "file-saver": "^2.0.5",
    "chart.js": "^4.4.0",
    "vue-chartjs": "^5.3.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0.0",
    "vite": "^5.0.0",
    "eslint": "^8.56.0",
    "prettier": "^3.1.0",
    "sass": "^1.69.0"
  }
}
```

### OCR Service Dependencies
```txt
# requirements.txt
vllm>=0.8.5
torch>=2.6.0
torchvision>=0.21.0
transformers>=4.51.1
fastapi>=0.109.0
uvicorn[standard]>=0.27.0
pillow>=10.2.0
python-multipart>=0.0.9
pydantic>=2.5.0
pydantic-settings>=2.1.0
python-dotenv>=1.0.0
aiofiles>=23.2.1
redis>=5.0.0
celery>=5.3.4
pymupdf>=1.23.8  # for PDF processing
numpy>=1.26.0
opencv-python>=4.9.0
```

---

## Docker Compose Configuration

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: ocr_postgres
    environment:
      POSTGRES_DB: ocr_db
      POSTGRES_USER: ocr_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./database/migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ocr_user -d ocr_db"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - ocr_network

  redis:
    image: redis:7-alpine
    container_name: ocr_redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - ocr_network

  backend:
    build: 
      context: ./backend
      dockerfile: Dockerfile
    container_name: ocr_backend
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_NAME: ocr_db
      DB_USER: ocr_user
      DB_PASSWORD: ${DB_PASSWORD}
      OCR_SERVICE_URL: http://ocr-service:8000
      REDIS_URL: redis://redis:6379
      JWT_SECRET: ${JWT_SECRET}
      PORT: 8080
    volumes:
      - ./storage:/app/storage
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      ocr-service:
        condition: service_started
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
    networks:
      - ocr_network

  ocr-service:
    build: 
      context: ./ocr-service
      dockerfile: Dockerfile
    container_name: ocr_service
    environment:
      MODEL_PATH: deepseek-ai/DeepSeek-OCR
      CUDA_VISIBLE_DEVICES: "0"
      REDIS_URL: redis://redis:6379
      MAX_WORKERS: 4
      BASE_SIZE: 1024
      IMAGE_SIZE: 640
      CROP_MODE: "true"
    volumes:
      - ./storage:/app/storage
      - model_cache:/root/.cache/huggingface
    ports:
      - "8000:8000"
    depends_on:
      redis:
        condition: service_healthy
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
    networks:
      - ocr_network

  frontend:
    build: 
      context: ./frontend
      dockerfile: Dockerfile
    container_name: ocr_frontend
    environment:
      VITE_API_URL: http://localhost:8080/api/v1
      VITE_WS_URL: ws://localhost:8080/ws
    ports:
      - "3000:3000"
    depends_on:
      - backend
    restart: unless-stopped
    networks:
      - ocr_network

  # Optional: Nginx reverse proxy for production
  nginx:
    image: nginx:alpine
    container_name: ocr_nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
    depends_on:
      - frontend
      - backend
    restart: unless-stopped
    networks:
      - ocr_network
    profiles:
      - production

volumes:
  postgres_data:
  redis_data:
  model_cache:

networks:
  ocr_network:
    driver: bridge
```

---

## Security Considerations

1. **Authentication & Authorization**:
   - JWT-based authentication with refresh tokens
   - Token rotation and blacklisting
   - Secure password hashing (bcrypt with cost factor 12+)
   - Multi-factor authentication (future enhancement)
   - API key authentication for programmatic access

2. **File Upload Security**:
   - Validate file types (whitelist approach)
   - File size limits (configurable per user tier)
   - Sanitize filenames (remove special characters)
   - Virus/malware scanning (ClamAV integration)
   - Generate unique file hashes for deduplication

3. **Database Security**:
   - Use parameterized queries (pgx handles this)
   - Principle of least privilege for DB users
   - Encrypt sensitive fields (passwords, API keys)
   - Regular backups with encryption
   - Connection pooling with limits

4. **Network Security**:
   - CORS configuration (whitelist trusted origins)
   - HTTPS/TLS for production
   - Rate limiting per IP/user (prevent abuse)
   - DDoS protection with reverse proxy
   - Internal service communication over private network

5. **Data Privacy**:
   - GDPR/privacy compliance
   - Data retention policies
   - Soft delete for user data recovery
   - Secure file storage with access controls
   - Audit logs for sensitive operations

6. **Input Validation**:
   - Validate all user inputs on backend
   - Sanitize output to prevent XSS
   - Content Security Policy (CSP) headers
   - Protection against CSRF attacks

7. **Dependency Security**:
   - Regular dependency updates
   - Vulnerability scanning (Snyk, Dependabot)
   - Use official Docker images
   - Pin dependency versions

8. **Secrets Management**:
   - Use environment variables
   - Never commit secrets to repo
   - Use secret managers (Vault, AWS Secrets Manager)
   - Rotate secrets regularly

---

## Performance Optimization

1. **OCR Service**:
   - Use vLLM for high-throughput inference
   - Implement job queue with Redis + Celery
   - Batch processing for multiple documents
   - Model caching and warm-up
   - GPU utilization monitoring
   - Dynamic resolution adjustment based on document type
   - Cache frequently processed documents (with hash comparison)

2. **Backend**:
   - Database connection pooling (pgxpool)
   - Query optimization with proper indexes
   - Caching with Redis (user sessions, frequent queries)
   - Async processing for long-running tasks
   - Pagination for list endpoints
   - Compression for API responses (gzip)
   - Database read replicas for scaling reads

3. **Frontend**:
   - Lazy loading components and routes
   - Image optimization (WebP format, thumbnails)
   - Progressive result loading (stream large results)
   - Service worker for offline capabilities
   - Asset bundling and minification
   - CDN for static assets
   - Virtual scrolling for large lists

4. **Infrastructure**:
   - Horizontal scaling with load balancer
   - Database partitioning for large datasets
   - File storage optimization (tiered storage)
   - Content delivery network (CDN)
   - Resource monitoring and auto-scaling

---

## Monitoring & Logging

1. **Logging**:
   - Structured logging with Zap (Go) and Python logging
   - Request/response logging with correlation IDs
   - OCR processing logs with performance metrics
   - Error tracking and stack traces
   - Log aggregation (ELK stack or Loki)
   - Log rotation and retention policies

2. **Metrics**:
   - API response times (p50, p95, p99)
   - OCR processing duration per resolution mode
   - Success/failure rates
   - Queue depth and wait times
   - Database query performance
   - Resource utilization (CPU, memory, GPU)
   - User activity metrics

3. **Health Checks**:
   - Database connectivity and replication lag
   - OCR service availability and GPU status
   - Redis connectivity
   - Disk space monitoring
   - Model loading status
   - External service dependencies

4. **Alerting**:
   - High error rates
   - Service downtime
   - Resource exhaustion
   - Long queue backlogs
   - Security incidents
   - PagerDuty/Slack integration

5. **Observability Tools**:
   - Prometheus for metrics collection
   - Grafana for visualization
   - Jaeger for distributed tracing
   - Application Performance Monitoring (APM)

---

## Future Enhancements

### Short-term (1-3 months)
1. **Cloud Storage**: S3/MinIO integration for scalable storage
2. **Real-time Updates**: WebSocket for live job progress
3. **Batch Processing**: Upload and process multiple files at once
4. **Export Formats**: Export results to PDF, DOCX, TXT
5. **User Quotas**: Implement usage limits and tier system

### Medium-term (3-6 months)
1. **Multi-language Support**: i18n for UI (English, Chinese, Spanish, etc.)
2. **Advanced OCR Features**:
   - Table extraction with structure preservation
   - Mathematical formula recognition (LaTeX output)
   - Layout analysis and section detection
   - Handwriting recognition improvements
3. **Search & Filter**: Full-text search across processed documents
4. **Document Comparison**: Compare OCR results side-by-side
5. **Template System**: Save and reuse OCR configurations

### Long-term (6-12 months)
1. **Collaboration Features**: 
   - Share documents/results with team members
   - Role-based access control (RBAC)
   - Comment and annotation system
2. **Mobile App**: React Native/Flutter app for on-the-go OCR
3. **API Ecosystem**:
   - Public API with documentation
   - SDKs for popular languages (Python, JavaScript, Go)
   - Webhooks for event notifications
4. **Analytics Dashboard**: 
   - Usage statistics and trends
   - Cost analysis
   - Performance insights
5. **AI Enhancements**:
   - Auto-categorization of documents
   - Intelligent error correction
   - Custom model fine-tuning
   - Multi-modal understanding (text + images)
6. **Enterprise Features**:
   - SSO integration (SAML, OAuth)
   - Audit logs and compliance reports
   - On-premise deployment option
   - SLA guarantees

---

## Testing Strategy

### Unit Testing
- **Backend (Go)**:
  - Test all service functions with table-driven tests
  - Mock database and external service calls
  - Achieve 80%+ code coverage
  - Tools: `testing` package, `testify`, `gomock`

- **Frontend (Vue.js)**:
  - Component unit tests with Vitest
  - Test user interactions and state changes
  - Tools: Vitest, Vue Test Utils

- **OCR Service (Python)**:
  - Test API endpoints and inference logic
  - Mock model calls for faster tests
  - Tools: pytest, pytest-asyncio

### Integration Testing
- API endpoint testing with real database
- End-to-end OCR workflow testing
- Database migration testing
- File upload and storage testing
- Tools: Postman/Thunder Client, Go integration tests

### Performance Testing
- Load testing with k6 or Locust
- OCR processing benchmarks
- Database query performance
- Concurrent user simulation
- API rate limit validation

### Security Testing
- OWASP Top 10 vulnerability scanning
- Penetration testing
- Dependency vulnerability scanning
- SQL injection and XSS testing
- Tools: OWASP ZAP, Snyk

### E2E Testing
- User journey automation
- Cross-browser testing
- Mobile responsiveness testing
- Tools: Playwright, Cypress

---

## Getting Started

### Prerequisites
- **Docker & Docker Compose**: v20.10+ and v2.0+
- **NVIDIA GPU**: CUDA-capable GPU (for OCR service)
- **NVIDIA Docker Runtime**: nvidia-docker2 installed
- **Hardware**:
  - 16GB+ RAM recommended (32GB for production)
  - 50GB+ disk space (models + data)
  - GPU: 8GB+ VRAM (16GB recommended for large models)
- **Software**:
  - Git
  - Make (optional, for convenience)
  - Node.js 18+ (for local frontend development)
  - Go 1.21+ (for local backend development)
  - Python 3.11+ (for local OCR service development)

### Environment Setup

1. **Clone repository**:
```bash
git clone https://github.com/yourusername/visekai.git
cd visekai
```

2. **Copy and configure environment file**:
```bash
cp .env.example .env
```

Edit `.env` with your configuration:
```bash
# Database
DB_PASSWORD=your_secure_password
POSTGRES_USER=ocr_user
POSTGRES_DB=ocr_db

# JWT
JWT_SECRET=your_jwt_secret_key_min_32_chars
JWT_EXPIRY=24h
REFRESH_TOKEN_EXPIRY=168h

# Redis
REDIS_PASSWORD=your_redis_password

# OCR Service
MODEL_PATH=deepseek-ai/DeepSeek-OCR
CUDA_VISIBLE_DEVICES=0
MAX_WORKERS=4

# Storage
STORAGE_PATH=/app/storage
MAX_FILE_SIZE=50MB

# Frontend
VITE_API_URL=http://localhost:8080/api/v1
```

3. **Initialize storage directories**:
```bash
mkdir -p storage/{uploads,results,temp,thumbnails}
chmod -R 755 storage
```

### Quick Start (Development)

Start all services:
```bash
docker-compose up -d
```

Check service status:
```bash
docker-compose ps
```

View logs:
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f ocr-service
```

### Accessing the Application

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080/api/v1
- **API Documentation**: http://localhost:8080/swagger/index.html
- **OCR Service Health**: http://localhost:8000/health

### Initial Setup

1. **Run database migrations**:
```bash
docker-compose exec backend ./migrate up
```

2. **Create admin user** (optional):
```bash
docker-compose exec backend ./create-admin
```

3. **Test OCR service**:
```bash
curl -X POST http://localhost:8000/health
```

### Development Workflow

#### Backend Development
```bash
cd backend
go mod download
make run  # Run with hot reload
make test # Run tests
```

#### Frontend Development
```bash
cd frontend
npm install
npm run dev   # Development server with hot reload
npm run build # Production build
```

#### OCR Service Development
```bash
cd ocr-service
pip install -r requirements.txt
python main.py  # Run FastAPI server
```

### Common Commands

```bash
# Stop all services
docker-compose down

# Rebuild services after code changes
docker-compose up -d --build

# Reset database (CAUTION: destroys all data)
docker-compose down -v
docker-compose up -d

# View resource usage
docker stats

# Clean up unused resources
docker system prune -a
```

### Troubleshooting

**GPU not detected**:
```bash
# Verify NVIDIA runtime
docker run --rm --gpus all nvidia/cuda:12.0-base nvidia-smi
```

**Database connection issues**:
```bash
# Check PostgreSQL logs
docker-compose logs postgres

# Connect to database
docker-compose exec postgres psql -U ocr_user -d ocr_db
```

**OCR service out of memory**:
- Reduce `MAX_WORKERS` in `.env`
- Use smaller resolution mode (`tiny` or `small`)
- Increase Docker memory limit

**Port conflicts**:
- Check if ports 3000, 8080, 8000, 5432 are available
- Modify port mappings in `docker-compose.yml`

---

## Contributing

Guidelines for contributing to the project (to be expanded).

## License

MIT License (or your preferred license)
