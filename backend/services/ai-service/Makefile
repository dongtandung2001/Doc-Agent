
# Makefile
.PHONY: help install proto run test clean docker-build docker-run

help:
	@echo "AI Service Microservice - Commands"
	@echo "==================================="
	@echo "install      - Install Python dependencies"
	@echo "proto        - Generate gRPC code from proto file"
	@echo "run          - Run the AI service"
	@echo "test         - Run tests"
	@echo "client       - Run test client"
	@echo "interactive  - Run interactive client"
	@echo "clean        - Clean generated files"
	@echo "docker-build - Build Docker image"
	@echo "docker-run   - Run service in Docker"

install:
	pip install -r requirements.txt

proto:
	mkdir -p generated
	python -m grpc_tools.protoc -I. --python_out=./generated --grpc_python_out=./generated ai_service.proto
	@echo "✓ gRPC code generated in ./generated/"

run:
	python main.py

test:
	pytest tests/ -v --cov=. --cov-report=html

client:
	python client.py

interactive:
	python client.py interactive

clean:
	rm -rf generated/
	rm -rf __pycache__/
	rm -rf .pytest_cache/
	rm -rf htmlcov/
	rm -rf .coverage
	find . -type d -name "__pycache__" -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete

docker-build:
	docker build -t ai-service:latest .

docker-run:
	docker-compose up -d

docker-logs:
	docker-compose logs -f ai-service

docker-stop:
	docker-compose down