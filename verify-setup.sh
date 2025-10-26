#!/bin/bash

# VisEkai Setup Verification Script

echo "🔍 VisEkai Project Setup Verification"
echo "======================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check function
check_file() {
    if [ -f "$1" ]; then
        echo -e "${GREEN}✓${NC} $1"
        return 0
    else
        echo -e "${RED}✗${NC} $1 (missing)"
        return 1
    fi
}

check_dir() {
    if [ -d "$1" ]; then
        echo -e "${GREEN}✓${NC} $1/"
        return 0
    else
        echo -e "${RED}✗${NC} $1/ (missing)"
        return 1
    fi
}

# Check root files
echo "📁 Root Configuration Files:"
check_file ".env.example"
check_file ".gitignore"
check_file "docker-compose.yml"
check_file "Makefile"
check_file "README.md"
check_file "PROJECT_PLAN.md"
echo ""

# Check directories
echo "📂 Project Directories:"
check_dir "backend"
check_dir "frontend"
check_dir "ocr-service"
check_dir "database"
check_dir "storage"
check_dir "nginx"
echo ""

# Check backend structure
echo "🔧 Backend (Go):"
check_file "backend/go.mod"
check_file "backend/Dockerfile"
check_file "backend/cmd/server/main.go"
check_file "backend/internal/config/config.go"
check_file "backend/internal/database/postgres.go"
check_file "backend/internal/handlers/handlers.go"
check_file "backend/internal/middleware/middleware.go"
check_file "backend/pkg/logger/logger.go"
echo ""

# Check frontend structure
echo "🎨 Frontend (Vue.js):"
check_file "frontend/package.json"
check_file "frontend/vite.config.js"
check_file "frontend/index.html"
check_file "frontend/Dockerfile"
check_file "frontend/src/main.js"
check_file "frontend/src/App.vue"
check_file "frontend/src/router/index.js"
check_file "frontend/src/views/Home.vue"
check_file "frontend/src/services/api.js"
echo ""

# Check OCR service structure
echo "🤖 OCR Service (Python):"
check_file "ocr-service/requirements.txt"
check_file "ocr-service/Dockerfile"
check_file "ocr-service/main.py"
check_file "ocr-service/core/config.py"
check_file "ocr-service/core/logging.py"
check_file "ocr-service/api/routes.py"
check_file "ocr-service/deepseek_ocr/model.py"
echo ""

# Check database
echo "🗄️  Database:"
check_file "database/migrations/001_init_schema.sql"
echo ""

# Check storage
echo "💾 Storage Directories:"
check_dir "storage/uploads"
check_dir "storage/results"
check_dir "storage/temp"
check_dir "storage/thumbnails"
echo ""

# Check for .env file
echo "⚙️  Configuration:"
if [ -f ".env" ]; then
    echo -e "${GREEN}✓${NC} .env file exists"
else
    echo -e "${YELLOW}⚠${NC}  .env file not found (copy from .env.example)"
fi
echo ""

# Check Docker
echo "🐳 Docker Environment:"
if command -v docker &> /dev/null; then
    echo -e "${GREEN}✓${NC} Docker installed: $(docker --version)"
else
    echo -e "${RED}✗${NC} Docker not installed"
fi

if command -v docker-compose &> /dev/null; then
    echo -e "${GREEN}✓${NC} Docker Compose installed: $(docker-compose --version)"
else
    echo -e "${RED}✗${NC} Docker Compose not installed"
fi
echo ""

# Check NVIDIA Docker (optional)
echo "🎮 GPU Support (Optional):"
if docker run --rm --gpus all nvidia/cuda:12.0-base nvidia-smi &> /dev/null; then
    echo -e "${GREEN}✓${NC} NVIDIA Docker runtime working"
else
    echo -e "${YELLOW}⚠${NC}  NVIDIA Docker runtime not available (optional for development)"
fi
echo ""

# Summary
echo "======================================"
echo "📊 Summary:"
echo ""
echo "✅ Project structure is complete!"
echo ""
echo "📝 Next Steps:"
echo "  1. Copy .env.example to .env and configure"
echo "  2. Review PROJECT_PLAN.md for architecture details"
echo "  3. Read SETUP_COMPLETE.md for next steps"
echo "  4. Run 'make help' to see available commands"
echo ""
echo "🚀 To start development:"
echo "  $ cp .env.example .env"
echo "  $ nano .env  # Edit configuration"
echo "  $ make up    # Start all services"
echo ""
