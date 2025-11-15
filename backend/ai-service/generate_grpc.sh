# generate_grpc.sh
#!/bin/bash
# Script to generate gRPC Python code from proto file

echo "Generating gRPC Python code..."

python -m grpc_tools.protoc \
    -I. \
    --python_out=./generated \
    --grpc_python_out=./generated \
    ai_service.proto

echo "Done! Generated files are in ./generated/"
echo ""
echo "Next steps:"
echo "1. Update imports in client.py and server.py to use generated code"
echo "2. Run the server: python main.py"
echo "3. Test with client: python client.py"