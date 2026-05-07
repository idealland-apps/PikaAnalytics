#!/bin/bash

echo "🚀 Building PikaAnalytics for production..."

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Default version
VERSION="dev"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        *)
            echo -e "${RED}❌ Unknown option: $1${NC}"
            echo "Usage: $0 [-v VERSION]"
            exit 1
            ;;
    esac
done

# Check if we're in the right directory
if [ ! -f "README.md" ]; then
    echo -e "${RED}❌ Please run this script from the project root directory${NC}"
    exit 1
fi

# Create dist directory
echo -e "${BLUE}📁 Creating distribution directory...${NC}"
rm -rf dist
mkdir -p dist

# Build frontend
echo -e "${BLUE}🔨 Building frontend...${NC}"
cd frontend
npm install
npm run build
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Frontend build failed${NC}"
    exit 1
fi

# Copy frontend build to backend
echo -e "${BLUE}📦 Preparing backend with frontend assets...${NC}"
cd ../backend
rm -rf frontend
mkdir -p frontend
cp -r ../frontend/build/* frontend/

# Build Go binary for current platform
echo -e "${BLUE}🔨 Building Go binary...${NC}"
go mod tidy
go build -o ../dist/pikaanalytics main.go
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Go build failed${NC}"
    exit 1
fi

# Copy necessary files to dist
cd ..
cp -r backend/frontend dist/
cp backend/pikaanalytics.db dist/ 2>/dev/null || echo "No existing database found, will be created on first run"

# Create version file
echo -e "${BLUE}📝 Creating version file...${NC}"
echo -n "$VERSION" > dist/version.txt

echo -e "${GREEN}✅ Build completed successfully!${NC}"
echo -e "${GREEN}📦 Production files are in the 'dist' directory${NC}"
echo -e "${GREEN}🚀 To run: cd dist && ./pikaanalytics${NC}"
echo -e "${GREEN}📌 Version: $VERSION${NC}"