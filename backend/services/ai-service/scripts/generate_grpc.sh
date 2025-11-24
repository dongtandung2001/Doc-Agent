# generate_grpc.sh
#!/bin/bash
# Script to generate gRPC Python code from proto file

echo "Generating gRPC Python code..."

python -m grpc_tools.protoc \
    -I./protos \
    --python_out=./generated \
    --grpc_python_out=./generated \
    ./protos/ai_service.proto

# Fix imports in generated file to work with package structure
if [ -f "./generated/ai_service_pb2_grpc.py" ]; then
    python3 << 'PYTHON_SCRIPT'
import re

file_path = "./generated/ai_service_pb2_grpc.py"
with open(file_path, 'r') as f:
    content = f.read()

# Replace the import statement
old_import = "import ai_service_pb2 as ai__service__pb2"
new_import = """try:
    import generated.ai_service_pb2 as ai__service__pb2
except ImportError:
    import ai_service_pb2 as ai__service__pb2"""

if old_import in content:
    content = content.replace(old_import, new_import)
    with open(file_path, 'w') as f:
        f.write(content)
    print("✓ Fixed imports in generated file")
else:
    print("⚠ Import statement not found (may already be fixed)")
PYTHON_SCRIPT
fi

echo "Done! Generated files are in ./generated/"
echo ""
echo "Next steps:"
echo "1. Run the server: python main.py"
echo "2. Test with client: python tests/test_client.py"