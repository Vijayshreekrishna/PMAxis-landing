@echo off
setlocal enabledelayedexpansion

REM Capture short git commit hash for rollback-friendly image tags
for /f %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_HASH=%%i
if "!GIT_HASH!"=="" set GIT_HASH=local

echo [1/4] Building PMAxis Core Images (git: !GIT_HASH!)...

REM 1. Discovery
docker build -t pmaxis-discovery:latest -t pmaxis-discovery:!GIT_HASH! -f services/discovery/Dockerfile .

REM 2. Ingestion
docker build -t pmaxis-ingestion:latest -t pmaxis-ingestion:!GIT_HASH! -f services/ingestion/Dockerfile .

REM 3. Processor
docker build -t pmaxis-processor:latest -t pmaxis-processor:!GIT_HASH! -f services/processor/Dockerfile .

REM 4. Storage
docker build -t pmaxis-storage:latest -t pmaxis-storage:!GIT_HASH! -f services/storage/Dockerfile .

REM 5. API
docker build -t pmaxis-api:latest -t pmaxis-api:!GIT_HASH! -f services/api/Dockerfile .

echo [2/4] Exporting Images to .tar files...
if not exist "deployments\package" mkdir "deployments\package"

docker save pmaxis-discovery:latest -o deployments\package\discovery.tar
docker save pmaxis-ingestion:latest -o deployments\package\ingestion.tar
docker save pmaxis-processor:latest -o deployments\package\processor.tar
docker save pmaxis-storage:latest -o deployments\package\storage.tar
docker save pmaxis-api:latest -o deployments\package\api.tar

echo [3/4] Preparing Deployment Package...
copy deployments\docker-compose.vps.yml deployments\package\docker-compose.yml
copy deployments\.env deployments\package\.env

echo [4/4] Done!
echo.
echo === SHIPMENT INSTRUCTIONS ===
echo Build tag: !GIT_HASH!
echo.
echo 1. Transfer the 'deployments\package' folder to your VPS.
echo 2. On the VPS, run:
echo    ls *.tar ^| xargs -I {} docker load -i {}
echo 3. Start the stack:
echo    docker-compose up -d
echo.
echo To roll back to this build later:
echo    docker-compose down
echo    REM Edit docker-compose.yml image tags to :!GIT_HASH!
echo    docker-compose up -d
echo =============================
pause
